package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillDetail is the API response type for GET /skills/{name}.
// Includes prompt body and source, unlike SkillMeta (metadata only)
// or Skill (which hides Source/Dir via json:"-" tags).
type SkillDetail struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Prompt        string            `json:"prompt"`
	Source        string            `json:"source"`
	License       string            `json:"license,omitempty"`
	Compatibility string            `json:"compatibility,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedTools  []string          `json:"allowed_tools,omitempty"`
}

// WriteGlobalSkill writes a skill to the global skills directory
// (~/.shannon/skills/<name>/SKILL.md). Same atomic write pattern
// as agents.WriteAgentSkill but different path root.
func WriteGlobalSkill(shannonDir string, skill *Skill) error {
	dir := filepath.Join(shannonDir, "skills", skill.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	fm := skillFrontmatter{
		Name:        skill.Name,
		Description: skill.Description,
		License:     skill.License,
		Compatibility: skill.Compatibility,
		Metadata:    skill.Metadata,
	}
	if len(skill.AllowedTools) > 0 {
		fm.AllowedTools = strings.Join(skill.AllowedTools, " ")
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(skill.Prompt)
	if !strings.HasSuffix(skill.Prompt, "\n") {
		buf.WriteString("\n")
	}

	return atomicWrite(filepath.Join(dir, "SKILL.md"), []byte(buf.String()))
}

// DeleteGlobalSkill removes a global skill directory.
func DeleteGlobalSkill(shannonDir, name string) error {
	if err := ValidateSkillName(name); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(shannonDir, "skills", name))
}

// atomicWrite writes data to a temp file then renames to path.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
