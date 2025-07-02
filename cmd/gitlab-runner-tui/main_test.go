package main

import (
	"io"
	"strings"
	"testing"

	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
	"github.com/larkinwc/gitlab-runner-tui/pkg/ui"
)

func TestModelView(t *testing.T) {
	tests := []struct {
		name        string
		model       model
		expectPanic bool
		expectEmpty bool
		contains    []string
	}{
		{
			name: "Quitting returns empty string",
			model: model{
				quitting: true,
			},
			expectEmpty: true,
		},
		{
			name: "Terminal too small shows message",
			model: model{
				height: 5,
				tabs:   []string{"Test"},
			},
			contains: []string{"Terminal too small"},
		},
		{
			name: "Normal view renders correctly",
			model: model{
				height:      30,
				width:       80,
				tabs:        []string{"Runners", "Logs", "Config", "System", "History"},
				activeTab:   0,
				initialized: map[int]bool{0: true},
				runnersView: &ui.RunnersView{},
				logsView:    &ui.LogsView{},
				configView:  &ui.ConfigView{},
				systemView:  &ui.SystemView{},
				historyView: &ui.HistoryView{},
			},
			contains: []string{"1. Runners", "2. Logs", "Tab/Shift+Tab: Switch tabs"},
		},
		{
			name: "Small but valid height doesn't panic",
			model: model{
				height:      10,
				width:       80,
				tabs:        []string{"Test"},
				activeTab:   0,
				initialized: map[int]bool{0: true},
				runnersView: &ui.RunnersView{},
			},
			expectPanic: false,
		},
		{
			name: "Debug mode shows indicator",
			model: model{
				height:      20,
				width:       80,
				tabs:        []string{"Test"},
				activeTab:   0,
				debugMode:   true,
				initialized: map[int]bool{0: true},
				runnersView: &ui.RunnersView{},
			},
			contains: []string{"[DEBUG]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.expectPanic {
					t.Errorf("panic = %v, expectPanic = %v", r != nil, tt.expectPanic)
				}
			}()

			result := tt.model.View()

			if tt.expectEmpty && result != "" {
				t.Errorf("expected empty string, got %q", result)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected output to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestModelSwitchTab(t *testing.T) {
	// Create a minimal model with mocked services
	service := &mockRunnerService{}
	m := model{
		tabs:        []string{"Tab1", "Tab2", "Tab3"},
		activeTab:   0,
		initialized: map[int]bool{0: true},
		runnersView: ui.NewRunnersView(service),
		logsView:    ui.NewLogsView(service),
		configView:  ui.NewConfigView("/tmp/test-config.toml"),
		systemView:  ui.NewSystemView(service),
		historyView: ui.NewHistoryView(service),
	}

	// Test switching to uninitialized tab
	m.activeTab = 1
	newModel, cmd := m.switchTab()

	if !newModel.initialized[1] {
		t.Error("Tab 1 should be marked as initialized")
	}

	if cmd == nil {
		t.Error("Expected init command for uninitialized tab")
	}

	// Test switching to already initialized tab
	m.initialized[2] = true
	m.activeTab = 2
	newModel, cmd = m.switchTab()

	if cmd != nil {
		t.Error("Expected no command for already initialized tab")
	}
}

func TestRenderStatusBar(t *testing.T) {
	tests := []struct {
		name      string
		activeTab int
		debugMode bool
		width     int
		contains  []string
	}{
		{
			name:      "Runners tab commands",
			activeTab: 0,
			width:     100,
			contains:  []string{"↑/↓: Navigate", "Enter: View logs", "r: Refresh"},
		},
		{
			name:      "Logs tab commands",
			activeTab: 1,
			width:     100,
			contains:  []string{"↑/↓: Scroll", "g/G: Top/Bottom", "a: Auto-scroll"},
		},
		{
			name:      "Debug mode indicator",
			activeTab: 0,
			debugMode: true,
			width:     100,
			contains:  []string{"[DEBUG]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				tabs:      []string{"Runners", "Logs", "Config", "System", "History"},
				activeTab: tt.activeTab,
				debugMode: tt.debugMode,
				width:     tt.width,
			}

			result := m.renderStatusBar()

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected status bar to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

// Mock runner service for testing
type mockRunnerService struct{}

func (m *mockRunnerService) ListRunners() ([]runner.Runner, error) {
	return []runner.Runner{}, nil
}

func (m *mockRunnerService) GetRunnerStatus(_ string) (*runner.Runner, error) {
	return nil, nil
}

func (m *mockRunnerService) GetRunnerLogs(_ string, _ int) ([]string, error) {
	return []string{}, nil
}

func (m *mockRunnerService) StreamRunnerLogs(_ string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockRunnerService) RestartRunner() error {
	return nil
}

func (m *mockRunnerService) GetSystemStatus() (*runner.SystemStatus, error) {
	return &runner.SystemStatus{}, nil
}

func (m *mockRunnerService) GetJobHistory(_ int) ([]runner.Job, error) {
	return []runner.Job{}, nil
}

func (m *mockRunnerService) SetDebugMode(_ bool) {}
