package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhiramnajith/html-artifacts/server/internal/storage"
)

// newTestServer builds a Server over a temp artifacts dir seeded with one
// artifact, and returns the handler plus the seeded id.
func newTestServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	id := "react-vs-vue-20260721-103000"
	body := "<!doctype html><html><body><h1>React vs Vue</h1></body></html>"
	if err := os.WriteFile(filepath.Join(dir, id+".html"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	srv, err := New(storage.New(dir))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv.Handler(), id
}

func do(t *testing.T, h http.Handler, method, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func TestRootRedirectsToArtifacts(t *testing.T) {
	h, _ := newTestServer(t)
	resp := do(t, h, "GET", "/")
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("GET /: want redirect, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/artifacts" {
		t.Fatalf("GET /: want Location /artifacts, got %q", loc)
	}
}

func TestArtifactsIndexListsArtifacts(t *testing.T) {
	h, id := newTestServer(t)
	resp := do(t, h, "GET", "/artifacts")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /artifacts: want 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, id) {
		t.Fatalf("GET /artifacts: body does not list id %q", id)
	}
	if !strings.Contains(body, "/view/"+id) {
		t.Fatalf("GET /artifacts: body has no link to /view/%s", id)
	}
}

func TestViewServesArtifactWithInjectedShell(t *testing.T) {
	h, id := newTestServer(t)
	resp := do(t, h, "GET", "/view/"+id)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /view/%s: want 200, got %d", id, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("GET /view: want text/html, got %q", ct)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "React vs Vue") {
		t.Fatal("GET /view: artifact content missing")
	}
	if !strings.Contains(body, "/_editor/shell.js") {
		t.Fatal("GET /view: editor shell was not injected")
	}
}

func TestViewUnknownIDReturns404(t *testing.T) {
	h, _ := newTestServer(t)
	resp := do(t, h, "GET", "/view/nope-20260101-000000")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /view/<missing>: want 404, got %d", resp.StatusCode)
	}
}

func TestViewTraversalIsRefused(t *testing.T) {
	h, _ := newTestServer(t)
	// Percent-encoded and literal traversal attempts must never 200 or leak a
	// file outside the artifacts dir.
	for _, bad := range []string{
		"/view/..%2f..%2fetc%2fpasswd",
		"/view/..",
		"/view/A",   // uppercase — invalid id
		"/view/a_b", // underscore — invalid id
	} {
		resp := do(t, h, "GET", bad)
		if resp.StatusCode == http.StatusOK {
			t.Fatalf("GET %s: traversal/invalid id returned 200", bad)
		}
	}
}

func TestEditorShellIsServed(t *testing.T) {
	h, _ := newTestServer(t)
	resp := do(t, h, "GET", "/_editor/shell.js")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /_editor/shell.js: want 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("GET /_editor/shell.js: want javascript content-type, got %q", ct)
	}
}

func TestListenAddrIsLoopbackOnly(t *testing.T) {
	// The server must never be reachable off-host: the bind address is always
	// 127.0.0.1, never a wildcard like ":7777" or "0.0.0.0".
	got := ListenAddr(7777)
	if got != "127.0.0.1:7777" {
		t.Fatalf("ListenAddr(7777) = %q, want 127.0.0.1:7777", got)
	}
	if strings.HasPrefix(got, ":") || strings.HasPrefix(got, "0.0.0.0") {
		t.Fatalf("ListenAddr must not bind a wildcard interface, got %q", got)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
