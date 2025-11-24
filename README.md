# ssm (ssh-menu)

A simple terminal-based SSH host selector written in Go. It parses your `~/.ssh/config` file and presents a interactive menu to choose which host to connect to.

## Features

- Parses SSH config files automatically
- Interactive TUI menu using [Bubbletea](https://github.com/charmbracelet/bubbletea)
- Tmux integration: creates new windows for SSH sessions when running inside tmux
- Fast and lightweight

## Installation

```bash
go install github.com/antonjah/ssm/cmd/ssm@latest
```

## Usage

1. Ensure your SSH config is set up at `~/.ssh/config` with host entries:

```sshconfig
Host server1
    HostName 192.168.1.100
    User myuser

Host server2
    HostName example.com
    User anotheruser
    IdentityFile /home/foo/.ssh/id_rsa
```

2. Run the program:

```bash
ssm
```

3. Use arrow keys to navigate, Enter to select, or type to filter hosts.

4. The program will connect to the selected host using SSH.

### Tmux Integration

When running inside a tmux session, the program will:
- Create a new tmux window for the SSH session
- Switch to existing window if one already exists for that host
- Name windows as `ssh:hostname`