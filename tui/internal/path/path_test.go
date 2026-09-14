package path

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveEnvOverride(t *testing.T) {
	t.Setenv(EnvVar, "/custom/chess")
	got := Resolve("/any/base")
	if got != "/custom/chess" {
		t.Fatalf("Resolve() = %q, want /custom/chess", got)
	}
}

func TestResolveDefault(t *testing.T) {
	t.Setenv(EnvVar, "")
	base := t.TempDir()
	got := Resolve(base)
	want := filepath.Clean(filepath.Join(base, DefaultRelative))
	if got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestValidateMissing(t *testing.T) {
	err := Validate(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error %q should name missing binary", err)
	}
}

func TestValidateOK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "chess")
	if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Validate(p); err != nil {
		t.Fatal(err)
	}
}
