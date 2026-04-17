package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadText_EmptyFileReturnsBody(t *testing.T) {
	got, err := readText("", "fallback body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fallback body" {
		t.Errorf("got %q, want %q", got, "fallback body")
	}
}

func TestReadText_EmptyFileEmptyBody(t *testing.T) {
	got, err := readText("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestReadText_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "desc.md")
	content := "# Heading\n\nBody text.\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	got, err := readText(path, "ignored body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestReadText_FileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.md")
	got, err := readText(path, "ignored body")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if got != "" {
		t.Errorf("expected empty result on error, got %q", got)
	}
}

func TestReadText_FromStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		r.Close()
	})

	input := "piped description\nsecond line\n"
	go func() {
		defer w.Close()
		_, _ = w.WriteString(input)
	}()

	got, err := readText("-", "ignored body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestReadText_FileTakesPrecedenceOverBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "content.txt")
	if err := os.WriteFile(path, []byte("from file"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	got, err := readText(path, "from body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from file" {
		t.Errorf("got %q, want %q", got, "from file")
	}
}
