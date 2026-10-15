package module

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofault/gofault/core"
)

// The topological sort walked a map, so Go's randomised iteration decided the
// start order: 40 runs over the same graph produced five different orders.
func TestSortModules_IsDeterministic(t *testing.T) {
	names := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf"}

	var first string
	for run := 0; run < 60; run++ {
		app := New()
		for _, n := range names {
			if err := app.RegisterModules(core.NewModule(n)); err != nil {
				t.Fatalf("run %d: %v", run, err)
			}
		}
		if err := app.Init(); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}

		var b strings.Builder
		for _, m := range app.modules {
			b.WriteString(m.Name)
			b.WriteByte(',')
		}
		got := b.String()

		if run == 0 {
			first = got
			continue
		}
		if got != first {
			t.Fatalf("run %d produced order %s, run 0 produced %s", run, got, first)
		}
	}
}

// Dependencies must still win over registration order.
func TestSortModules_RespectsDependencies(t *testing.T) {
	app := New()

	// Registered in reverse dependency order on purpose.
	app.RegisterModules(core.NewModule("last"))
	app.RegisterModules(core.NewModule("middle"))
	app.RegisterModules(core.NewModule("first"))

	app.moduleMap["middle"].Depends = []string{"first"}
	app.moduleMap["last"].Depends = []string{"middle"}

	if err := app.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var order []string
	for _, m := range app.modules {
		order = append(order, m.Name)
	}
	got := strings.Join(order, ",")
	if got != "first,middle,last" {
		t.Errorf("order = %s, want first,middle,last", got)
	}
}

// A cycle must be reported in a stable order so the message is reproducible.
func TestSortModules_CycleMessageIsStable(t *testing.T) {
	var first string
	for run := 0; run < 30; run++ {
		app := New()
		for _, n := range []string{"a", "b", "c", "d"} {
			app.RegisterModules(core.NewModule(n))
		}
		for _, m := range app.moduleMap {
			m.Depends = []string{"next"}
		}
		app.moduleMap["a"].Depends = []string{"d"}
		app.moduleMap["b"].Depends = []string{"a"}
		app.moduleMap["c"].Depends = []string{"b"}
		app.moduleMap["d"].Depends = []string{"c"}

		err := app.Init()
		if err == nil {
			t.Fatal("expected a cycle error")
		}
		if run == 0 {
			first = err.Error()
			continue
		}
		if err.Error() != first {
			t.Fatalf("run %d message %q, run 0 message %q", run, err, first)
		}
	}
}

// Sorting must be idempotent: Init is reachable more than once.
func TestSortModules_Idempotent(t *testing.T) {
	app := New()
	app.RegisterModules(core.NewModule("b"))
	app.RegisterModules(core.NewModule("a"))
	app.moduleMap["b"].Depends = []string{"a"}

	if err := app.Init(); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	first := app.modules[0].Name

	if err := app.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if app.modules[0].Name != first {
		t.Errorf("second Init reordered modules: %s then %s", first, app.modules[0].Name)
	}
}

// All modules must still be present after sorting, none dropped.
func TestSortModules_LosesNoModules(t *testing.T) {
	app := New()
	names := []string{"a", "b", "c", "d", "e", "f"}
	for _, n := range names {
		app.RegisterModules(core.NewModule(n))
	}
	app.moduleMap["f"].Depends = []string{"a", "b", "c", "d", "e"}

	if err := app.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(app.modules) != len(names) {
		t.Fatalf("got %d modules, want %d", len(app.modules), len(names))
	}
	if app.modules[len(app.modules)-1].Name != "f" {
		t.Errorf("the module with dependencies on everything should sort last, got %s",
			app.modules[len(app.modules)-1].Name)
	}
}

// Stop is conventionally called from a signal handler, which can arrive while
// Start is still bootstrapping. Stop reads a.modules to run the shutdown hooks
// and Bootstrap writes it while sorting, so the two must not overlap. Confirmed
// as a data race on module.go:216 before the mutex was extended to cover the
// slice.
func TestApp_StopRacingBootstrap(t *testing.T) {
	for round := 0; round < 200; round++ {
		app := New()
		mod := core.NewModule("m")
		mod.RegisterOnShutdown(&dummyHook{})
		app.RegisterModules(mod)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = app.Bootstrap() }()
		go func() { defer wg.Done(); _ = app.Stop() }()
		wg.Wait()
	}
}

// Restructuring bootstrap must not lose the OnInit hooks: they are what a module
// uses to open its resources.
func TestApp_InitHooksStillRun(t *testing.T) {
	app := New()
	mod := core.NewModule("m")
	hook := &recordingHook{name: "init", order: new([]string)}
	mod.RegisterOnInit(hook)
	app.RegisterModules(mod)

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if len(*hook.order) != 1 || (*hook.order)[0] != "init" {
		t.Errorf("OnInit hooks ran %v, want [init]", *hook.order)
	}
}

// Routes must still be registered, since bootstrap was restructured to run under
// the lock.
func TestApp_RoutesStillRegistered(t *testing.T) {
	app := New()
	mod := core.NewModule("m")
	mod.RegisterControllers(&dummyController{})
	app.RegisterModules(mod)

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if !app.booted {
		t.Error("app not marked booted")
	}

	rtr := app.Router()
	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; routes were not registered", w.Code)
	}
}
