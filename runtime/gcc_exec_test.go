package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGCCExecutionFixtureDirectoryExists(t *testing.T) {
	path := filepath.Join("testdata", "gcc-exec")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}
