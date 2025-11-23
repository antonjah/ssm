package tmux

import (
	"os"
	"testing"
)

func TestIsTmuxSession(t *testing.T) {
	// Test when TMUX env var is not set
	os.Unsetenv("TMUX")
	if IsTmuxSession() {
		t.Error("Expected false when TMUX env var is not set")
	}

	// Test when TMUX env var is set
	os.Setenv("TMUX", "/tmp/tmux-1000/default,1234,0")
	if !IsTmuxSession() {
		t.Error("Expected true when TMUX env var is set")
	}

	// Clean up
	os.Unsetenv("TMUX")
}
