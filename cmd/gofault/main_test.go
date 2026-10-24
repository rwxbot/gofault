package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	valid := []string{"myapp", "my-app", "app_1", "MyApp"}
	for _, name := range valid {
		if err := validateProjectName(name); err != nil {
			t.Errorf("validateProjectName(%q) = %v, want nil", name, err)
		}
	}

	// A name that is not a single path segment could create a directory
	// outside the working directory or yield an unresolvable module path.
	invalid := []string{"", ".", "..", "../evil", "a/b", "/abs", `a\b`, "nul\x00", "ctrl\nchar"}
	for _, name := range invalid {
		if err := validateProjectName(name); err == nil {
			t.Errorf("validateProjectName(%q) = nil, want an error", name)
		}
	}
}

func TestExecuteTemplate(t *testing.T) {
	out, err := executeTemplate("hello {{.Name}}", map[string]string{"Name": "world"})
	if err != nil {
		t.Fatalf("executeTemplate: %v", err)
	}
	if out != "hello world" {
		t.Errorf("output = %q, want %q", out, "hello world")
	}

	if _, err := executeTemplate("{{.Unclosed", map[string]string{}); err == nil {
		t.Error("expected a parse error for a malformed template")
	}
}

// A traversal name must be refused, and must not create anything on disk.
func TestCreateProject_RejectsTraversal(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	for _, name := range []string{"../escaped", "sub/dir", ".."} {
		if err := createProject(name); err == nil {
			t.Errorf("createProject(%q) = nil, want an error", name)
		}
	}

	entries, err := os.ReadDir(wd)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected the working directory to stay empty, got %d entries", len(entries))
	}
}

func TestCreateProject_WritesScaffold(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	if err := createProject("myapp"); err != nil {
		t.Fatalf("createProject: %v", err)
	}

	for _, rel := range []string{"main.go", "go.mod", "config.yaml", "controllers/hello.go"} {
		if _, err := os.Stat(filepath.Join(wd, "myapp", rel)); err != nil {
			t.Errorf("expected %s to be scaffolded: %v", rel, err)
		}
	}

	goMod, err := os.ReadFile(filepath.Join(wd, "myapp", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	// The module path must match the directory name, and the framework
	// requirement must not claim a release that does not exist.
	if !strings.Contains(string(goMod), "module myapp") {
		t.Errorf("go.mod does not declare the module path:\n%s", goMod)
	}
	if !strings.Contains(string(goMod), "github.com/gofault/gofault v0.0.0") {
		t.Errorf("go.mod pins a released version instead of the local replace:\n%s", goMod)
	}

	mainGo, err := os.ReadFile(filepath.Join(wd, "myapp", "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	// The module path is {{.Name}}, so the controllers import must be
	// "<name>/controllers" -- not a path under github.com/gofault/gofault.
	if !strings.Contains(string(mainGo), `"myapp/controllers"`) {
		t.Errorf("main.go imports the wrong controllers path:\n%s", mainGo)
	}
	if strings.Contains(string(mainGo), "github.com/gofault/gofault/myapp") {
		t.Errorf("main.go treats the project as a subpackage of the framework:\n%s", mainGo)
	}
}

// A scaffolded project that does not compile is worse than no project at all,
// so build the generated code for real.
func TestCreateProject_OutputCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compile check in short mode")
	}

	// Resolve the framework root before chdir, since the relative path is
	// only meaningful from the package directory.
	absGofault, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve framework root: %v", err)
	}

	wd := t.TempDir()
	t.Chdir(wd)

	if err := createProject("buildme"); err != nil {
		t.Fatalf("createProject: %v", err)
	}

	// Point the generated module at this checkout, standing in for the
	// ../gofault relative replace that only resolves inside the repo.

	cmd := exec.Command("go", "mod", "edit",
		"-replace", "github.com/gofault/gofault="+absGofault)
	cmd.Dir = filepath.Join(wd, "buildme")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod edit: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = filepath.Join(wd, "buildme")
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("scaffolded project failed to build: %v\n%s", err, out)
	}
}

func TestRunNew(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"gofault"}
	if err := runNew(); err == nil {
		t.Error("expected an error when the project name is missing")
	}

	os.Args = []string{"gofault", "new", "app1"}
	if err := runNew(); err != nil {
		t.Fatalf("runNew: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wd, "app1", "main.go")); err != nil {
		t.Errorf("expected app1 to be scaffolded: %v", err)
	}
}

// printUsage must document every subcommand the switch in main handles,
// otherwise the help output drifts away from the actual CLI surface.
func TestPrintUsage(t *testing.T) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	printUsage()
	// Close the write end so the read below terminates.
	_ = w.Close()
	os.Stdout = orig

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read usage: %v", err)
	}
	usage := string(out)
	for _, want := range []string{"new", "version", "help", "gofault new myapp"} {
		if !strings.Contains(usage, want) {
			t.Errorf("usage output missing %q:\n%s", want, usage)
		}
	}
}

// An execution error (not just a parse error) must surface from
// executeTemplate: accessing a field on a string value parses fine and
// only fails when the template runs.
func TestExecuteTemplate_ExecuteError(t *testing.T) {
	if _, err := executeTemplate("{{.Name.Missing}}", map[string]string{"Name": "x"}); err == nil {
		t.Error("expected an execution error for a field lookup on a string value")
	}
}

// The very first MkdirAll (the project directory itself) must fail when
// the name is already taken by a regular file, and the error must name
// the project directory failure.
func TestCreateProject_ProjectDirExistsAsFile(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	if err := os.WriteFile(filepath.Join(wd, "taken"), []byte("occupied"), 0644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}

	err := createProject("taken")
	if err == nil {
		t.Fatal("expected an error when the project name is a regular file")
	}
	if !strings.Contains(err.Error(), "failed to create project directory") {
		t.Errorf("error = %v, want it to name the project directory failure", err)
	}
}

// MkdirAll failing partway through the file loop must abort with a
// wrapped error naming the directory. "controllers/hello.go" is written
// last, so pre-creating controllers as a regular file makes its MkdirAll
// fail without touching the earlier files.
func TestCreateProject_MkdirFails(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	if err := os.Mkdir(filepath.Join(wd, "blocked"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wd, "blocked", "controllers"), []byte("not a directory"), 0644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}

	err := createProject("blocked")
	if err == nil {
		t.Fatal("expected an error when a required directory cannot be created")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("error = %v, want it to name the directory failure", err)
	}
}

// A write failure (here: go.mod replaced by a directory) must abort with
// a wrapped error naming the file.
func TestCreateProject_WriteFails(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	if err := os.Mkdir(filepath.Join(wd, "nowrite"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(wd, "nowrite", "go.mod"), 0755); err != nil {
		t.Fatalf("mkdir go.mod blocker: %v", err)
	}

	err := createProject("nowrite")
	if err == nil {
		t.Fatal("expected an error when a scaffold file cannot be written")
	}
	if !strings.Contains(err.Error(), "failed to write") {
		t.Errorf("error = %v, want it to name the write failure", err)
	}
}
