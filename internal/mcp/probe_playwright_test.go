package mcp

import (
	"context"
	"fmt"
	"testing"
)

type mockCallToolManager struct {
	result string
	isErr  bool
	err    error
}

func (m *mockCallToolManager) CallTool(ctx context.Context, server, tool string, args map[string]any) (string, bool, error) {
	return m.result, m.isErr, m.err
}

func TestPlaywrightProbe_Healthy(t *testing.T) {
	mgr := &mockCallToolManager{result: "tab list output"}
	probe := &PlaywrightProbe{}
	result, err := probe.Probe(context.Background(), mgr, "playwright")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Degraded {
		t.Error("expected healthy, got degraded")
	}
}

func TestPlaywrightProbe_Degraded(t *testing.T) {
	mgr := &mockCallToolManager{result: "error", isErr: true}
	probe := &PlaywrightProbe{}
	result, err := probe.Probe(context.Background(), mgr, "playwright")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Degraded {
		t.Error("expected degraded")
	}
}

func TestPlaywrightProbe_Disconnected(t *testing.T) {
	mgr := &mockCallToolManager{err: fmt.Errorf("transport: broken pipe")}
	probe := &PlaywrightProbe{}
	_, err := probe.Probe(context.Background(), mgr, "playwright")
	if err == nil {
		t.Error("expected error for disconnected")
	}
}
