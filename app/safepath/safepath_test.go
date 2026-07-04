package safepath_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/heidericklucas/go-security-swe-bench/app/safepath"
)

func TestOpen_ServesFileInsideBase(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "avatar.png"), []byte("PNGDATA"), 0o600); err != nil {
		t.Fatal(err)
	}
	rc, err := safepath.Open(base, "avatar.png")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if string(b) != "PNGDATA" {
		t.Fatalf("want PNGDATA, got %q", b)
	}
}
