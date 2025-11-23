package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSSHHosts(t *testing.T) {
	// Create a temporary SSH config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	configContent := `# This is a comment
Host *
    User default

Host server1
    HostName 192.168.1.100

Host server2
    HostName example.com

# Another comment
Host server3
    HostName test.com
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	hosts, err := GetSSHHostsFromPath(configPath)
	if err != nil {
		t.Fatalf("GetSSHHostsFromPath failed: %v", err)
	}

	expected := []Host{
		{Alias: "server1", HostName: "192.168.1.100"},
		{Alias: "server2", HostName: "example.com"},
		{Alias: "server3", HostName: "test.com"},
	}

	if len(hosts) != len(expected) {
		t.Fatalf("Expected %d hosts, got %d", len(expected), len(hosts))
	}

	for i, host := range hosts {
		if host != expected[i] {
			t.Errorf("Expected host %+v, got %+v", expected[i], host)
		}
	}
}

func TestGetSSHHosts_NoConfig(t *testing.T) {
	_, err := GetSSHHostsFromPath("/non/existent/path")
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
}
