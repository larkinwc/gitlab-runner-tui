package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

type RunnersView struct {
	table       table.Model
	runners     []runner.Runner
	service     runner.Service
	width       int
	height      int
	loading     bool
	spinner     spinner.Model
	err         error
	selectedIdx int
}

func NewRunnersView(service runner.Service) *RunnersView {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 12},
		{Title: "Executor", Width: 15},
		{Title: "Tags", Width: 30},
		{Title: "ID", Width: 10},
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

	return &RunnersView{
		table:   t,
		service: service,
		spinner: sp,
		loading: true,
	}
}

func (v *RunnersView) Init() tea.Cmd {
	return tea.Batch(
		v.loadRunners,
		v.spinner.Tick,
	)
}

func (v *RunnersView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.table.SetHeight(v.height - 10)
		return v, nil

	case runnersLoadedMsg:
		v.runners = msg.runners
		v.loading = false
		v.err = msg.err
		v.updateTable()
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			v.loading = true
			return v, v.loadRunners
		case "enter":
			if len(v.runners) > 0 && v.table.Cursor() < len(v.runners) {
				v.selectedIdx = v.table.Cursor()
			}
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

func (v *RunnersView) View() string {
	if v.err != nil {
		return ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", v.err))
	}

	if v.loading {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			HeaderStyle.Render("GitLab Runners"),
			"",
			v.spinner.View()+" Loading runners...",
		)
	}

	content := []string{
		HeaderStyle.Render("GitLab Runners"),
		"",
	}

	if len(v.runners) == 0 {
		content = append(content, InfoBoxStyle.Render("No runners found"))
	} else {
		content = append(content, v.table.View())
	}

	// Help is now shown in the status bar

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (v *RunnersView) updateTable() {
	rows := []table.Row{}
	for i := range v.runners {
		r := &v.runners[i]
		tags := strings.Join(r.TagList, ", ")
		if tags == "" {
			tags = "-"
		}

		status := "unknown"
		if r.Online {
			status = "online"
		} else if r.Status != "" {
			status = r.Status
		}

		rows = append(rows, table.Row{
			TruncateString(r.Name, 30),
			RenderStatus(status),
			r.Executor,
			TruncateString(tags, 30),
			r.ID,
		})
	}
	v.table.SetRows(rows)
}

func (v *RunnersView) loadRunners() tea.Msg {
	runners, err := v.service.ListRunners()
	if err != nil {
		return runnersLoadedMsg{err: err}
	}

	for i := range runners {
		status, err := v.service.GetRunnerStatus(runners[i].Name)
		if err == nil && status != nil {
			runners[i].Status = status.Status
			runners[i].Online = status.Online
		}
	}

	return runnersLoadedMsg{runners: runners}
}

func (v *RunnersView) GetSelectedRunner() *runner.Runner {
	if v.selectedIdx >= 0 && v.selectedIdx < len(v.runners) {
		return &v.runners[v.selectedIdx]
	}
	return nil
}

type runnersLoadedMsg struct {
	runners []runner.Runner
	err     error
}
