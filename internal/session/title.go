package session

import (
	"fmt"
	"strings"
)

// AgentTitle returns a fixed title for a named agent's long-lived session.
func AgentTitle(agentName string) string {
	return fmt.Sprintf("%s conversation", agentName)
}

// Title creates a short, readable title from user input.
// Truncates to 50 chars at a word boundary, strips leading/trailing whitespace
// and newlines, and ensures single-line output.
func Title(input string) string {
	// Take first line only
	if idx := strings.IndexAny(input, "\n\r"); idx >= 0 {
		input = input[:idx]
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return "New session"
	}
	const maxLen = 50
	if len(input) <= maxLen {
		return input
	}
	// Truncate at word boundary
	truncated := input[:maxLen]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}
