// Package menu provides an interactive terminal UI for selecting SSH hosts.
package menu

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/antonjah/ssm/internal/config"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Use Catppuccin Mocha flavor
var mocha = catppuccin.Mocha

var docStyle = lipgloss.NewStyle().Margin(1, 2)

var titleCaser = cases.Title(language.English)

// sshKeyCapitalization maps lowercase SSH config keys to their proper capitalization
var sshKeyCapitalization = map[string]string{
	"hostname":              "HostName",
	"identityfile":          "IdentityFile",
	"identitiesonly":        "IdentitiesOnly",
	"proxycommand":          "ProxyCommand",
	"controlmaster":         "ControlMaster",
	"controlpath":           "ControlPath",
	"controlpersist":        "ControlPersist",
	"userknownhostsfile":    "UserKnownHostsFile",
	"globalknownhostsfile":  "GlobalKnownHostsFile",
	"stricthostkeychecking": "StrictHostKeyChecking",
	"user":                  "User",
	"port":                  "Port",
	"forwardagent":          "ForwardAgent",
	"forwardx11":            "ForwardX11",
	"compression":           "Compression",
	"serveraliveinterval":   "ServerAliveInterval",
	"serveralivecountmax":   "ServerAliveCountMax",
}

// capitalizeSSHKey returns the properly capitalized version of an SSH config key
func capitalizeSSHKey(key string) string {
	if proper, exists := sshKeyCapitalization[key]; exists {
		return proper
	}
	// Fallback to title case for unknown keys
	return titleCaser.String(key)
}

const (
	// defaultListWidth is the default width of the host selection list
	defaultListWidth = 20
	// defaultListHeight is the default height of the host selection list
	defaultListHeight = 10
)

// customDelegate provides a custom list delegate with additional keybinding help.
type customDelegate struct {
	defaultDelegate list.DefaultDelegate
}

func newCustomDelegate() customDelegate {
	d := list.NewDefaultDelegate()

	// Apply Catppuccin Mocha colors
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color(mocha.Mauve().Hex)).
		BorderForeground(lipgloss.Color(mocha.Mauve().Hex))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color(mocha.Subtext0().Hex)).
		BorderForeground(lipgloss.Color(mocha.Mauve().Hex))
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(lipgloss.Color(mocha.Text().Hex))
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(lipgloss.Color(mocha.Subtext1().Hex))
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		Foreground(lipgloss.Color(mocha.Overlay0().Hex))
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(lipgloss.Color(mocha.Overlay0().Hex))

	return customDelegate{defaultDelegate: d}
}

func (d customDelegate) Height() int {
	return d.defaultDelegate.Height()
}

func (d customDelegate) Spacing() int {
	return d.defaultDelegate.Spacing()
}

func (d customDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.defaultDelegate.Update(msg, m)
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	d.defaultDelegate.Render(w, m, index, item)
}

func (d customDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit config")),
		key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view details")),
	}
}

func (d customDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit config")),
			key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view details")),
		},
		{
			key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
			key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
		},
	}
}

// popupKeyMap provides key bindings for the popup view.
type popupKeyMap struct{}

func (p popupKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "go back")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func (p popupKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "go back")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		},
	}
}

// HostItem represents a selectable SSH host item in the list.
type HostItem struct {
	host config.Host
}

// FilterValue returns the value used for filtering the item.
func (i HostItem) FilterValue() string { return i.host.Alias }

// Title returns the display title for the item.
func (i HostItem) Title() string { return i.host.Alias }

// Description returns the description for the item.
func (i HostItem) Description() string { return i.host.HostName }

// Model represents the state of the SSH host selection menu.
type Model struct {
	list        list.Model
	help        help.Model
	choice      string
	done        bool
	viewing     bool
	hostDetails *HostDetails
	width       int
	height      int
}

// NewModel creates a new menu model with the given SSH hosts.
func NewModel(hosts []config.Host) Model {
	hostItems := make([]list.Item, len(hosts))
	for index, host := range hosts {
		hostItems[index] = HostItem{host: host}
	}

	hostList := list.New(hostItems, newCustomDelegate(), defaultListWidth, defaultListHeight)
	hostList.SetFilteringEnabled(true)
	hostList.SetShowTitle(false)

	// Apply Catppuccin Mocha colors to list styles
	hostList.Styles.Title = hostList.Styles.Title.
		Foreground(lipgloss.Color(mocha.Mauve().Hex)).
		Bold(true)
	hostList.Styles.FilterPrompt = hostList.Styles.FilterPrompt.
		Foreground(lipgloss.Color(mocha.Mauve().Hex))
	hostList.Styles.FilterCursor = hostList.Styles.FilterCursor.
		Foreground(lipgloss.Color(mocha.Pink().Hex))

	return Model{
		list: hostList,
		help: help.New(),
	}
}

// Init initializes the Bubble Tea model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(HostItem); ok {
				m.choice = item.Title()
				m.done = true
				return m, tea.Quit
			}
		case "e":
			return m, m.openEditor()
		case "v":
			return m, m.showHostDetails()
		case "esc":
			if m.viewing {
				m.viewing = false
				m.hostDetails = nil
				return m, nil
			}
			// Exit the application when pressing esc at the main menu
			m.choice = "exit"
			m.done = true
			return m, tea.Quit
		case "q", "ctrl+c":
			m.choice = "exit"
			m.done = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.done {
		return ""
	}

	if m.viewing {
		popupContent := m.createPopupView()

		popupStyle := lipgloss.NewStyle().
			Margin(1, 2).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(mocha.Mauve().Hex)).
			Foreground(lipgloss.Color(mocha.Text().Hex)).
			Background(lipgloss.Color(mocha.Base().Hex)).
			Width(60) // Fixed width for better centering

		styledPopup := popupStyle.Render(popupContent)
		centeredPopup := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, styledPopup)

		helpView := m.help.View(popupKeyMap{})

		return lipgloss.JoinVertical(lipgloss.Left, centeredPopup, helpView)
	}

	return docStyle.Render(m.list.View())
}

// createPopupView creates a styled popup view for host details
func (m Model) createPopupView() string {
	if m.hostDetails == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s\n\n", m.hostDetails.Alias))

	var pairs []struct{ key, value string }
	for key, value := range m.hostDetails.Details {
		pairs = append(pairs, struct{ key, value string }{key, value})
	}
	if _, exists := m.hostDetails.Details["HostName"]; !exists && m.hostDetails.HostName != "" {
		pairs = append(pairs, struct{ key, value string }{"HostName", m.hostDetails.HostName})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	maxKeyLen := 0
	maxValueLen := 0
	for _, pair := range pairs {
		if len(pair.key) > maxKeyLen {
			maxKeyLen = len(pair.key)
		}
		if len(pair.value) > maxValueLen {
			maxValueLen = len(pair.value)
		}
	}

	maxValueLen += 4

	for _, pair := range pairs {
		builder.WriteString(fmt.Sprintf("%-*s    %s\n", maxKeyLen, pair.key, pair.value))
	}

	return builder.String()
}

// openEditor opens the SSH config file in the user's preferred editor.
func (m Model) openEditor() tea.Cmd {
	editor := getEditor()
	if editor == "" {
		return nil
	}
	configPath := getSSHConfigPath()

	return tea.ExecProcess(exec.Command(editor, configPath), func(err error) tea.Msg {
		return nil
	})
}

// showHostDetails displays detailed information about the currently selected host.
func (m *Model) showHostDetails() tea.Cmd {
	if item, ok := m.list.SelectedItem().(HostItem); ok {
		details, err := getHostDetails(item.host)
		if err != nil {
			return nil
		}
		m.viewing = true
		m.hostDetails = details
	}
	return nil
}

// HostDetails contains detailed configuration information for an SSH host.
type HostDetails struct {
	Alias    string
	HostName string
	Details  map[string]string
}

func getHostDetails(host config.Host) (*HostDetails, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config file: %w", err)
	}
	defer file.Close()

	details := make(map[string]string)

	inHost := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.ToLower(parts[0])
			configValue := strings.Join(parts[1:], " ")

			if key == "host" && strings.Contains(configValue, host.Alias) {
				inHost = true
			} else if key == "host" && inHost {
				break // Next host started
			} else if inHost {
				details[capitalizeSSHKey(key)] = configValue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read SSH config file: %w", err)
	}

	return &HostDetails{
		Alias:    host.Alias,
		HostName: host.HostName,
		Details:  details,
	}, nil
}

// getEditor returns the user's preferred editor from the EDITOR environment variable.
// Returns an empty string if EDITOR is not set.
func getEditor() string {
	return os.Getenv("EDITOR")
}

// getSSHConfigPath returns the path to the user's SSH configuration file.
func getSSHConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "config")
}

// RenderMenu displays an interactive menu for selecting SSH hosts and returns
// the selected host alias or "exit" if the user chose to quit.
func RenderMenu(hosts []config.Host) (string, error) {
	program := tea.NewProgram(NewModel(hosts))
	model, err := program.Run()
	if err != nil {
		return "", err
	}
	menuModel := model.(Model)
	return menuModel.choice, nil
}
