package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kocoro-lab/shan/internal/mcp"
	"gopkg.in/yaml.v3"
)

func TestMCPDefaultsYAMLParses(t *testing.T) {
	if len(mcpDefaultsYAML) == 0 {
		t.Fatal("embedded mcp_defaults.yaml is empty")
	}

	var defaults map[string]mcp.MCPServerConfig
	if err := yaml.Unmarshal(mcpDefaultsYAML, &defaults); err != nil {
		t.Fatalf("failed to parse mcp_defaults.yaml: %v", err)
	}

	expected := []string{
		"playwright", "fetch", "memory", "sequential-thinking",
		"x-twitter", "github", "notion", "slack", "google-workspace", "postgres",
	}
	for _, name := range expected {
		srv, ok := defaults[name]
		if !ok {
			t.Errorf("missing expected server %q", name)
			continue
		}
		if !srv.Disabled {
			t.Errorf("server %q should be disabled by default", name)
		}
		if srv.Command == "" {
			t.Errorf("server %q has empty command", name)
		}
	}

	if len(defaults) != len(expected) {
		t.Errorf("expected %d servers, got %d", len(expected), len(defaults))
	}
}

func TestMergeDefaultMCPServers_EmptyUserConfig(t *testing.T) {
	cfg := &Config{}
	mergeDefaultMCPServers(cfg)

	if len(cfg.MCPServers) != 10 {
		t.Errorf("expected 10 default servers, got %d", len(cfg.MCPServers))
	}
	if _, ok := cfg.MCPServers["playwright"]; !ok {
		t.Error("missing playwright in merged defaults")
	}
}

func TestMergeDefaultMCPServers_UserOverride(t *testing.T) {
	cfg := &Config{
		MCPServers: map[string]mcp.MCPServerConfig{
			"x-twitter": {
				Command:  "npx",
				Args:     []string{"-y", "@enescinar/twitter-mcp"},
				Disabled: false, // user enabled it
				Env:      map[string]string{"API_KEY": "my-key"},
			},
		},
	}
	mergeDefaultMCPServers(cfg)

	// User's x-twitter should be preserved (not overwritten by default)
	xt := cfg.MCPServers["x-twitter"]
	if xt.Disabled {
		t.Error("user's x-twitter should remain enabled")
	}
	if xt.Env["API_KEY"] != "my-key" {
		t.Error("user's API_KEY should be preserved")
	}

	// Other defaults should still be added
	if _, ok := cfg.MCPServers["playwright"]; !ok {
		t.Error("default playwright should be merged in")
	}
	if len(cfg.MCPServers) != 10 {
		t.Errorf("expected 10 total servers, got %d", len(cfg.MCPServers))
	}
}

func TestInitMCPDefaults_FreshConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, []byte("endpoint: test\napi_key: abc\n"), 0600)

	initMCPDefaults(dir, configPath)

	data, _ := os.ReadFile(configPath)
	content := string(data)

	// Should have mcp_servers appended
	if !strings.Contains(content, "mcp_servers:") {
		t.Error("mcp_servers: not appended to config")
	}
	if !strings.Contains(content, "playwright:") {
		t.Error("playwright not in config")
	}
	if !strings.Contains(content, "x-twitter:") {
		t.Error("x-twitter not in config")
	}
	// Original content preserved
	if !strings.Contains(content, "endpoint: test") {
		t.Error("original content lost")
	}
	// Marker present
	if !strings.Contains(content, mcpDefaultsMarker) {
		t.Error("marker not present")
	}
	// No reference file needed
	if _, err := os.Stat(filepath.Join(dir, "mcp_servers.yaml")); err == nil {
		t.Error("reference file should not be created when defaults are in config")
	}

	// Idempotent
	initMCPDefaults(dir, configPath)
	data2, _ := os.ReadFile(configPath)
	if !bytes.Equal(data, data2) {
		t.Error("should not modify config on second run")
	}
}

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

func TestInitMCPDefaults_ExistingMCPServers(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	original := "endpoint: test\nmcp_servers:\n  github:\n    command: npx\n"
	os.WriteFile(configPath, []byte(original), 0600)

	initMCPDefaults(dir, configPath)

	// Config should only have marker appended (no server entries added)
	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.HasPrefix(content, original) {
		t.Error("original config content should be preserved")
	}
	if !strings.Contains(content, mcpDefaultsMarker) {
		t.Error("marker should be appended")
	}
	// No default server entries in config
	if strings.Contains(content, "playwright:") {
		t.Error("default servers should NOT be in config when user has mcp_servers")
	}

	// Reference file should be created
	refData, err := os.ReadFile(filepath.Join(dir, "mcp_servers.yaml"))
	if err != nil {
		t.Fatal("reference file should be created")
	}
	if !bytes.Equal(refData, mcpDefaultsYAML) {
		t.Error("reference file content doesn't match defaults")
	}

	// Idempotent — second run doesn't notify again
	initMCPDefaults(dir, configPath)
	data2, _ := os.ReadFile(configPath)
	if !bytes.Equal(data, data2) {
		t.Error("should not modify config on second run")
	}
}
