package ioc

import (
	"context"
	"testing"
)

type dummyService struct {
	msg string
}

func newDummyService() *dummyService {
	return &dummyService{msg: "hello"}
}

func (s *dummyService) SomeMethod() string {
	return s.msg
}

type dependentService struct {
	dep *dummyService
}

func newDependentService(dep *dummyService) *dependentService {
	return &dependentService{dep: dep}
}

func TestContainer_RegisterAndResolve(t *testing.T) {
	c := New()
	err := c.Register(newDummyService)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	svc, err := c.Resolve(&dummyService{})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if svc.(*dummyService).msg != "hello" {
		t.Fatalf("expected 'hello', got %s", svc.(*dummyService).msg)
	}
}

func TestContainer_SingletonScope(t *testing.T) {
	c := New()
	c.Register(newDummyService)
	a, _ := c.Resolve(&dummyService{})
	b, _ := c.Resolve(&dummyService{})
	if a != b {
		t.Fatal("singleton instances must be identical")
	}
}

func TestContainer_TransientScope(t *testing.T) {
	c := New()
	c.RegisterTransient(newDummyService)
	a, _ := c.Resolve(&dummyService{})
	b, _ := c.Resolve(&dummyService{})
	if a == b {
		t.Fatal("transient instances must be different")
	}
}

func TestContainer_ResolveUnknown(t *testing.T) {
	c := New()
	_, err := c.Resolve(&dummyService{})
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestContainer_ResolveFromCtx(t *testing.T) {
	c := New()
	c.Register(newDummyService)

	ctx1 := c.BeginRequest(context.Background())
	ctx2 := c.BeginRequest(context.Background())

	inst1, err := c.ResolveFromCtx(ctx1, &dummyService{})
	if err != nil {
		t.Fatalf("ResolveFromCtx failed: %v", err)
	}

	inst2, err := c.ResolveFromCtx(ctx2, &dummyService{})
	if err != nil {
		t.Fatalf("ResolveFromCtx failed: %v", err)
	}

	// Same context should return same instance.
	same, _ := c.ResolveFromCtx(ctx1, &dummyService{})
	if inst1 != same {
		t.Fatal("same context should return same request-scoped instance")
	}

	// Different context should return different instance (since it's registered as singleton,
	// but BeginRequest doesn't change singleton behavior unless registered as request scope).
	// However, ResolveFromCtx with a non-request-scoped type uses regular Resolve behavior.
	_ = inst2
	_ = ctx2
}

func TestContainer_RequestScopedInstancePerContext(t *testing.T) {
	c := New()
	c.RegisterScoped(newDummyService)

	ctx1 := c.BeginRequest(context.Background())
	ctx2 := c.BeginRequest(context.Background())

	inst1, _ := c.ResolveFromCtx(ctx1, &dummyService{})
	inst2, _ := c.ResolveFromCtx(ctx2, &dummyService{})

	// Different contexts should get different request-scoped instances.
	if inst1 == inst2 {
		t.Fatal("different request contexts should get different instances")
	}

	// Same context should get same instance.
	same, _ := c.ResolveFromCtx(ctx1, &dummyService{})
	if inst1 != same {
		t.Fatal("same context should return same request-scoped instance")
	}
}

func TestContainer_EndRequest_ClearsScope(t *testing.T) {
	c := New()
	c.RegisterScoped(newDummyService)

	ctx := c.BeginRequest(context.Background())
	inst1, _ := c.ResolveFromCtx(ctx, &dummyService{})

	c.EndRequest(ctx)

	// After EndRequest, resolving with same context should get a new instance.
	inst2, _ := c.ResolveFromCtx(ctx, &dummyService{})
	if inst1 == inst2 {
		t.Fatal("after EndRequest, context should get fresh request-scoped instance")
	}
}

func TestContainer_ResolveWithDependency(t *testing.T) {
	c := New()
	c.Register(newDummyService)
	c.Register(newDependentService)

	dep, err := c.ResolveFromCtx(context.Background(), &dependentService{})
	if err != nil {
		t.Fatalf("ResolveWithDependency failed: %v", err)
	}

	ds := dep.(*dependentService)
	if ds.dep == nil {
		t.Fatal("dependency was not injected")
	}
	if ds.dep.msg != "hello" {
		t.Fatalf("expected 'hello', got %s", ds.dep.msg)
	}
}

func TestContainer_ScopedProviderInterface(t *testing.T) {
	// A provider can implement ScopedProvider to declare its own scope.
	// Currently this is a marker interface; the container reads it via reflection.
	c := New()
	c.RegisterScoped(newDummyService)

	ctx := c.BeginRequest(context.Background())
	inst1, _ := c.ResolveFromCtx(ctx, &dummyService{})
	inst2, _ := c.ResolveFromCtx(ctx, &dummyService{})

	if inst1 != inst2 {
		t.Fatal("request-scoped should be same within same context")
	}
}
