package config

import (
	"os"
	"strings"
	"testing"

	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

func TestTOMLConfigManager_LoadAndSave(t *testing.T) {
	testConfig := `concurrent = 4
check_interval = 3
log_level = "info"

[[runners]]
  name = "test-runner"
  url = "https://gitlab.example.com"
  token = "test-token"
  executor = "docker"
  
  [runners.docker]
    image = "alpine:latest"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cm := NewTOMLConfigManager(tmpFile.Name())

	if err := cm.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm.GetConfig()
	if config == nil {
		t.Fatal("Config is nil after loading")
	}

	if config.Concurrent != 4 {
		t.Errorf("Expected concurrent=4, got %d", config.Concurrent)
	}

	if len(config.Runners) != 1 {
		t.Fatalf("Expected 1 runner, got %d", len(config.Runners))
	}

	if config.Runners[0].Name != "test-runner" {
		t.Errorf("Expected runner name 'test-runner', got '%s'", config.Runners[0].Name)
	}

	if err := cm.UpdateConcurrency(8); err != nil {
		t.Errorf("Failed to update concurrency: %v", err)
	}

	if err := cm.Save(); err != nil {
		t.Errorf("Failed to save config: %v", err)
	}

	cm2 := NewTOMLConfigManager(tmpFile.Name())
	if err := cm2.Load(); err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if cm2.GetConfig().Concurrent != 8 {
		t.Errorf("Expected concurrent=8 after save/reload, got %d", cm2.GetConfig().Concurrent)
	}
}

func TestTOMLConfigManager_Updates(t *testing.T) {
	cm := &TOMLConfigManager{}
	cm.config = &runner.Config{
		Concurrent: 1,
		Runners: []runner.RunnerConfig{
			{
				Name:     "test-runner",
				URL:      "https://gitlab.example.com",
				Token:    "test-token",
				Executor: "docker",
				Docker:   &runner.DockerConfig{Image: "alpine:latest"},
			},
		},
	}

	tests := []struct {
		name        string
		testFunc    func() error
		validate    func() bool
		expectError bool
	}{
		{
			name: "Update check interval",
			testFunc: func() error {
				return cm.UpdateCheckInterval(10)
			},
			validate: func() bool {
				return cm.config.CheckInterval == 10
			},
		},
		{
			name: "Update check interval negative",
			testFunc: func() error {
				return cm.UpdateCheckInterval(-1)
			},
			expectError: true,
		},
		{
			name: "Update log level valid",
			testFunc: func() error {
				return cm.UpdateLogLevel("DEBUG")
			},
			validate: func() bool {
				return cm.config.LogLevel == "debug"
			},
		},
		{
			name: "Update log level invalid",
			testFunc: func() error {
				return cm.UpdateLogLevel("invalid")
			},
			expectError: true,
		},
		{
			name: "Update runner limit",
			testFunc: func() error {
				return cm.UpdateRunnerLimit("test-runner", 5)
			},
			validate: func() bool {
				return cm.config.Runners[0].Limit == 5
			},
		},
		{
			name: "Update runner limit negative",
			testFunc: func() error {
				return cm.UpdateRunnerLimit("test-runner", -1)
			},
			expectError: true,
		},
		{
			name: "Update non-existent runner",
			testFunc: func() error {
				return cm.UpdateRunnerLimit("non-existent", 5)
			},
			expectError: true,
		},
		{
			name: "Update runner tags",
			testFunc: func() error {
				return cm.UpdateRunnerTags("test-runner", []string{"tag1", "tag2"})
			},
			validate: func() bool {
				tags := cm.config.Runners[0].TagList
				return len(tags) == 2 && tags[0] == "tag1" && tags[1] == "tag2"
			},
		},
		{
			name: "Update runner untagged",
			testFunc: func() error {
				return cm.UpdateRunnerUntagged("test-runner", true)
			},
			validate: func() bool {
				return cm.config.Runners[0].RunUntagged == true
			},
		},
		{
			name: "Update runner locked",
			testFunc: func() error {
				return cm.UpdateRunnerLocked("test-runner", true)
			},
			validate: func() bool {
				return cm.config.Runners[0].Locked == true
			},
		},
		{
			name: "Update runner max builds",
			testFunc: func() error {
				return cm.UpdateRunnerMaxBuilds("test-runner", 10)
			},
			validate: func() bool {
				return cm.config.Runners[0].MaxBuilds == 10
			},
		},
		{
			name: "Update runner request concurrency",
			testFunc: func() error {
				return cm.UpdateRunnerRequestConcurrency("test-runner", 5)
			},
			validate: func() bool {
				return cm.config.Runners[0].RequestConcurrency == 5
			},
		},
		{
			name: "Update runner output limit",
			testFunc: func() error {
				return cm.UpdateRunnerOutputLimit("test-runner", 4096)
			},
			validate: func() bool {
				return cm.config.Runners[0].OutputLimit == 4096
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil && !tt.validate() {
					t.Error("validation failed")
				}
			}
		})
	}
}

func TestTOMLConfigManager_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "no config loaded",
		},
		{
			name: "Invalid concurrent",
			config: &runner.Config{
				Concurrent: 0,
			},
			expectError: true,
			errorMsg:    "concurrent must be at least 1",
		},
		{
			name: "Runner without name",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{URL: "https://gitlab.com", Token: "token", Executor: "docker"},
				},
			},
			expectError: true,
			errorMsg:    "has no name",
		},
		{
			name: "Runner without URL",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{Name: "test", Token: "token", Executor: "docker"},
				},
			},
			expectError: true,
			errorMsg:    "has no URL",
		},
		{
			name: "Runner without token",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{Name: "test", URL: "https://gitlab.com", Executor: "docker"},
				},
			},
			expectError: true,
			errorMsg:    "has no token",
		},
		{
			name: "Runner without executor",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{Name: "test", URL: "https://gitlab.com", Token: "token"},
				},
			},
			expectError: true,
			errorMsg:    "has no executor",
		},
		{
			name: "Docker executor without image",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{
						Name:     "test",
						URL:      "https://gitlab.com",
						Token:    "token",
						Executor: "docker",
						Docker:   &runner.DockerConfig{},
					},
				},
			},
			expectError: true,
			errorMsg:    "docker executor requires image",
		},
		{
			name: "Kubernetes executor without image",
			config: &runner.Config{
				Concurrent: 1,
				Runners: []runner.RunnerConfig{
					{
						Name:       "test",
						URL:        "https://gitlab.com",
						Token:      "token",
						Executor:   "kubernetes",
						Kubernetes: &runner.KubernetesConfig{},
					},
				},
			},
			expectError: true,
			errorMsg:    "kubernetes executor requires image",
		},
		{
			name: "Valid config",
			config: &runner.Config{
				Concurrent: 4,
				Runners: []runner.RunnerConfig{
					{
						Name:     "test",
						URL:      "https://gitlab.com",
						Token:    "token",
						Executor: "shell",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &TOMLConfigManager{config: tt.config}
			err := cm.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTOMLConfigManager_NoConfigLoaded(t *testing.T) {
	cm := &TOMLConfigManager{}

	// Test all methods that require config to be loaded
	if err := cm.UpdateConcurrency(5); err == nil {
		t.Error("expected error when no config loaded")
	}

	if err := cm.UpdateCheckInterval(5); err == nil {
		t.Error("expected error when no config loaded")
	}

	if err := cm.UpdateLogLevel("info"); err == nil {
		t.Error("expected error when no config loaded")
	}

	if err := cm.Save(); err == nil {
		t.Error("expected error when no config loaded")
	}

	if runner, idx := cm.GetRunner("test"); runner != nil || idx != -1 {
		t.Error("expected nil runner and -1 index when no config loaded")
	}
}

func TestNewTOMLConfigManager(t *testing.T) {
	// Test with custom path
	cm := NewTOMLConfigManager("/custom/path.toml")
	if cm.path != "/custom/path.toml" {
		t.Errorf("path = %q, expected %q", cm.path, "/custom/path.toml")
	}

	// Test with empty path (should use default)
	cm = NewTOMLConfigManager("")
	if cm.path != DefaultConfigPath {
		t.Errorf("path = %q, expected default path %q", cm.path, DefaultConfigPath)
	}
}
