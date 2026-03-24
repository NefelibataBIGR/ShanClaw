package context

import (
	"strings"

	"github.com/Kocoro-lab/ShanClaw/internal/client"
)

// SanitizeHistory repairs malformed message history that would cause API errors.
// Specifically handles:
//   - tool role messages with plain text (no tool_result blocks) → dropped
//   - assistant messages that are just "[tool_call: ...]" placeholders → dropped
//   - consecutive assistant messages without intervening user → merged into one
//   - assistant error messages (FriendlyAgentError output) → dropped
//
// Returns a new slice; the original is not modified.
func SanitizeHistory(messages []client.Message) []client.Message {
	if len(messages) == 0 {
		return messages
	}

	// First pass: drop invalid messages.
	var cleaned []client.Message
	for _, msg := range messages {
		if shouldDrop(msg) {
			continue
		}
		cleaned = append(cleaned, msg)
	}

	// Second pass: fix consecutive same-role messages.
	// Claude API requires strict user/assistant alternation.
	var result []client.Message
	for i, msg := range cleaned {
		if i > 0 && msg.Role == cleaned[i-1].Role {
			switch msg.Role {
			case "assistant":
				// Keep the later one (drop the earlier, already in result).
				result[len(result)-1] = msg
				continue
			case "user":
				// Keep the later one (drop the earlier).
				result[len(result)-1] = msg
				continue
			}
		}
		result = append(result, msg)
	}

	return result
}

// shouldDrop returns true for messages that are malformed or would cause API errors.
func shouldDrop(msg client.Message) bool {
	text := msg.Content.Text()

	switch msg.Role {
	case "tool":
		// tool messages must have tool_result content blocks.
		// Plain-text tool messages are from buggy heartbeat persistence.
		if !msg.Content.HasBlocks() {
			return true
		}
		// Even with blocks, check they contain tool_result type.
		hasToolResult := false
		for _, b := range msg.Content.Blocks() {
			if b.Type == "tool_result" {
				hasToolResult = true
				break
			}
		}
		return !hasToolResult

	case "assistant":
		// Drop placeholder tool call text (from old heartbeat bug).
		if strings.HasPrefix(text, "[tool_call:") {
			return true
		}
		// Drop error marker from old heartbeat failures.
		if text == "[error: agent failed to respond]" {
			return true
		}
		// Drop persisted FriendlyAgentError messages — they contain no useful
		// context and just waste tokens.
		if isFriendlyError(text) {
			return true
		}
	}

	return false
}

// isFriendlyError returns true for messages produced by FriendlyAgentError.
func isFriendlyError(text string) bool {
	switch text {
	case "The request was cancelled or timed out.",
		"Sorry, the AI service is currently rate-limited. Please try again in a moment.",
		"Sorry, the AI service is temporarily overloaded. Please try again shortly.",
		"Sorry, the AI service encountered a temporary error. Please try again.",
		"Sorry, the connection to the AI service was interrupted. Please try again.",
		"Sorry, an unexpected error occurred. Please try again.":
		return true
	}
	return false
}
