// Package main provides the command-line interface for the SSH Session Manager.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/antonjah/ssm/internal/config"
	"github.com/antonjah/ssm/internal/menu"
	"github.com/antonjah/ssm/internal/tmux"
)

func main() {
	hosts, err := config.GetSSHHosts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH config: %v\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Fprintf(os.Stderr, "No SSH hosts found in ~/.ssh/config\n")
		os.Exit(1)
	}

	host, err := menu.RenderMenu(hosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering menu: %v\n", err)
		os.Exit(1)
	}

	if host == "exit" {
		os.Exit(0)
	}

	// Clean up host selection (remove any trailing spaces or extra parts)
	if strings.Contains(host, " ") {
		host = strings.Split(host, " ")[0]
	}

	fmt.Printf("Connecting to %s ...\n", host)

	// Verify ssh command is available
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ssh command not found: %v\n", err)
		os.Exit(1)
	}

	// Handle tmux window management
	tmux.SSHWindow(host)

	// Execute ssh command
	err = syscall.Exec(sshPath, []string{"ssh", host}, os.Environ())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute ssh: %v\n", err)
		os.Exit(1)
	}
}
