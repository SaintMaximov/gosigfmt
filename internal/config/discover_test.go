package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	cfg, path, err := FindConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("want empty path, got %q", path)
	}
	if cfg.LineLength != 100 {
		t.Errorf("want defaults; LineLength=%d", cfg.LineLength)
	}
}

func TestFindConfig_FoundInSameDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".gosigfmt.yaml")
	if err := os.WriteFile(cfgPath, []byte("line_length: 60\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := FindConfig(dir)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if path != cfgPath {
		t.Errorf("path: want %q, got %q", cfgPath, path)
	}
	if cfg.LineLength != 60 {
		t.Errorf("LineLength: want 60, got %d", cfg.LineLength)
	}
}

func TestFindConfig_WalksUp(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, ".gosigfmt.yaml")
	if err := os.WriteFile(cfgPath, []byte("line_length: 60\n"), 0644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := FindConfig(deep)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if path != cfgPath {
		t.Errorf("path: want %q, got %q", cfgPath, path)
	}
	if cfg.LineLength != 60 {
		t.Errorf("LineLength: want 60, got %d", cfg.LineLength)
	}
}
