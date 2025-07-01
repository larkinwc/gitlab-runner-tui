package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

type LogsView struct {
	viewport   viewport.Model
	service    runner.Service
	runnerName string
	logs       []string
	loading    bool
	spinner    spinner.Model
	err        error
	width      int
	height     int
	autoScroll bool
	filterText string
}

func NewLogsView(service runner.Service) *LogsView {
	vp := viewport.New(80, 20)
	vp.Style = LogStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return &LogsView{
		viewport:   vp,
		service:    service,
		spinner:    sp,
		autoScroll: true,
	}
}

func (v *LogsView) Init() tea.Cmd {
	return tea.Batch(
		v.loadLogs,
		v.spinner.Tick,
	)
}

func (v *LogsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.viewport.Width = v.width - 2
		v.viewport.Height = v.height - 8
		return v, nil

	case logsLoadedMsg:
		v.logs = msg.logs
		v.loading = false
		v.err = msg.err
		v.updateViewport()
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			v.loading = true
			return v, v.loadLogs
		case "a", "A":
			v.autoScroll = !v.autoScroll
			if v.autoScroll {
				v.viewport.GotoBottom()
			}
		case "g":
			v.viewport.GotoTop()
		case "G":
			v.viewport.GotoBottom()
		case "/":
			// TODO: Implement search/filter
		case "c", "C":
			v.logs = []string{}
			v.updateViewport()
		}
	}

	if v.loading {
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *LogsView) View() string {
	if v.err != nil {
		return ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", v.err))
	}

	header := "Logs"
	if v.runnerName != "" {
		header = fmt.Sprintf("Logs - %s", v.runnerName)
	}

	content := []string{
		HeaderStyle.Render(header),
		"",
	}

	if v.loading {
		content = append(content, v.spinner.View()+" Loading logs...")
	} else if len(v.logs) == 0 {
		content = append(content, InfoBoxStyle.Render("No logs available"))
	} else {
		statusBar := fmt.Sprintf("Lines: %d | Position: %d%%", len(v.logs), int(v.viewport.ScrollPercent()*100))
		if v.autoScroll {
			statusBar += " | Auto-scroll: ON"
		}
		content = append(content,
			v.viewport.View(),
			"",
			lipgloss.NewStyle().Foreground(ColorMuted).Render(statusBar),
		)
	}

	// Help is now shown in the status bar

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (v *LogsView) SetRunner(name string) {
	v.runnerName = name
	v.loading = true
}

func (v *LogsView) updateViewport() {
	content := strings.Join(v.logs, "\n")
	v.viewport.SetContent(content)

	if v.autoScroll {
		v.viewport.GotoBottom()
	}
}

func (v *LogsView) loadLogs() tea.Msg {
	logs, err := v.service.GetRunnerLogs(v.runnerName, 1000)
	if err != nil {
		return logsLoadedMsg{err: err}
	}

	return logsLoadedMsg{logs: logs}
}

type logsLoadedMsg struct {
	logs []string
	err  error
}
