package cmd

import (
	"fmt"

	"github.com/Kocoro-lab/shan/internal/tools"
	"github.com/spf13/cobra"
)

var ghosttyCmd = &cobra.Command{
	Use:   "ghostty",
	Short: "Ghostty terminal integration",
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace <agent1> [agent2] ...",
	Short: "Open a Ghostty window with one tab per agent",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tools.GhosttyAvailable() {
			return fmt.Errorf("Ghostty is not installed. Install from https://ghostty.org")
		}
		// Use "shan" (expects it in PATH via homebrew install).
		shanBin := "shan"
		script := tools.GhosttyWorkspaceScript(shanBin, args)
		if script == "" {
			return fmt.Errorf("ghostty workspace requires macOS")
		}
		return tools.ExecGhosttyScript(script)
	},
}

func init() {
	ghosttyCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(ghosttyCmd)
}
