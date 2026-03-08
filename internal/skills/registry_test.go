package skills

import (
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	s := &Skill{Name: "review-pr", Description: "Review a PR", Type: SkillTypePrompt, Prompt: "Review this PR"}
	if err := r.Register(s); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, ok := r.Get("review-pr")
	if !ok {
		t.Fatal("skill not found")
	}
	if got.Description != "Review a PR" {
		t.Errorf("description = %q, want %q", got.Description, "Review a PR")
	}
	if r.Len() != 1 {
		t.Errorf("Len() = %d, want 1", r.Len())
	}
}

func TestRegistry_DuplicateReject(t *testing.T) {
	r := NewRegistry()
	s := &Skill{Name: "test", Type: SkillTypePrompt}
	r.Register(s)
	if err := r.Register(s); err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegistry_EmptyNameReject(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(&Skill{Type: SkillTypePrompt}); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRegistry_TriggerMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "review", Type: SkillTypePrompt, Trigger: `review PR #(\d+)`})
	r.Register(&Skill{Name: "deploy", Type: SkillTypePrompt, Trigger: `deploy to (\w+)`})

	if s := r.Match("review PR #42"); s == nil || s.Name != "review" {
		t.Errorf("expected match for 'review', got %v", s)
	}
	if s := r.Match("deploy to prod"); s == nil || s.Name != "deploy" {
		t.Errorf("expected match for 'deploy', got %v", s)
	}
	if s := r.Match("unrelated input"); s != nil {
		t.Errorf("expected no match, got %v", s.Name)
	}
}

func TestRegistry_InvalidTrigger(t *testing.T) {
	r := NewRegistry()
	err := r.Register(&Skill{Name: "bad", Type: SkillTypePrompt, Trigger: "[invalid"})
	if err == nil {
		t.Error("expected error for invalid regex trigger")
	}
}

func TestRegistry_ForAgent(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "s1", Type: SkillTypePrompt, Source: "agent-a"})
	r.Register(&Skill{Name: "s2", Type: SkillTypePrompt, Source: "agent-b"})
	r.Register(&Skill{Name: "s3", Type: SkillTypePrompt, Source: "agent-a"})

	skills := r.ForAgent("agent-a")
	if len(skills) != 2 {
		t.Errorf("got %d skills for agent-a, want 2", len(skills))
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "b", Type: SkillTypePrompt})
	r.Register(&Skill{Name: "a", Type: SkillTypePrompt})

	names := r.Names()
	if len(names) != 2 || names[0] != "b" || names[1] != "a" {
		t.Errorf("names = %v, want [b a] (registration order)", names)
	}
}
