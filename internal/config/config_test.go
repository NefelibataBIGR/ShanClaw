package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendAllowedCommand(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("endpoint: https://example.com\n"), 0644)

	err := AppendAllowedCommand(dir, "git status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfgPath)
	content := string(data)
	if !strings.Contains(content, "git status") {
		t.Errorf("should contain 'git status', got:\n%s", content)
	}

	// Append another
	err = AppendAllowedCommand(dir, "ls -la")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ = os.ReadFile(cfgPath)
	content = string(data)
	if !strings.Contains(content, "ls -la") {
		t.Errorf("should contain 'ls -la', got:\n%s", content)
	}
	if !strings.Contains(content, "git status") {
		t.Errorf("should still contain 'git status', got:\n%s", content)
	}
}

func TestAppendAllowedCommand_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("permissions:\n  allowed_commands:\n    - \"git status\"\n"), 0644)

	err := AppendAllowedCommand(dir, "git status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfgPath)
	if strings.Count(string(data), "git status") > 1 {
		t.Error("should not duplicate existing command")
	}
}
