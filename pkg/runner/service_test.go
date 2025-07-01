package runner

import (
	"strings"
	"testing"
	"time"
)

func TestParseRunnerLine(t *testing.T) {
	s := &gitlabRunnerService{}

	tests := []struct {
		name     string
		line     string
		expected *Runner
	}{
		{
			name: "Parse valid runner line",
			line: "Name=test-runner Token=abc123def Executor=docker",
			expected: &Runner{
				Name:     "test-runner",
				Token:    "abc123def",
				ID:       "abc123de",
				Executor: "docker",
				Status:   "unknown",
			},
		},
		{
			name: "Parse runner with spaces in name",
			line: "Name=my test runner Token=xyz789 Executor=shell",
			expected: &Runner{
				Name:     "my test runner",
				Token:    "xyz789",
				ID:       "xyz789",
				Executor: "shell",
				Status:   "unknown",
			},
		},
		{
			name:     "Invalid line returns nil",
			line:     "Invalid format",
			expected: nil,
		},
		{
			name:     "Empty line returns nil",
			line:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.parseRunnerLine(tt.line)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			
			if result == nil {
				t.Error("expected runner, got nil")
				return
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %q, expected %q", result.Name, tt.expected.Name)
			}
			if result.Token != tt.expected.Token {
				t.Errorf("Token = %q, expected %q", result.Token, tt.expected.Token)
			}
			if result.ID != tt.expected.ID {
				t.Errorf("ID = %q, expected %q", result.ID, tt.expected.ID)
			}
			if result.Executor != tt.expected.Executor {
				t.Errorf("Executor = %q, expected %q", result.Executor, tt.expected.Executor)
			}
		})
	}
}

func TestParseJobFromLog(t *testing.T) {
	s := &gitlabRunnerService{}

	tests := []struct {
		name     string
		line     string
		hasJob   bool
		validate func(*testing.T, *Job)
	}{
		{
			name:   "Parse job start",
			line:   "Jan 01 12:34:56 job=12345 project=67890 runner=test-runner",
			hasJob: true,
			validate: func(t *testing.T, job *Job) {
				if job.ID != 12345 {
					t.Errorf("ID = %d, expected 12345", job.ID)
				}
				if job.Project != "67890" {
					t.Errorf("Project = %q, expected %q", job.Project, "67890")
				}
				if job.RunnerName != "test-runner" {
					t.Errorf("RunnerName = %q, expected %q", job.RunnerName, "test-runner")
				}
				if job.Status != "running" {
					t.Errorf("Status = %q, expected %q", job.Status, "running")
				}
			},
		},
		{
			name:   "Parse job status update",
			line:   "job=12345 status=success",
			hasJob: true,
			validate: func(t *testing.T, job *Job) {
				if job.ID != 12345 {
					t.Errorf("ID = %d, expected 12345", job.ID)
				}
				if job.Status != "success" {
					t.Errorf("Status = %q, expected %q", job.Status, "success")
				}
			},
		},
		{
			name:   "Parse job completion",
			line:   "job=12345 duration=123.45s",
			hasJob: true,
			validate: func(t *testing.T, job *Job) {
				if job.ID != 12345 {
					t.Errorf("ID = %d, expected 12345", job.ID)
				}
				if job.Status != "completed" {
					t.Errorf("Status = %q, expected %q", job.Status, "completed")
				}
				expectedDuration := time.Duration(123.45 * float64(time.Second))
				if job.Duration != expectedDuration {
					t.Errorf("Duration = %v, expected %v", job.Duration, expectedDuration)
				}
			},
		},
		{
			name:   "Invalid log line",
			line:   "Random log message without job info",
			hasJob: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.parseJobFromLog(tt.line)
			
			if !tt.hasJob {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			
			if result == nil {
				t.Error("expected job, got nil")
				return
			}
			
			tt.validate(t, result)
		})
	}
}

func TestExtractTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid timestamp format",
			input:    "ActiveEnterTimestamp=Mon 2024-01-01 12:34:56 UTC",
			expected: "Mon 2024-01-01 12:34:56 UTC",
		},
		{
			name:     "No equals sign",
			input:    "Invalid format",
			expected: "",
		},
		{
			name:     "Empty value after equals",
			input:    "ActiveEnterTimestamp=",
			expected: "",
		},
		{
			name:     "Multiple equals signs",
			input:    "Key=Value=Extra",
			expected: "Value=Extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("extractTimestamp(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseRunnerList(t *testing.T) {
	s := &gitlabRunnerService{}

	input := `Listing configured runners...
Name=runner1 Token=token1 Executor=docker
Name=runner2 Token=token2 Executor=shell
Invalid line should be ignored
Name=runner3 Token=token3 Executor=kubernetes
`

	runners, err := s.parseRunnerList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runners) != 3 {
		t.Fatalf("expected 3 runners, got %d", len(runners))
	}

	expectedNames := []string{"runner1", "runner2", "runner3"}
	for i, runner := range runners {
		if runner.Name != expectedNames[i] {
			t.Errorf("runner[%d].Name = %q, expected %q", i, runner.Name, expectedNames[i])
		}
	}
}

func TestSetDebugMode(t *testing.T) {
	s := &gitlabRunnerService{}
	
	// Initially should be false
	if s.debugMode {
		t.Error("debugMode should be false initially")
	}
	
	// Set to true
	s.SetDebugMode(true)
	if !s.debugMode {
		t.Error("debugMode should be true after setting")
	}
	
	// Set to false
	s.SetDebugMode(false)
	if s.debugMode {
		t.Error("debugMode should be false after unsetting")
	}
}

func TestNewService(t *testing.T) {
	// Test with custom path
	customPath := "/custom/path/config.toml"
	service := NewService(customPath)
	if s, ok := service.(*gitlabRunnerService); ok {
		if s.configPath != customPath {
			t.Errorf("configPath = %q, expected %q", s.configPath, customPath)
		}
	} else {
		t.Error("NewService did not return *gitlabRunnerService")
	}
	
	// Test with empty path (should use default)
	service = NewService("")
	if s, ok := service.(*gitlabRunnerService); ok {
		if s.configPath != "/etc/gitlab-runner/config.toml" {
			t.Errorf("configPath = %q, expected default path", s.configPath)
		}
	}
}

func TestGetRunnerLogs_DebugMode(t *testing.T) {
	s := &gitlabRunnerService{
		debugMode: true,
	}
	
	// This test would need to mock exec.Command, which is complex
	// For now, we just ensure the method handles debug mode without panicking
	_, err := s.GetRunnerLogs("test-runner", 10)
	// We expect an error because the command won't exist in test environment
	if err == nil {
		t.Skip("Skipping test that requires journalctl")
	}
	
	// The important thing is that it didn't panic
	if !strings.Contains(err.Error(), "failed to get logs") {
		t.Errorf("unexpected error: %v", err)
	}
}