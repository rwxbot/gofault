package middleware

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gofault/gofault/core"
)

// A plain strings.HasPrefix containment check would accept a sibling directory
// whose name merely starts with the root's name (/tmp/x-secret matches root
// /tmp/x), exposing files outside the served tree.
func TestStaticMiddleware_RejectsSiblingDirectoryWithSharedPrefix(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "www")
	secret := filepath.Join(parent, "www-secret")

	os.MkdirAll(root, 0755)
	os.MkdirAll(secret, 0755)
	os.WriteFile(filepath.Join(root, "public.txt"), []byte("public"), 0644)
	os.WriteFile(filepath.Join(secret, "private.txt"), []byte("private"), 0644)

	mw := StaticMiddleware(StaticConfig{Dir: root})

	req := httptest.NewRequest("GET", "/../www-secret/private.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	called := false
	if err := mw(ctx, func(ctx *core.Ctx) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body := w.Body.String(); body == "private" {
		t.Fatal("served a file from a sibling directory outside the root")
	}
	if !called {
		t.Error("expected traversal to fall through to the next handler")
	}
}

func TestStaticMiddleware_ServesFilesInsideRoot(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "ok.txt"), []byte("inside"), 0644)

	mw := StaticMiddleware(StaticConfig{Dir: root})

	req := httptest.NewRequest("GET", "/ok.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Body.String() != "inside" {
		t.Errorf("expected file to be served, got %q", w.Body.String())
	}
}

func TestIsWithinRoot(t *testing.T) {
	cases := []struct {
		root, target string
		want         bool
	}{
		{"/srv/www", "/srv/www", true},
		{"/srv/www", "/srv/www/a.txt", true},
		{"/srv/www", "/srv/www/sub/a.txt", true},
		{"/srv/www", "/srv/www-secret/a.txt", false},
		{"/srv/www", "/srv/wwwx", false},
		{"/srv/www", "/srv/other/a.txt", false},
		{"/srv/www", "/etc/passwd", false},
		{"/srv/www/", "/srv/www/a.txt", true},
	}

	for _, c := range cases {
		if got := isWithinRoot(c.root, c.target); got != c.want {
			t.Errorf("isWithinRoot(%q, %q) = %v, want %v", c.root, c.target, got, c.want)
		}
	}
}

func TestSanitizeFileName(t *testing.T) {
	ok := []struct{ in, want string }{
		{"photo.jpg", "photo.jpg"},
		{"../../etc/passwd", "passwd"},
		{"..\\..\\windows\\system32\\cfg", "cfg"},
		{"/etc/shadow", "shadow"},
		{"dir/sub/file.txt", "file.txt"},
		{"  spaced.txt  ", "spaced.txt"},
		{"..", ""}, // rejected, checked below
	}
	for _, c := range ok {
		got, err := sanitizeFileName(c.in)
		if c.want == "" {
			if err == nil {
				t.Errorf("sanitizeFileName(%q) should have been rejected, got %q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("sanitizeFileName(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("sanitizeFileName(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	bad := []string{"", ".", "..", "nul\x00byte", "ctrl\nchar"}
	for _, in := range bad {
		if got, err := sanitizeFileName(in); err == nil {
			t.Errorf("sanitizeFileName(%q) should have been rejected, got %q", in, got)
		}
	}
}

// buildMultipart constructs a single-file multipart body whose filename is
// attacker-controlled, mirroring how a client can set an arbitrary name.
func buildMultipart(t *testing.T, field, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// A traversal filename must not write outside the storage root.
func TestLocalStorage_StoreRejectsTraversalFilename(t *testing.T) {
	base := t.TempDir()
	storage := NewLocalStorage(base)

	body, ct := buildMultipart(t, "file", "../../escaped.txt", "pwned")
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)

	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := req.MultipartForm.File["file"][0]

	// Force the raw traversal name back in, since mime/multipart would have
	// stripped it, to prove the backend sanitizes independently.
	file.Filename = "../../escaped.txt"

	if _, err := storage.Store(file, "uploads"); err != nil {
		t.Fatalf("Store should sanitize rather than fail: %v", err)
	}

	if _, err := os.Stat(filepath.Join(parentOf(base), "escaped.txt")); err == nil {
		t.Fatal("file was written outside the storage root")
	}
	if _, err := os.Stat(filepath.Join(base, "uploads", "escaped.txt")); err != nil {
		t.Fatalf("expected sanitized file inside base/uploads: %v", err)
	}
}

func TestLocalStorage_StoreSanitizesStoredPath(t *testing.T) {
	base := t.TempDir()
	storage := NewLocalStorage(base)

	body, ct := buildMultipart(t, "file", "/etc/shadow", "data")
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := req.MultipartForm.File["file"][0]
	file.Filename = "/etc/shadow"

	stored, err := storage.Store(file, "uploads")
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if stored != "uploads/shadow" {
		t.Errorf("stored path = %q, want %q", stored, "uploads/shadow")
	}
}

func TestLocalStorage_DeleteRejectsEscape(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(parentOf(base), "victim.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0644); err != nil {
		t.Fatalf("seed outside file: %v", err)
	}

	storage := NewLocalStorage(base)
	if err := storage.Delete("../victim.txt"); err == nil {
		t.Error("expected Delete to reject a path escaping the root")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Error("Delete removed a file outside the storage root")
	}
}

// Go's mime/multipart already reduces the filename to its base element, so a
// traversal name cannot reach the middleware over HTTP. Storage backends can
// still be called directly (or by other producers), so the traversal case is
// covered against LocalStorage above. This test pins the end-to-end invariant:
// whatever the client sends, nothing is written outside the root.
func TestUploadMiddleware_neverWritesOutsideRoot(t *testing.T) {
	for _, name := range []string{"../../evil.txt", "/etc/shadow", "..\\..\\win.txt", "ok.txt"} {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			cfg := DefaultUploadConfig()
			cfg.Storage = NewLocalStorage(base)
			mw := UploadMiddleware(cfg)

			body, ct := buildMultipart(t, "file", name, "data")
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", ct)
			w := httptest.NewRecorder()
			ctx := core.NewCtx(w, req)

			if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entries, err := os.ReadDir(parentOf(base))
			if err != nil {
				t.Fatalf("read parent: %v", err)
			}
			if len(entries) != 1 {
				t.Errorf("expected only the storage dir in parent, got %d entries", len(entries))
			}
		})
	}
}

// A traversal name handed straight to the backend must be rejected outright
// rather than silently rewritten, so callers learn the upload was malformed.
func TestLocalStorage_StoreRejectsPureTraversalName(t *testing.T) {
	base := t.TempDir()
	storage := NewLocalStorage(base)

	fh := &multipart.FileHeader{Filename: ".."}
	if _, err := storage.Store(fh, "uploads"); err == nil {
		t.Error("expected Store to reject a bare '..' filename")
	}
}

func parentOf(p string) string {
	return filepath.Dir(p)
}

// A lexical containment check cannot see through a symlink, so <root>/link ->
// /etc passed it while resolving outside the root. FollowSymLinks defaults to
// false, which is supposed to prevent exactly that.
func TestStaticMiddleware_RejectsSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "public")
	outside := filepath.Join(parent, "secret")

	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "key.txt"), []byte("TOP-SECRET"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	mw := StaticMiddleware(StaticConfig{Dir: root})

	req := httptest.NewRequest("GET", "/link/key.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Body.String() == "TOP-SECRET" {
		t.Error("served a file outside the root through a symlink")
	}
}

// Opting in must still serve a symlink that stays inside the root, so the check
// above cannot be satisfied by simply refusing every link.
func TestStaticMiddleware_ServesSymlinkWithinRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "real.txt"), []byte("inside"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "real.txt"), filepath.Join(root, "alias.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	mw := StaticMiddleware(StaticConfig{Dir: root, FollowSymLinks: true})

	req := httptest.NewRequest("GET", "/alias.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Body.String() != "inside" {
		t.Errorf("body = %q, want %q (FollowSymLinks=true)", w.Body.String(), "inside")
	}
}

// A broken symlink must fall through rather than error or hang.
func TestStaticMiddleware_BrokenSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "dangling")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	mw := StaticMiddleware(StaticConfig{Dir: root})

	req := httptest.NewRequest("GET", "/dangling", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	called := false
	if err := mw(ctx, func(ctx *core.Ctx) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected a broken symlink to fall through to the next handler")
	}
}

// Mode decides which list wins when an IP is on both. It was documented but
// never read, so Mode="allow" silently behaved like "block".
func TestIPFilter_Mode(t *testing.T) {
	// 10.0.0.1 is on both lists; 10.0.0.2 is only blocked.
	allow := []string{"10.0.0.1"}
	block := []string{"10.0.0.0/24"}

	serve := func(mode, ip string) bool {
		cfg := IPFilterConfig{Enabled: true, Allow: allow, Block: block, Mode: mode}
		mw := IPFilterMiddleware(cfg)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"
		w := httptest.NewRecorder()
		ctx := core.NewCtx(w, req)

		reached := false
		if err := mw(ctx, func(ctx *core.Ctx) error {
			reached = true
			return nil
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return reached
	}

	// "block" (the default): the block list wins, so the IP on both lists is denied.
	if serve("block", "10.0.0.1") {
		t.Error("Mode=block should deny an IP that is on both lists")
	}
	// "allow": the allow list wins, so the same IP is admitted.
	if !serve("allow", "10.0.0.1") {
		t.Error("Mode=allow should admit an IP that is on both lists")
	}
	// An unrecognised mode must fall back to the safe default rather than
	// silently allowing everything.
	if serve("nonsense", "10.0.0.1") {
		t.Error("an unknown Mode should deny, matching the block default")
	}
	// An IP on neither list is denied in both modes when an allow list exists.
	if serve("allow", "10.0.0.9") {
		t.Error("Mode=allow should still deny an IP that is not on the allow list")
	}
}

// A byte-wise truncation cut mid-rune leaves an invalid trailing byte that
// goes to disk and back in FileInfo.FileName, and a cut across the extension
// boundary turns "x.png" into an extension-less blob. Truncation must land
// on a rune boundary and keep the extension.
func TestSanitizeFileName_TruncationKeepsRuneBoundaryAndExtension(t *testing.T) {
	// 300-byte base of multi-byte runes plus a real extension.
	base := strings.Repeat("é", 150) // 300 bytes
	name := base + ".png"

	got, err := sanitizeFileName(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > maxFileNameLen {
		t.Errorf("len = %d, want <= %d", len(got), maxFileNameLen)
	}
	if !utf8.ValidString(got) {
		t.Errorf("truncated name is not valid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, ".png") {
		t.Errorf("truncated name lost its extension: %q", got)
	}

	// A name ending exactly mid-rune at the cap must not keep a partial rune.
	long := strings.Repeat("a", 254) + "bcdef" // 254 ASCII + 2×3-byte runes
	got2, err := sanitizeFileName(long)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got2) > maxFileNameLen {
		t.Errorf("len = %d, want <= %d", len(got2), maxFileNameLen)
	}
	if !utf8.ValidString(got2) {
		t.Errorf("cap landed mid-rune: %q", got2)
	}
}

// A pre-existing symlink inside the storage root defeats a lexical prefix
// containment check: <root>/uploads -> <outside> passes the check while every
// write resolves through the link. Containment must be enforced against the
// real filesystem (os.Root), not the joined string.
func TestLocalStorage_StoreRejectsSymlinkedDirEscape(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "storage")
	outside := filepath.Join(parent, "elsewhere")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(base, "uploads")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	storage := NewLocalStorage(base)

	body, ct := buildMultipart(t, "file", "evil.txt", "pwned")
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	fh := req.MultipartForm.File["file"][0]

	if _, err := storage.Store(fh, "uploads"); err == nil {
		t.Error("expected Store to fail when the destination dir is a symlink out of the root")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatalf("read outside: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("%d files written outside the root through the symlink", len(entries))
	}
}

// Same exposure on the delete side, via a symlinked *prefix*: with
// <root>/uploads -> <outside>, the lexical check happily accepts
// "uploads/secret.txt" and os.Remove follows the directory link and unlinks
// the victim outside. Under os.Root the kernel refuses to traverse the link.
func TestLocalStorage_DeleteRejectsSymlinkedDirPrefix(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "storage")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	victim := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(victim, []byte("keep"), 0644); err != nil {
		t.Fatalf("seed victim: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(base, "uploads")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	storage := NewLocalStorage(base)
	if err := storage.Delete("uploads/secret.txt"); err == nil {
		t.Error("expected Delete to reject a path traversing a symlinked dir out of the root")
	}
	if _, err := os.Stat(victim); err != nil {
		t.Error("Delete removed a file outside the storage root")
	}
}

// A symlinked file component in the write path must not be followed either:
// uploads/pwned.txt -> outside/pwned.txt would let an upload overwrite an
// outside file with O_TRUNC.
func TestLocalStorage_StoreRejectsSymlinkedFileEscape(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "storage")
	outside := filepath.Join(parent, "outside")
	uploads := filepath.Join(base, "uploads")
	if err := os.MkdirAll(uploads, 0755); err != nil {
		t.Fatalf("mkdir uploads: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	victim := filepath.Join(outside, "pwned.txt")
	if err := os.WriteFile(victim, []byte("keep"), 0644); err != nil {
		t.Fatalf("seed victim: %v", err)
	}
	if err := os.Symlink(victim, filepath.Join(uploads, "pwned.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	storage := NewLocalStorage(base)

	body, ct := buildMultipart(t, "file", "pwned.txt", "pwned")
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	fh := req.MultipartForm.File["file"][0]

	if _, err := storage.Store(fh, "uploads"); err == nil {
		t.Error("expected Store to reject writing through a symlink out of the root")
	}
	data, err := os.ReadFile(victim)
	if err != nil {
		t.Fatalf("stat victim: %v", err)
	}
	if string(data) != "keep" {
		t.Errorf("victim file overwritten through symlink: %q", data)
	}
}
