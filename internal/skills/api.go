package skills

import (
	"fmt"
	"os"
	"os/exec"
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
		Name:          skill.Name,
		Description:   skill.Description,
		License:       skill.License,
		Compatibility: skill.Compatibility,
		Metadata:      skill.Metadata,
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

// DownloadableSkill describes a skill available for download from Anthropic's repo.
type DownloadableSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

// DownloadableSkills is the registry of skills that can be downloaded on demand.
// These have a proprietary license and cannot be bundled, but users can
// download them directly from Anthropic's repository.
var DownloadableSkills = []struct {
	Name        string
	Description string
}{
	{"docx", "Document creation, editing, and analysis with tracked changes and comments"},
	{"pdf", "PDF extraction, creation, merging, splitting, and form filling"},
	{"pptx", "Presentation creation, editing, and analysis"},
	{"xlsx", "Spreadsheet creation, editing, analysis with formulas and formatting"},
}

// IsDownloadable returns true if the skill name is in the downloadable registry.
func IsDownloadable(name string) bool {
	for _, s := range DownloadableSkills {
		if s.Name == name {
			return true
		}
	}
	return false
}

// InstallSkillFromRepo downloads a skill from Anthropic's skills repo
// into the global skills directory (~/.shannon/skills/<name>/).
// Uses git sparse checkout to fetch only the requested skill directory.
func InstallSkillFromRepo(shannonDir, name string) error {
	if err := ValidateSkillName(name); err != nil {
		return err
	}
	if !IsDownloadable(name) {
		return fmt.Errorf("skill %q is not available for download", name)
	}

	destDir := filepath.Join(shannonDir, "skills", name)
	if _, err := os.Stat(filepath.Join(destDir, "SKILL.md")); err == nil {
		return fmt.Errorf("skill %q is already installed", name)
	}

	tmpDir, err := os.MkdirTemp(shannonDir, "skill-install-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := runGit(tmpDir, "clone", "--depth=1", "--filter=blob:none", "--sparse",
		"https://github.com/anthropics/skills.git", "."); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	if err := runGit(tmpDir, "sparse-checkout", "set", "skills/"+name); err != nil {
		return fmt.Errorf("git sparse-checkout: %w", err)
	}

	srcDir := filepath.Join(tmpDir, "skills", name)
	if _, err := os.Stat(filepath.Join(srcDir, "SKILL.md")); err != nil {
		return fmt.Errorf("skill %q not found in Anthropic repo", name)
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0700); err != nil {
		return err
	}
	return os.Rename(srcDir, destDir)
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// atomicWrite writes data to a temp file then renames to path.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
