package mcp

import (
	"context"
	"fmt"
	"time"
)

// PlaywrightProbe tests whether Playwright MCP can talk to Chrome
// by calling browser_tabs with action=list.
type PlaywrightProbe struct{}

func (p *PlaywrightProbe) Probe(ctx context.Context, caller ToolCaller, serverName string) (ProbeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, isErr, err := caller.CallTool(ctx, serverName, "browser_tabs", map[string]any{"action": "list"})
	if err != nil {
		return ProbeResult{}, fmt.Errorf("browser_tabs probe failed: %w", err)
	}
	if isErr {
		return ProbeResult{Degraded: true, Detail: "browser_tabs returned error (Chrome may not be connected)"}, nil
	}
	return ProbeResult{}, nil
}
