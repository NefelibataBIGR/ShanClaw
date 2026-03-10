package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

func execGhosttyScript(script string) (string, error) {
	var cmdArgs []string
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cmdArgs = append(cmdArgs, "-e", trimmed)
		}
	}
	out, err := exec.Command("osascript", cmdArgs...).CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("osascript error: %w\n%s", err, result)
	}
	return result, nil
}

func ghosttyNewTab(command, title, color string) (windowIdx, tabIdx int, err error) {
	script := `tell application "Ghostty"
	activate
	tell front window
		set newTab to make new tab
	end tell
end tell`
	_, err = execGhosttyScript(script)
	if err != nil {
		return 0, 0, fmt.Errorf("new_tab: %w", err)
	}
	if title != "" {
		setTabTitle(title)
	}
	if color != "" {
		setTabColor(color)
	}
	if command != "" {
		sendKeys(command + "\n")
	}
	idxScript := `tell application "Ghostty"
	set winIdx to index of front window
	tell front window
		set tabIdx to count of tabs
	end tell
	return (winIdx as text) & "," & (tabIdx as text)
end tell`
	result, err := execGhosttyScript(idxScript)
	if err != nil {
		return 1, 1, nil
	}
	fmt.Sscanf(result, "%d,%d", &windowIdx, &tabIdx)
	return windowIdx, tabIdx, nil
}

func ghosttyNewSplit(direction, command, title, color string) (windowIdx, tabIdx int, err error) {
	dir := "vertical"
	if direction == "down" {
		dir = "horizontal"
	}
	script := fmt.Sprintf(`tell application "Ghostty"
	activate
	tell front window
		make new split with properties {direction:"%s"}
	end tell
end tell`, dir)
	_, err = execGhosttyScript(script)
	if err != nil {
		return 0, 0, fmt.Errorf("new_split: %w", err)
	}
	if title != "" {
		setTabTitle(title)
	}
	if color != "" {
		setTabColor(color)
	}
	if command != "" {
		sendKeys(command + "\n")
	}
	return 1, 1, nil
}

func ghosttySendInput(windowIdx, tabIdx int, text string) error {
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Ghostty"
	tell window %d
		tell tab %d
			write text "%s"
		end tell
	end tell
end tell`, windowIdx, tabIdx, escaped)
	_, err := execGhosttyScript(script)
	return err
}

func setTabTitle(title string) {
	escaped := strings.ReplaceAll(title, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Ghostty"
	tell front window
		set name of current tab to "%s"
	end tell
end tell`, escaped)
	execGhosttyScript(script)
}

func setTabColor(hexColor string) {
	escaped := strings.ReplaceAll(hexColor, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Ghostty"
	tell front window
		set color of current tab to "%s"
	end tell
end tell`, escaped)
	execGhosttyScript(script)
}

func sendKeys(text string) {
	escaped := strings.ReplaceAll(text, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Ghostty"
	tell front window
		write text "%s"
	end tell
end tell`, escaped)
	execGhosttyScript(script)
}

// SetGhosttyTabAppearance sets tab title and color for the current terminal.
// Exported for use from tui package.
func SetGhosttyTabAppearance(agentName string) {
	if agentName == "" {
		return
	}
	setTabTitle(agentName)
	setTabColor(agentColor(agentName))
}

func ghosttyWorkspaceScript(shanBinary string, agentNames []string) string {
	var sb strings.Builder
	sb.WriteString(`tell application "Ghostty"` + "\n")
	sb.WriteString("\tactivate\n")
	first := agentNames[0]
	color := agentColor(first)
	sb.WriteString("\tmake new window\n")
	sb.WriteString("\ttell front window\n")
	sb.WriteString(fmt.Sprintf("\t\tset name of current tab to \"%s\"\n", first))
	sb.WriteString(fmt.Sprintf("\t\tset color of current tab to \"%s\"\n", color))
	sb.WriteString(fmt.Sprintf("\t\twrite text \"%s --agent %s\"\n", shanBinary, first))
	for _, name := range agentNames[1:] {
		c := agentColor(name)
		sb.WriteString("\t\tset newTab to make new tab\n")
		sb.WriteString(fmt.Sprintf("\t\tset name of current tab to \"%s\"\n", name))
		sb.WriteString(fmt.Sprintf("\t\tset color of current tab to \"%s\"\n", c))
		sb.WriteString(fmt.Sprintf("\t\twrite text \"%s --agent %s\"\n", shanBinary, name))
	}
	sb.WriteString("\tend tell\n")
	sb.WriteString("end tell")
	return sb.String()
}

// GhosttyWorkspaceScript is the exported wrapper for cmd package.
func GhosttyWorkspaceScript(shanBinary string, agentNames []string) string {
	return ghosttyWorkspaceScript(shanBinary, agentNames)
}

// ExecGhosttyScript is the exported wrapper for cmd package.
func ExecGhosttyScript(script string) error {
	_, err := execGhosttyScript(script)
	return err
}
