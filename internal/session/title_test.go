package session

import "testing"

func TestTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello world"},
		{"", "New session"},
		{"   ", "New session"},
		{"line1\nline2", "line1"},
		{"a]very long input that exceeds fifty characters and should be truncated at word boundary", "a]very long input that exceeds fifty characters..."},
		{"short\r\nwith crlf", "short"},
	}
	for _, tt := range tests {
		got := Title(tt.input)
		if got != tt.want {
			t.Errorf("Title(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAgentTitle(t *testing.T) {
	if got := AgentTitle("ops-bot"); got != "ops-bot conversation" {
		t.Errorf("AgentTitle(ops-bot) = %q", got)
	}
}
