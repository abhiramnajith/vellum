package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadArtifactRejectsInvalidID(t *testing.T) {
	// A secret file living OUTSIDE the artifacts dir. No invalid id must ever
	// resolve to it (or anywhere outside the dir).
	base := t.TempDir()
	artDir := filepath.Join(base, "artifacts")
	if err := os.MkdirAll(artDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret.html"), []byte("TOP SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := New(artDir)

	tests := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"dotdot", ".."},
		{"parent traversal", "../secret"},
		{"nested traversal", "../../etc/passwd"},
		{"forward slash", "a/b"},
		{"back slash", `a\b`},
		{"absolute path", "/etc/passwd"},
		{"single dot", "."},
		{"leading dot", ".hidden"},
		{"embedded dots", "a..b"},
		{"uppercase", "React"},
		{"underscore", "a_b"},
		{"space", "a b"},
		{"unicode", "café"},
		{"null byte", "a\x00b"},
		{"dotdot with html", "..%2fsecret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.ReadArtifact(tt.id)
			if !errors.Is(err, ErrInvalidID) {
				t.Fatalf("ReadArtifact(%q): want ErrInvalidID, got %v", tt.id, err)
			}
		})
	}
}

func TestReadArtifactReadsValidArtifact(t *testing.T) {
	dir := t.TempDir()
	id := "react-vs-vue-20260721-103000"
	want := []byte("<!doctype html><title>hi</title>")
	if err := os.WriteFile(filepath.Join(dir, id+".html"), want, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := New(dir).ReadArtifact(id)
	if err != nil {
		t.Fatalf("ReadArtifact: unexpected error %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("ReadArtifact: got %q, want %q", got, want)
	}
}

func TestReadArtifactMissingReturnsNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := New(dir).ReadArtifact("does-not-exist-20260721-000000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestListReturnsArtifactsSortedNewestFirst(t *testing.T) {
	dir := t.TempDir()
	// Only .html files are artifacts; annotation/other files are ignored.
	for _, name := range []string{"a-1.html", "b-2.html", "c-3.html", "notes.txt", "a-1.annotations.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := New(dir).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("List: want 3 artifacts, got %d (%v)", len(got), got)
	}
	ids := map[string]bool{}
	for _, a := range got {
		ids[a.ID] = true
		if a.ID == "" {
			t.Fatal("artifact has empty id")
		}
	}
	for _, want := range []string{"a-1", "b-2", "c-3"} {
		if !ids[want] {
			t.Fatalf("List: missing artifact %q in %v", want, got)
		}
	}
}

func TestListOnMissingDirReturnsEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not-created-yet")
	got, err := New(dir).List()
	if err != nil {
		t.Fatalf("List on missing dir: want nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List on missing dir: want empty, got %v", got)
	}
}
