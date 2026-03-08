package skills

import (
	"fmt"
	"regexp"
	"sync"
)

// SkillType defines how a skill is executed.
type SkillType string

const (
	// SkillTypePrompt injects a prompt template into the conversation.
	SkillTypePrompt SkillType = "prompt"
	// SkillTypeToolChain executes a sequence of tool calls.
	SkillTypeToolChain SkillType = "tool_chain"
	// SkillTypeSubAgent delegates to another agent.
	SkillTypeSubAgent SkillType = "sub_agent"
)

// ToolChainStep defines a single step in a tool chain skill.
type ToolChainStep struct {
	Tool      string         `yaml:"tool"`
	Args      map[string]any `yaml:"args"`
	OutputVar string         `yaml:"output_var,omitempty"` // variable name to store result
}

// Requirements declares what a skill needs to run.
type Requirements struct {
	Tools      []string `yaml:"tools,omitempty"`
	MCPServers []string `yaml:"mcp_servers,omitempty"`
}

// Skill is a composable capability that can be a prompt template,
// a multi-step tool chain, or a delegation to another agent.
type Skill struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Trigger     string       `yaml:"trigger,omitempty"` // regex pattern for auto-trigger
	Type        SkillType    `yaml:"type"`
	Requires    Requirements `yaml:"requires,omitempty"`

	// For SkillTypePrompt
	Prompt string `yaml:"prompt,omitempty"`

	// For SkillTypeToolChain
	Steps []ToolChainStep `yaml:"steps,omitempty"`

	// For SkillTypeSubAgent
	DelegateTo string `yaml:"delegate_to,omitempty"`
	Context    string `yaml:"context,omitempty"` // context to pass to sub-agent

	// Metadata
	Source string `yaml:"-"` // which agent or file defined this

	triggerRe *regexp.Regexp // compiled trigger pattern
}

// Registry holds skills and supports lookup by name or trigger matching.
type Registry struct {
	mu     sync.RWMutex
	skills map[string]*Skill
	order  []string
}

// NewRegistry creates an empty skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]*Skill),
	}
}

// Register adds a skill to the registry.
// Returns an error if a skill with the same name already exists.
func (r *Registry) Register(s *Skill) error {
	if s.Name == "" {
		return fmt.Errorf("skill name is required")
	}

	// Compile trigger regex if provided
	if s.Trigger != "" && s.triggerRe == nil {
		re, err := regexp.Compile(s.Trigger)
		if err != nil {
			return fmt.Errorf("invalid trigger pattern %q: %w", s.Trigger, err)
		}
		s.triggerRe = re
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[s.Name]; exists {
		return fmt.Errorf("skill %q already registered", s.Name)
	}

	r.skills[s.Name] = s
	r.order = append(r.order, s.Name)
	return nil
}

// Get returns a skill by name.
func (r *Registry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
	return s, ok
}

// Match finds the first skill whose trigger pattern matches the input.
// Returns nil if no skill matches.
func (r *Registry) Match(input string) *Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, name := range r.order {
		s := r.skills[name]
		if s.triggerRe != nil && s.triggerRe.MatchString(input) {
			return s
		}
	}
	return nil
}

// List returns all registered skills in registration order.
func (r *Registry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Skill, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.skills[name])
	}
	return result
}

// ForAgent returns skills registered by a specific agent.
func (r *Registry) ForAgent(agentName string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*Skill
	for _, name := range r.order {
		s := r.skills[name]
		if s.Source == agentName {
			result = append(result, s)
		}
	}
	return result
}

// Names returns the ordered list of skill names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// Len returns the number of registered skills.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}
