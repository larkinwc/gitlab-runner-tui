package config

import (
	"os"
	"testing"
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

	if _, err := tmpFile.Write([]byte(testConfig)); err != nil {
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