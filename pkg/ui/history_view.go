package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

type HistoryView struct {
	table    table.Model
	jobs     []runner.Job
	service  runner.Service
	width    int
	height   int
	loading  bool
	spinner  spinner.Model
	err      error
}

func NewHistoryView(service runner.Service) *HistoryView {
	columns := []table.Column{
		{Title: "Job ID", Width: 10},
		{Title: "Status", Width: 12},
		{Title: "Runner", Width: 20},
		{Title: "Project", Width: 15},
		{Title: "Started", Width: 20},
		{Title: "Duration", Width: 10},
		{Title: "Exit Code", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorSecondary).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(ColorBg).
		Background(ColorPrimary).
		Bold(false)
	t.SetStyles(s)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return &HistoryView{
		table:   t,
		service: service,
		spinner: sp,
		loading: true,
	}
}

func (v *HistoryView) Init() tea.Cmd {
	return tea.Batch(
		v.loadHistory,
		v.spinner.Tick,
		tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
			return historyTickMsg(t)
		}),
	)
}

func (v *HistoryView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.table.SetHeight(v.height - 10)
		return v, nil

	case historyLoadedMsg:
		v.jobs = msg.jobs
		v.loading = false
		v.err = msg.err
		v.updateTable()
		return v, nil

	case historyTickMsg:
		if !v.loading {
			v.loading = true
			return v, v.loadHistory
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			v.loading = true
			return v, v.loadHistory
		}
	}

	if v.loading {
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *HistoryView) View() string {
	if v.err != nil {
		return ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", v.err))
	}

	if v.loading {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			HeaderStyle.Render("Job History"),
			"",
			v.spinner.View()+" Loading job history...",
		)
	}

	content := []string{
		HeaderStyle.Render("Job History"),
		"",
	}

	if len(v.jobs) == 0 {
		content = append(content, InfoBoxStyle.Render("No job history available"))
	} else {
		content = append(content, v.table.View())
		content = append(content, "", 
			lipgloss.NewStyle().Foreground(ColorMuted).Render(
				fmt.Sprintf("Showing %d recent jobs", len(v.jobs)),
			),
		)
	}

	content = append(content, "", HelpStyle.Render("Press 'r' to refresh • 'q' to go back"))

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (v *HistoryView) updateTable() {
	rows := []table.Row{}
	for _, job := range v.jobs {
		status := job.Status
		statusStyle := StatusUnknownStyle
		switch strings.ToLower(status) {
		case "success", "passed", "completed":
			statusStyle = StatusActiveStyle
			status = "✓ " + status
		case "failed", "error":
			statusStyle = StatusInactiveStyle
			status = "✗ " + status
		case "running", "pending":
			statusStyle = lipgloss.NewStyle().Foreground(ColorWarning)
			status = "⟳ " + status
		case "canceled", "skipped":
			statusStyle = lipgloss.NewStyle().Foreground(ColorMuted)
			status = "⊘ " + status
		}

		started := "-"
		if !job.Started.IsZero() {
			started = job.Started.Format("2006-01-02 15:04:05")
		}

		duration := "-"
		if job.Duration > 0 {
			duration = formatJobDuration(job.Duration)
		}

		exitCode := "-"
		if job.ExitCode != 0 {
			exitCode = fmt.Sprintf("%d", job.ExitCode)
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("#%d", job.ID),
			statusStyle.Render(status),
			TruncateString(job.RunnerName, 20),
			TruncateString(job.Project, 15),
			started,
			duration,
			exitCode,
		})
	}
	v.table.SetRows(rows)
}

func (v *HistoryView) loadHistory() tea.Msg {
	jobs, err := v.service.GetJobHistory(50)
	if err != nil {
		return historyLoadedMsg{err: err}
	}
	return historyLoadedMsg{jobs: jobs}
}

func formatJobDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

type historyLoadedMsg struct {
	jobs []runner.Job
	err  error
}

type historyTickMsg time.Time