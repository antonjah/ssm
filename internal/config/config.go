// Package config provides functionality for parsing SSH configuration files.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Host represents an SSH host configuration entry.
type Host struct {
	// Alias is the host alias/name used in the SSH config (e.g., "myserver").
	Alias string
	// HostName is the actual hostname or IP address to connect to.
	HostName string
}

// GetSSHHosts reads SSH hosts from the default configuration file (~/.ssh/config).
func GetSSHHosts() ([]Host, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(home, ".ssh", "config")
	return GetSSHHostsFromPath(configPath)
}

// parseSSHConfig parses SSH configuration from the provided reader and returns a map of hosts.
func parseSSHConfig(scanner *bufio.Scanner) (map[string]Host, error) {
	hosts := make(map[string]Host)
	var currentHost string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.ToLower(parts[0])
			value := strings.Join(parts[1:], " ")

			if key == "host" && value != "*" {
				currentHost = value
				hosts[currentHost] = Host{Alias: currentHost}
			} else if currentHost != "" {
				if key == "hostname" {
					host := hosts[currentHost]
					host.HostName = value
					hosts[currentHost] = host
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read SSH config: %w", err)
	}

	return hosts, nil
}

// GetSSHHostsFromPath reads SSH hosts from the specified configuration file path.
func GetSSHHostsFromPath(configPath string) ([]Host, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config file %q: %w", configPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hosts, err := parseSSHConfig(scanner)
	if err != nil {
		return nil, err
	}

	var result []Host
	for _, host := range hosts {
		result = append(result, host)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Alias < result[j].Alias
	})

	return result, nil
}
