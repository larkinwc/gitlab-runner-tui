package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/larkin/gitlab-runner-tui/pkg/runner"
)

type TOMLConfigManager struct {
	path   string
	config *runner.Config
}

func NewTOMLConfigManager(path string) *TOMLConfigManager {
	if path == "" {
		path = "/etc/gitlab-runner/config.toml"
	}
	return &TOMLConfigManager{
		path: path,
	}
}

func (cm *TOMLConfigManager) Load() error {
	config := &runner.Config{}
	
	if _, err := toml.DecodeFile(cm.path, config); err != nil {
		return fmt.Errorf("failed to parse TOML config: %w", err)
	}

	cm.config = config
	return nil
}

func (cm *TOMLConfigManager) Save() error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	encoder.Indent = ""
	
	if err := encoder.Encode(cm.config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	tmpFile := cm.path + ".tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	backupFile := cm.path + ".bak"
	if _, err := os.Stat(cm.path); err == nil {
		if err := os.Rename(cm.path, backupFile); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("failed to backup config file: %w", err)
		}
	}

	if err := os.Rename(tmpFile, cm.path); err != nil {
		os.Rename(backupFile, cm.path)
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

func (cm *TOMLConfigManager) GetConfig() *runner.Config {
	return cm.config
}

func (cm *TOMLConfigManager) UpdateConcurrency(concurrent int) error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}
	
	if concurrent < 1 {
		return fmt.Errorf("concurrent must be at least 1")
	}
	
	cm.config.Concurrent = concurrent
	return nil
}

func (cm *TOMLConfigManager) UpdateCheckInterval(interval int) error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}
	
	if interval < 0 {
		return fmt.Errorf("check_interval must be non-negative")
	}
	
	cm.config.CheckInterval = interval
	return nil
}

func (cm *TOMLConfigManager) UpdateLogLevel(level string) error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}
	
	level = strings.ToLower(level)
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}
	
	if !validLevels[level] {
		return fmt.Errorf("invalid log level: %s", level)
	}
	
	cm.config.LogLevel = level
	return nil
}

func (cm *TOMLConfigManager) GetRunner(name string) (*runner.RunnerConfig, int) {
	if cm.config == nil {
		return nil, -1
	}
	
	for i, r := range cm.config.Runners {
		if r.Name == name {
			return &cm.config.Runners[i], i
		}
	}
	
	return nil, -1
}

func (cm *TOMLConfigManager) UpdateRunnerLimit(name string, limit int) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	
	cm.config.Runners[idx].Limit = limit
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerTags(name string, tags []string) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	cm.config.Runners[idx].TagList = tags
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerUntagged(name string, runUntagged bool) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	cm.config.Runners[idx].RunUntagged = runUntagged
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerLocked(name string, locked bool) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	cm.config.Runners[idx].Locked = locked
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerMaxBuilds(name string, maxBuilds int) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	if maxBuilds < 0 {
		return fmt.Errorf("max_builds must be non-negative")
	}
	
	cm.config.Runners[idx].MaxBuilds = maxBuilds
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerRequestConcurrency(name string, concurrency int) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	if concurrency < 0 {
		return fmt.Errorf("request_concurrency must be non-negative")
	}
	
	cm.config.Runners[idx].Request_concurrency = concurrency
	return nil
}

func (cm *TOMLConfigManager) UpdateRunnerOutputLimit(name string, limit int) error {
	runner, idx := cm.GetRunner(name)
	if runner == nil {
		return fmt.Errorf("runner %s not found", name)
	}
	
	if limit < 0 {
		return fmt.Errorf("output_limit must be non-negative")
	}
	
	cm.config.Runners[idx].Output_limit = limit
	return nil
}

func (cm *TOMLConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}
	
	if cm.config.Concurrent < 1 {
		return fmt.Errorf("concurrent must be at least 1")
	}
	
	for i, runner := range cm.config.Runners {
		if runner.Name == "" {
			return fmt.Errorf("runner %d has no name", i)
		}
		if runner.URL == "" {
			return fmt.Errorf("runner %s has no URL", runner.Name)
		}
		if runner.Token == "" {
			return fmt.Errorf("runner %s has no token", runner.Name)
		}
		if runner.Executor == "" {
			return fmt.Errorf("runner %s has no executor", runner.Name)
		}
		
		switch runner.Executor {
		case "docker", "docker+machine", "docker-ssh", "docker-ssh+machine":
			if runner.Docker == nil || runner.Docker.Image == "" {
				return fmt.Errorf("runner %s: docker executor requires image", runner.Name)
			}
		case "kubernetes":
			if runner.Kubernetes == nil || runner.Kubernetes.Image == "" {
				return fmt.Errorf("runner %s: kubernetes executor requires image", runner.Name)
			}
		}
	}
	
	return nil
}