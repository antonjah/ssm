// Package tmux provides functionality for managing tmux windows during SSH sessions.
package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// IsTmuxSession returns true if the current process is running inside a tmux session.
func IsTmuxSession() bool {
	return os.Getenv("TMUX") != ""
}

// SSHWindow creates or switches to a tmux window for the given SSH host.
// If a window with the name "ssh:<host>" already exists, it switches to it.
// Otherwise, it creates a new window with that name.
func SSHWindow(host string) {
	if !IsTmuxSession() {
		return
	}

	cmd := exec.Command("tmux", "list-windows", "-F", "#{window_index},#{window_name}")
	output, err := cmd.Output()
	if err != nil {
		// tmux not available or error, skip silently
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) == 2 && parts[1] == "ssh:"+host {
			tmuxPath, err := exec.LookPath("tmux")
			if err != nil {
				return
			}
			syscall.Exec(tmuxPath, []string{"tmux", "select-window", "-t", parts[0]}, os.Environ())
		}
	}

	// Create new window
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return
	}
	syscall.Exec(tmuxPath, []string{"tmux", "new-window", "-n", "ssh:" + host, "ssh", host}, os.Environ())
}
