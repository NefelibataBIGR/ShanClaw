package context

import (
	"testing"

	"github.com/Kocoro-lab/ShanClaw/internal/client"
)

func TestSanitizeHistory_Empty(t *testing.T) {
	result := SanitizeHistory(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
	result = SanitizeHistory([]client.Message{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestSanitizeHistory_CleanPassthrough(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("hi there")},
		{Role: "user", Content: client.NewTextContent("how are you")},
		{Role: "assistant", Content: client.NewTextContent("doing well")},
	}
	result := SanitizeHistory(msgs)
	if len(result) != 4 {
		t.Fatalf("expected 4, got %d", len(result))
	}
	for i, m := range result {
		if m.Role != msgs[i].Role || m.Content.Text() != msgs[i].Content.Text() {
			t.Errorf("msg %d mismatch", i)
		}
	}
}

func TestSanitizeHistory_DropsToolCallPlaceholders(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("search for X")},
		{Role: "assistant", Content: client.NewTextContent("[tool_call: web_search]")},
		{Role: "assistant", Content: client.NewTextContent("[tool_call: web_search]")},
		{Role: "assistant", Content: client.NewTextContent("here are the results")},
	}
	result := SanitizeHistory(msgs)
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[0].Role != "user" {
		t.Errorf("expected user, got %s", result[0].Role)
	}
	if result[1].Content.Text() != "here are the results" {
		t.Errorf("expected final assistant text, got %q", result[1].Content.Text())
	}
}

func TestSanitizeHistory_DropsPlainTextToolMessages(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("let me search")},
		{Role: "tool", Content: client.NewTextContent("Search results for: shoes")},
		{Role: "assistant", Content: client.NewTextContent("found some shoes")},
	}
	result := SanitizeHistory(msgs)
	// tool msg dropped, consecutive assistants merged → keep last
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[1].Content.Text() != "found some shoes" {
		t.Errorf("expected merged assistant, got %q", result[1].Content.Text())
	}
}

func TestSanitizeHistory_DropsErrorMessages(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("Sorry, the AI service encountered a temporary error. Please try again.")},
		{Role: "user", Content: client.NewTextContent("try again")},
		{Role: "assistant", Content: client.NewTextContent("Sorry, the AI service encountered a temporary error. Please try again.")},
	}
	result := SanitizeHistory(msgs)
	// Both error assistants dropped, consecutive users merged → keep last
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0].Content.Text() != "try again" {
		t.Errorf("expected last user, got %q", result[0].Content.Text())
	}
}

func TestSanitizeHistory_DropsAgentFailedError(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("Say hi")},
		{Role: "assistant", Content: client.NewTextContent("[error: agent failed to respond]")},
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("hi!")},
	}
	result := SanitizeHistory(msgs)
	// error dropped, consecutive users merged → keep last
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[0].Content.Text() != "hello" {
		t.Errorf("expected second user msg, got %q", result[0].Content.Text())
	}
}

func TestSanitizeHistory_MergesConsecutiveAssistant(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("first response")},
		{Role: "assistant", Content: client.NewTextContent("second response")},
		{Role: "assistant", Content: client.NewTextContent("third response")},
	}
	result := SanitizeHistory(msgs)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[1].Content.Text() != "third response" {
		t.Errorf("expected last assistant kept, got %q", result[1].Content.Text())
	}
}

func TestSanitizeHistory_MergesConsecutiveUser(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("first")},
		{Role: "user", Content: client.NewTextContent("second")},
		{Role: "user", Content: client.NewTextContent("third")},
		{Role: "assistant", Content: client.NewTextContent("got it")},
	}
	result := SanitizeHistory(msgs)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0].Content.Text() != "third" {
		t.Errorf("expected last user kept, got %q", result[0].Content.Text())
	}
}

func TestSanitizeHistory_FullCorruptionScenario(t *testing.T) {
	// Reproduce the exact little-v corruption pattern
	msgs := []client.Message{
		{Role: "user", Content: client.NewTextContent("check rankings")},
		{Role: "assistant", Content: client.NewTextContent("The request was cancelled or timed out.")},
		{Role: "user", Content: client.NewTextContent("heartbeat prompt")},
		{Role: "assistant", Content: client.NewTextContent("I noticed the timeout")},
		{Role: "assistant", Content: client.NewTextContent("urgent confirmation")},
		{Role: "assistant", Content: client.NewTextContent("[tool_call: web_search]")},
		{Role: "assistant", Content: client.NewTextContent("[tool_call: web_search]")},
		{Role: "assistant", Content: client.NewTextContent("[tool_call: web_search]")},
		{Role: "tool", Content: client.NewTextContent("Search results 1")},
		{Role: "tool", Content: client.NewTextContent("Search results 2")},
		{Role: "tool", Content: client.NewTextContent("Search results 3")},
		{Role: "assistant", Content: client.NewTextContent("Here are the rankings")},
		{Role: "user", Content: client.NewTextContent("你好呀！")},
		{Role: "assistant", Content: client.NewTextContent("Sorry, the AI service encountered a temporary error. Please try again.")},
		{Role: "user", Content: client.NewTextContent("今天有没有什么更新")},
		{Role: "assistant", Content: client.NewTextContent("Sorry, the AI service encountered a temporary error. Please try again.")},
	}

	result := SanitizeHistory(msgs)

	// Verify alternation
	for i := 1; i < len(result); i++ {
		if result[i].Role == result[i-1].Role {
			t.Errorf("consecutive same role at %d: %s", i, result[i].Role)
		}
	}

	// Should not contain any error or tool_call messages
	for i, m := range result {
		text := m.Content.Text()
		if m.Role == "tool" {
			t.Errorf("tool message at %d should be dropped", i)
		}
		if text == "Sorry, the AI service encountered a temporary error. Please try again." {
			t.Errorf("error message at %d should be dropped", i)
		}
		if text == "[tool_call: web_search]" {
			t.Errorf("tool_call placeholder at %d should be dropped", i)
		}
	}

	// Verify we kept meaningful content
	found := false
	for _, m := range result {
		if m.Content.Text() == "Here are the rankings" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Here are the rankings' to survive")
	}

	t.Logf("sanitized %d → %d messages", len(msgs), len(result))
	for i, m := range result {
		t.Logf("  [%d] %s: %s", i, m.Role, truncStr(m.Content.Text(), 50))
	}
}

func TestSanitizeHistory_PreservesSystemMessages(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: client.NewTextContent("you are helpful")},
		{Role: "user", Content: client.NewTextContent("hello")},
		{Role: "assistant", Content: client.NewTextContent("hi")},
	}
	result := SanitizeHistory(msgs)
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0].Role != "system" {
		t.Errorf("system message should be preserved")
	}
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
