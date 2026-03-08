package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkills_Basic(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(skillsDir, 0700)

	yaml := `name: review-pr
description: Review a pull request
type: prompt
trigger: "review PR #(\\d+)"
requires:
  tools: [file_read, grep]
  mcp_servers: [github]
prompt: |
  Review the PR carefully.
`
	os.WriteFile(filepath.Join(skillsDir, "review-pr.yaml"), []byte(yaml), 0600)

	skills, err := LoadSkills(dir, "review-agent")
	if err != nil {
		t.Fatalf("LoadSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("got %d skills, want 1", len(skills))
	}

	s := skills[0]
	if s.Name != "review-pr" {
		t.Errorf("name = %q", s.Name)
	}
	if s.Type != SkillTypePrompt {
		t.Errorf("type = %q", s.Type)
	}
	if s.Source != "review-agent" {
		t.Errorf("source = %q", s.Source)
	}
	if len(s.Requires.Tools) != 2 {
		t.Errorf("requires.tools = %v", s.Requires.Tools)
	}
	if len(s.Requires.MCPServers) != 1 {
		t.Errorf("requires.mcp_servers = %v", s.Requires.MCPServers)
	}
}

func TestLoadSkills_MissingDir(t *testing.T) {
	skills, err := LoadSkills(t.TempDir(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skills != nil {
		t.Errorf("expected nil for missing dir, got %v", skills)
	}
}

func TestLoadSkills_MissingName(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(skillsDir, 0700)
	os.WriteFile(filepath.Join(skillsDir, "bad.yaml"), []byte("type: prompt\n"), 0600)

	_, err := LoadSkills(dir, "test")
	if err == nil {
		t.Error("expected error for missing skill name")
	}
}

func TestLoadSkills_ToolChain(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(skillsDir, 0700)

	yaml := `name: build-and-test
description: Build then test
type: tool_chain
steps:
  - tool: bash
    args:
      command: "go build ./..."
    output_var: build_output
  - tool: bash
    args:
      command: "go test ./..."
`
	os.WriteFile(filepath.Join(skillsDir, "build.yaml"), []byte(yaml), 0600)

	skills, err := LoadSkills(dir, "ci-agent")
	if err != nil {
		t.Fatalf("LoadSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("got %d skills, want 1", len(skills))
	}
	if skills[0].Type != SkillTypeToolChain {
		t.Errorf("type = %q", skills[0].Type)
	}
	if len(skills[0].Steps) != 2 {
		t.Errorf("steps = %d, want 2", len(skills[0].Steps))
	}
}
