package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
	"github.com/larkinwc/gitlab-runner-tui/pkg/ui"
)

type model struct {
	tabs         []string
	activeTab    int
	runnersView  *ui.RunnersView
	logsView     *ui.LogsView
	configView   *ui.ConfigView
	systemView   *ui.SystemView
	historyView  *ui.HistoryView
	width        int
	height       int
	quitting     bool
	debugMode    bool
	initialized  map[int]bool
}

func initialModel(configPath string, debugMode bool) model {
	service := runner.NewService(configPath)
	service.SetDebugMode(debugMode)

	m := model{
		tabs:        []string{"Runners", "Logs", "Config", "System", "History"},
		activeTab:   0,
		runnersView: ui.NewRunnersView(service),
		logsView:    ui.NewLogsView(service),
		configView:  ui.NewConfigView(configPath),
		systemView:  ui.NewSystemView(service),
		historyView: ui.NewHistoryView(service),
		debugMode:   debugMode,
		initialized: make(map[int]bool),
	}
	m.initialized[0] = true // Mark first tab as initialized
	return m
}

func (m model) Init() tea.Cmd {
	// Only initialize the first view
	return m.runnersView.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.runnersView.Update(msg)
		m.logsView.Update(msg)
		m.configView.Update(msg)
		m.systemView.Update(msg)
		m.historyView.Update(msg)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.activeTab == 1 {
				m.activeTab = 0
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m.switchTab()

		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			return m.switchTab()

		case "1", "2", "3", "4", "5":
			if idx := int(msg.String()[0] - '1'); idx < len(m.tabs) {
				m.activeTab = idx
				return m.switchTab()
			}
			return m, nil

		case "enter":
			if m.activeTab == 0 {
				if runner := m.runnersView.GetSelectedRunner(); runner != nil {
					m.logsView.SetRunner(runner.Name)
					m.activeTab = 1
					return m.switchTab()
				}
			}
		}
	}

	switch m.activeTab {
	case 0:
		updatedView, cmd := m.runnersView.Update(msg)
		m.runnersView = updatedView.(*ui.RunnersView)
		cmds = append(cmds, cmd)
	case 1:
		updatedView, cmd := m.logsView.Update(msg)
		m.logsView = updatedView.(*ui.LogsView)
		cmds = append(cmds, cmd)
	case 2:
		updatedView, cmd := m.configView.Update(msg)
		m.configView = updatedView.(*ui.ConfigView)
		cmds = append(cmds, cmd)
	case 3:
		updatedView, cmd := m.systemView.Update(msg)
		m.systemView = updatedView.(*ui.SystemView)
		cmds = append(cmds, cmd)
	case 4:
		updatedView, cmd := m.historyView.Update(msg)
		m.historyView = updatedView.(*ui.HistoryView)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) switchTab() (model, tea.Cmd) {
	if !m.initialized[m.activeTab] {
		m.initialized[m.activeTab] = true
		switch m.activeTab {
		case 1:
			return m, m.logsView.Init()
		case 2:
			return m, m.configView.Init()
		case 3:
			return m, m.systemView.Init()
		case 4:
			return m, m.historyView.Init()
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	tabBar := m.renderTabBar()

	var content string
	switch m.activeTab {
	case 0:
		content = m.runnersView.View()
	case 1:
		content = m.logsView.View()
	case 2:
		content = m.configView.View()
	case 3:
		content = m.systemView.View()
	case 4:
		content = m.historyView.View()
	}

	statusBar := m.renderStatusBar()
	
	// Calculate available height for content
	availableHeight := m.height - lipgloss.Height(tabBar) - lipgloss.Height(statusBar) - 1
	
	// Ensure content doesn't overflow
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > availableHeight {
		content = strings.Join(contentLines[:availableHeight], "\n")
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		strings.Repeat("\n", m.height-lipgloss.Height(tabBar)-lipgloss.Height(content)-lipgloss.Height(statusBar)),
		statusBar,
	)
}

func (m model) renderTabBar() string {
	var tabs []string

	for i, tab := range m.tabs {
		style := ui.TabStyle
		if i == m.activeTab {
			style = ui.ActiveTabStyle
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d. %s", i+1, tab)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m model) renderStatusBar() string {
	// Create styles for the status bar
	statusStyle := lipgloss.NewStyle().
		Background(ui.ColorPrimary).
		Foreground(ui.ColorBg).
		Padding(0, 1)
		
	helpStyle := lipgloss.NewStyle().
		Background(ui.ColorSecondary).
		Foreground(ui.ColorBg).
		Padding(0, 1)
		
	// Build contextual help based on active tab
	var commands []string
	
	// Global commands
	commands = append(commands, "Tab/Shift+Tab: Switch tabs", "1-5: Jump to tab", "q: Quit")
	
	// Tab-specific commands
	switch m.activeTab {
	case 0: // Runners
		commands = append(commands, "↑/↓: Navigate", "Enter: View logs", "r: Refresh")
	case 1: // Logs
		commands = append(commands, "↑/↓: Scroll", "g/G: Top/Bottom", "a: Auto-scroll", "c: Clear", "r: Refresh")
	case 2: // Config
		commands = append(commands, "Tab: Next field", "Ctrl+S: Save", "r: Edit runners")
	case 3: // System
		commands = append(commands, "r: Refresh", "s: Restart service")
	case 4: // History
		commands = append(commands, "↑/↓: Navigate", "r: Refresh")
	}
	
	// Add debug mode indicator if enabled
	statusText := ""
	if m.debugMode {
		statusText = " [DEBUG] "
	}
	
	// Combine status and help
	status := statusStyle.Render(statusText + m.tabs[m.activeTab])
	help := helpStyle.Render(strings.Join(commands, " • "))
	
	// Fill the remaining width
	statusBarContent := lipgloss.JoinHorizontal(lipgloss.Top, status, " ", help)
	remainingWidth := m.width - lipgloss.Width(statusBarContent)
	if remainingWidth > 0 {
		filler := lipgloss.NewStyle().
			Background(ui.ColorSecondary).
			Render(strings.Repeat(" ", remainingWidth))
		statusBarContent = lipgloss.JoinHorizontal(lipgloss.Top, statusBarContent, filler)
	}
	
	return statusBarContent
}

func main() {
	var configPath string
	var debugMode bool
	var showHelp bool
	
	defaultConfig := "/etc/gitlab-runner/config.toml"
	
	flag.StringVar(&configPath, "config", defaultConfig, "Path to GitLab Runner config file")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode for verbose logging")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.BoolVar(&showHelp, "h", false, "Show help information")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GitLab Runner TUI - Terminal User Interface for managing GitLab runners\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDefault config paths checked:\n")
		fmt.Fprintf(os.Stderr, "  1. %s (system-wide)\n", defaultConfig)
		fmt.Fprintf(os.Stderr, "  2. $HOME/.gitlab-runner/config.toml (user-specific)\n")
		fmt.Fprintf(os.Stderr, "\nIf no config is found at the default path, the user-specific path is tried.\n")
	}
	
	flag.Parse()
	
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Check if config exists at specified path
	if configPath == defaultConfig {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			altPath := os.ExpandEnv("$HOME/.gitlab-runner/config.toml")
			if _, err := os.Stat(altPath); err == nil {
				configPath = altPath
				if debugMode {
					fmt.Printf("Using config from: %s\n", configPath)
				}
			}
		}
	}

	p := tea.NewProgram(
		initialModel(configPath, debugMode),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
