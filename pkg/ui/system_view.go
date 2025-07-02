package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

type SystemView struct {
	service      runner.Service
	systemStatus *runner.SystemStatus
	loading      bool
	spinner      spinner.Model
	cpuProgress  progress.Model
	memProgress  progress.Model
	err          error
	width        int
	height       int
}

func NewSystemView(service runner.Service) *SystemView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	cpuProg := progress.New(progress.WithDefaultGradient())
	memProg := progress.New(progress.WithDefaultGradient())

	return &SystemView{
		service:     service,
		spinner:     sp,
		cpuProgress: cpuProg,
		memProgress: memProg,
		loading:     true,
	}
}

func (v *SystemView) Init() tea.Cmd {
	return tea.Batch(
		v.loadSystemStatus,
		v.spinner.Tick,
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (v *SystemView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.cpuProgress.Width = v.width - 30
		v.memProgress.Width = v.width - 30
		return v, nil

	case systemStatusLoadedMsg:
		v.systemStatus = msg.status
		v.loading = false
		v.err = msg.err
		return v, nil

	case tickMsg:
		return v, v.loadSystemStatus

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			v.loading = true
			return v, v.loadSystemStatus
		case "s", "S":
			if v.systemStatus != nil {
				return v, v.restartService
			}
		}
	}

	if v.loading {
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *SystemView) View() string {
	content := []string{
		HeaderStyle.Render("System Status"),
		"",
	}

	if v.err != nil {
		content = append(content, ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", v.err)))
	} else if v.loading {
		content = append(content, v.spinner.View()+" Loading system status...")
	} else if v.systemStatus != nil {
		status := v.systemStatus

		serviceStatus := "Inactive"
		serviceStyle := StatusInactiveStyle
		if status.ServiceActive {
			serviceStatus = "Active"
			serviceStyle = StatusActiveStyle
		}

		enabledStatus := "Disabled"
		enabledStyle := StatusInactiveStyle
		if status.ServiceEnabled {
			enabledStatus = "Enabled"
			enabledStyle = StatusActiveStyle
		}

		infoItems := []string{
			fmt.Sprintf("Service Status: %s", serviceStyle.Render(serviceStatus)),
			fmt.Sprintf("Service Enabled: %s", enabledStyle.Render(enabledStatus)),
			fmt.Sprintf("Process Count: %d", status.ProcessCount),
			fmt.Sprintf("Uptime: %s", formatDuration(status.Uptime)),
		}

		content = append(content, InfoBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, infoItems...)), "")

		cpuPercent := status.CPUUsage / 100.0
		if cpuPercent > 1.0 {
			cpuPercent = 1.0
		}
		content = append(content,
			fmt.Sprintf("CPU Usage: %.1f%%", status.CPUUsage),
			v.cpuProgress.ViewAs(cpuPercent),
			"")

		memoryMB := float64(status.MemoryUsage) / 1024 / 1024
		memPercent := float64(status.MemoryUsage) / (8 * 1024 * 1024 * 1024)
		if memPercent > 1.0 {
			memPercent = 1.0
		}
		content = append(content,
			fmt.Sprintf("Memory Usage: %.1f MB", memoryMB),
			v.memProgress.ViewAs(memPercent))
	}

	// Help is now shown in the status bar

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (v *SystemView) loadSystemStatus() tea.Msg {
	status, err := v.service.GetSystemStatus()
	if err != nil {
		return systemStatusLoadedMsg{err: err}
	}
	return systemStatusLoadedMsg{status: status}
}

func (v *SystemView) restartService() tea.Msg {
	if err := v.service.RestartRunner(); err != nil {
		return systemStatusLoadedMsg{err: err}
	}
	time.Sleep(2 * time.Second)
	return v.loadSystemStatus()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

type systemStatusLoadedMsg struct {
	status *runner.SystemStatus
	err    error
}

type tickMsg time.Time
