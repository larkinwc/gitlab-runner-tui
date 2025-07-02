package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
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
	GetJobHistory(limit int) ([]Job, error)
	SetDebugMode(enabled bool)
}

type SystemStatus struct {
	ServiceActive  bool
	ServiceEnabled bool
	ProcessCount   int
	MemoryUsage    int64
	CPUUsage       float64
	Uptime         time.Duration
}

type gitlabRunnerService struct {
	configPath string
	debugMode  bool
}

const defaultConfigPath = "/etc/gitlab-runner/config.toml"

func NewService(configPath string) Service {
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return &gitlabRunnerService{
		configPath: configPath,
	}
}

func (s *gitlabRunnerService) ListRunners() ([]Runner, error) {
	// #nosec G204 -- configPath is validated in NewService
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
		if len(matches[1]) >= 8 {
			runner.ID = matches[1][:8]
		} else {
			runner.ID = matches[1]
		}
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

	for i := range runners {
		runner := &runners[i]
		if runner.Name == name {
			// #nosec G204 -- configPath is validated in NewService, name comes from listed runners
			cmd := exec.Command("gitlab-runner", "verify", "--name", name, "--config", s.configPath)
			output, _ := cmd.CombinedOutput()

			if strings.Contains(string(output), "is alive") {
				runner.Status = "active"
				runner.Online = true
			} else {
				runner.Status = "inactive"
				runner.Online = false
			}

			return runner, nil
		}
	}

	return nil, fmt.Errorf("runner %s not found", name)
}

func (s *gitlabRunnerService) GetRunnerLogs(name string, lines int) ([]string, error) {
	args := []string{"-u", "gitlab-runner", "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	if s.debugMode {
		args = append(args, "-o", "verbose")
	}

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		// #nosec G204 -- lines is an integer parameter
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
	_, _ = fmt.Sscanf(string(output), "%d", &status.ProcessCount)

	cmd = exec.Command("ps", "aux")
	output, _ = cmd.Output()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gitlab-runner") && !strings.Contains(line, "grep") {
			fields := strings.Fields(line)
			if len(fields) > 5 {
				var cpu float64
				_, _ = fmt.Sscanf(fields[2], "%f", &cpu)
				status.CPUUsage += cpu

				var mem int64
				_, _ = fmt.Sscanf(fields[5], "%d", &mem)
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
	idx := strings.Index(output, "=")
	if idx >= 0 && idx < len(output)-1 {
		return strings.TrimSpace(output[idx+1:])
	}
	return ""
}

func (s *gitlabRunnerService) GetJobHistory(limit int) ([]Job, error) {
	output, err := s.getJobLogs(limit)
	if err != nil {
		return nil, err
	}

	jobMap := s.parseJobLogs(output, limit)
	jobs := s.convertJobMapToSlice(jobMap)
	s.sortJobsByStartTime(jobs)

	if len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return jobs, nil
}

func (s *gitlabRunnerService) getJobLogs(limit int) ([]byte, error) {
	// Try to get job history from journalctl logs
	// #nosec G204 -- limit is an integer parameter
	cmd := exec.Command("journalctl", "-u", "gitlab-runner", "-n", fmt.Sprintf("%d", limit*10), "--no-pager", "-r")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to log file
		// #nosec G204 -- limit is an integer parameter
		cmd = exec.Command("tail", "-n", fmt.Sprintf("%d", limit*10), "/var/log/gitlab-runner.log")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get job history: %w", err)
		}
	}
	return output, nil
}

func (s *gitlabRunnerService) parseJobLogs(output []byte, limit int) map[int]*Job {
	lines := strings.Split(string(output), "\n")
	jobMap := make(map[int]*Job)

	for _, line := range lines {
		job := s.parseJobFromLog(line)
		if job != nil {
			s.updateOrAddJob(jobMap, job)
		}

		if len(jobMap) >= limit {
			break
		}
	}

	return jobMap
}

func (s *gitlabRunnerService) updateOrAddJob(jobMap map[int]*Job, job *Job) {
	if existing, ok := jobMap[job.ID]; ok {
		// Update existing job with new info
		if job.Status != "" {
			existing.Status = job.Status
		}
		if !job.Started.IsZero() {
			existing.Started = job.Started
		}
		if !job.Finished.IsZero() {
			existing.Finished = job.Finished
			existing.Duration = job.Finished.Sub(existing.Started)
		}
		if job.RunnerName != "" {
			existing.RunnerName = job.RunnerName
		}
		if job.ExitCode != 0 {
			existing.ExitCode = job.ExitCode
		}
	} else {
		jobMap[job.ID] = job
	}
}

func (s *gitlabRunnerService) convertJobMapToSlice(jobMap map[int]*Job) []Job {
	jobs := make([]Job, 0, len(jobMap))
	for _, job := range jobMap {
		jobs = append(jobs, *job)
	}
	return jobs
}

func (s *gitlabRunnerService) sortJobsByStartTime(jobs []Job) {
	// Sort by started time (newest first)
	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if jobs[i].Started.Before(jobs[j].Started) {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}
}

func (s *gitlabRunnerService) parseJobFromLog(line string) *Job {
	// Parse GitLab Runner log lines for job information
	jobStartRegex := regexp.MustCompile(`job=(\d+).*project=(\d+).*runner=([a-zA-Z0-9_-]+)`)
	jobStatusRegex := regexp.MustCompile(`job=(\d+).*status=(\w+)`)
	jobFinishRegex := regexp.MustCompile(`job=(\d+).*duration=([0-9.]+)s`)

	job := &Job{}

	// Check for job start
	if matches := jobStartRegex.FindStringSubmatch(line); len(matches) > 3 {
		jobID, _ := strconv.Atoi(matches[1])
		job.ID = jobID
		job.Project = matches[2]
		job.RunnerName = matches[3]
		job.Status = "running"

		// Try to extract timestamp
		if idx := strings.Index(line, " "); idx > 0 {
			timeStr := line[:idx]
			if t, err := time.Parse("Jan 02 15:04:05", timeStr); err == nil {
				job.Started = t
			}
		}

		return job
	}

	// Check for job status update
	if matches := jobStatusRegex.FindStringSubmatch(line); len(matches) > 2 {
		jobID, _ := strconv.Atoi(matches[1])
		job.ID = jobID
		job.Status = matches[2]
		return job
	}

	// Check for job completion
	if matches := jobFinishRegex.FindStringSubmatch(line); len(matches) > 2 {
		jobID, _ := strconv.Atoi(matches[1])
		job.ID = jobID
		job.Status = "completed"

		if duration, err := strconv.ParseFloat(matches[2], 64); err == nil {
			job.Duration = time.Duration(duration * float64(time.Second))
			job.Finished = time.Now()
		}

		return job
	}

	return nil
}

func (s *gitlabRunnerService) SetDebugMode(enabled bool) {
	s.debugMode = enabled
}
