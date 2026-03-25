package mcp

import (
	"context"
	"testing"
	"time"
)

func TestConnectAll_StoresConfigOnFailure(t *testing.T) {
	mgr := NewClientManager()
	servers := map[string]MCPServerConfig{
		"bad": {Command: "/nonexistent/binary"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := mgr.ConnectAll(ctx, servers)
	if err == nil {
		t.Fatal("expected error for bad command")
	}
	mgr.mu.Lock()
	_, hasCfg := mgr.configs["bad"]
	mgr.mu.Unlock()
	if !hasCfg {
		t.Error("expected config to be stored for failed server")
	}
}
