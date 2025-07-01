package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Service interface {
	ListRunners() ([]Runner, error)
	GetRunnerStatus(name string) (*Runner, error)
	GetRunnerLogs(name string, lines int) ([]string, error)
	StreamRunnerLogs(name string) (io.ReadCloser, error)
	RestartRunner() error
	GetSystemStatus() (*SystemStatus, error)
}

type SystemStatus struct {
	ServiceActive   bool
	ServiceEnabled  bool
	ProcessCount    int
	MemoryUsage     int64
	CPUUsage        float64
	Uptime          time.Duration
}

type gitlabRunnerService struct {
	configPath string
}

func NewService(configPath string) Service {
	if configPath == "" {
		configPath = "/etc/gitlab-runner/config.toml"
	}
	return &gitlabRunnerService{
		configPath: configPath,
	}
}

func (s *gitlabRunnerService) ListRunners() ([]Runner, error) {
	cmd := exec.Command("gitlab-runner", "list", "--config", s.configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list runners: %w", err)
	}

	return s.parseRunnerList(string(output))
}

func (s *gitlabRunnerService) parseRunnerList(output string) ([]Runner, error) {
	var runners []Runner
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}
		
		runner := s.parseRunnerLine(line)
		if runner != nil {
			runners = append(runners, *runner)
		}
	}
	
	return runners, nil
}

func (s *gitlabRunnerService) parseRunnerLine(line string) *Runner {
	tokenRegex := regexp.MustCompile(`Token\s*=\s*(\S+)`)
	nameRegex := regexp.MustCompile(`Name\s*=\s*(.+?)(?:\s+Token|$)`)
	executorRegex := regexp.MustCompile(`Executor\s*=\s*(\S+)`)

	runner := &Runner{
		Status: "unknown",
	}

	if matches := tokenRegex.FindStringSubmatch(line); len(matches) > 1 {
		runner.Token = matches[1]
		runner.ID = matches[1][:8]
	}

	if matches := nameRegex.FindStringSubmatch(line); len(matches) > 1 {
		runner.Name = strings.TrimSpace(matches[1])
	}

	if matches := executorRegex.FindStringSubmatch(line); len(matches) > 1 {
		runner.Executor = matches[1]
	}

	if runner.Name == "" && runner.Token == "" {
		return nil
	}

	return runner
}

func (s *gitlabRunnerService) GetRunnerStatus(name string) (*Runner, error) {
	runners, err := s.ListRunners()
	if err != nil {
		return nil, err
	}

	for _, runner := range runners {
		if runner.Name == name {
			cmd := exec.Command("gitlab-runner", "verify", "--name", name, "--config", s.configPath)
			output, _ := cmd.CombinedOutput()
			
			if strings.Contains(string(output), "is alive") {
				runner.Status = "active"
				runner.Online = true
			} else {
				runner.Status = "inactive"
				runner.Online = false
			}
			
			return &runner, nil
		}
	}

	return nil, fmt.Errorf("runner %s not found", name)
}

func (s *gitlabRunnerService) GetRunnerLogs(name string, lines int) ([]string, error) {
	cmd := exec.Command("journalctl", "-u", "gitlab-runner", "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("tail", "-n", fmt.Sprintf("%d", lines), "/var/log/gitlab-runner.log")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get logs: %w", err)
		}
	}

	logLines := strings.Split(string(output), "\n")
	var filteredLogs []string
	
	for _, line := range logLines {
		if name == "" || strings.Contains(line, name) {
			filteredLogs = append(filteredLogs, line)
		}
	}

	return filteredLogs, nil
}

func (s *gitlabRunnerService) StreamRunnerLogs(name string) (io.ReadCloser, error) {
	cmd := exec.Command("journalctl", "-u", "gitlab-runner", "-f", "--no-pager")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start log streaming: %w", err)
	}

	if name == "" {
		return stdout, nil
	}

	pr, pw := io.Pipe()
	
	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, name) {
				fmt.Fprintln(pw, line)
			}
		}
	}()

	return pr, nil
}

func (s *gitlabRunnerService) RestartRunner() error {
	cmd := exec.Command("systemctl", "restart", "gitlab-runner")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("service", "gitlab-runner", "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart gitlab-runner service: %w", err)
		}
	}
	return nil
}

func (s *gitlabRunnerService) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{}

	cmd := exec.Command("systemctl", "is-active", "gitlab-runner")
	output, _ := cmd.Output()
	status.ServiceActive = strings.TrimSpace(string(output)) == "active"

	cmd = exec.Command("systemctl", "is-enabled", "gitlab-runner")
	output, _ = cmd.Output()
	status.ServiceEnabled = strings.TrimSpace(string(output)) == "enabled"

	cmd = exec.Command("pgrep", "-c", "gitlab-runner")
	output, _ = cmd.Output()
	fmt.Sscanf(string(output), "%d", &status.ProcessCount)

	cmd = exec.Command("ps", "aux")
	output, _ = cmd.Output()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gitlab-runner") && !strings.Contains(line, "grep") {
			fields := strings.Fields(line)
			if len(fields) > 5 {
				var cpu float64
				fmt.Sscanf(fields[2], "%f", &cpu)
				status.CPUUsage += cpu
				
				var mem int64
				fmt.Sscanf(fields[5], "%d", &mem)
				status.MemoryUsage += mem * 1024
			}
		}
	}

	cmd = exec.Command("systemctl", "show", "gitlab-runner", "--property=ActiveEnterTimestamp")
	output, _ = cmd.Output()
	if timestamp := extractTimestamp(string(output)); timestamp != "" {
		if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", timestamp); err == nil {
			status.Uptime = time.Since(t)
		}
	}

	return status, nil
}

func extractTimestamp(output string) string {
	parts := strings.Split(output, "=")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}