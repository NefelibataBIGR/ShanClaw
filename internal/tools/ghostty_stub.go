//go:build !darwin

package tools

import "fmt"

var errNotDarwin = fmt.Errorf("ghostty integration requires macOS")

func ghosttyNewTab(command, title, color string) (int, int, error)             { return 0, 0, errNotDarwin }
func ghosttyNewSplit(direction, command, title, color string) (int, int, error) { return 0, 0, errNotDarwin }
func ghosttySendInput(windowIdx, tabIdx int, text string) error                { return errNotDarwin }
func SetGhosttyTabAppearance(agentName string)                                 {}
func ghosttyWorkspaceScript(shanBinary string, agentNames []string) string     { return "" }
func GhosttyWorkspaceScript(shanBinary string, agentNames []string) string     { return "" }
func ExecGhosttyScript(script string) error                                    { return errNotDarwin }
