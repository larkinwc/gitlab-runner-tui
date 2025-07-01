package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larkinwc/gitlab-runner-tui/pkg/config"
	"github.com/larkinwc/gitlab-runner-tui/pkg/runner"
)

type ConfigView struct {
	configMgr      *config.TOMLConfigManager
	config         *runner.Config
	inputs         []textinput.Model
	focusIndex     int
	err            error
	successMsg     string
	width          int
	height         int
	selectedRunner int
	editingRunner  bool
}

const (
	inputConcurrent = iota
	inputCheckInterval
	inputLogLevel
	inputRunnerLimit
	inputRunnerMaxBuilds
	inputRunnerTags
	inputCount
)

func NewConfigView(configPath string) *ConfigView {
	configMgr := config.NewTOMLConfigManager(configPath)

	inputs := make([]textinput.Model, inputCount)
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 256
		inputs[i] = t
	}

	inputs[inputConcurrent].Placeholder = "Concurrent jobs"
	inputs[inputConcurrent].Prompt = "Concurrent: "

	inputs[inputCheckInterval].Placeholder = "Check interval (seconds)"
	inputs[inputCheckInterval].Prompt = "Check Interval: "

	inputs[inputLogLevel].Placeholder = "Log level (debug/info/warn/error)"
	inputs[inputLogLevel].Prompt = "Log Level: "

	inputs[inputRunnerLimit].Placeholder = "Runner job limit"
	inputs[inputRunnerLimit].Prompt = "Job Limit: "

	inputs[inputRunnerMaxBuilds].Placeholder = "Max concurrent builds"
	inputs[inputRunnerMaxBuilds].Prompt = "Max Builds: "

	inputs[inputRunnerTags].Placeholder = "Comma-separated tags"
	inputs[inputRunnerTags].Prompt = "Tags: "

	return &ConfigView{
		configMgr:  configMgr,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (v *ConfigView) Init() tea.Cmd {
	return v.loadConfig
}

func (v *ConfigView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case configLoadedMsg:
		v.config = msg.config
		v.err = msg.err
		if v.config != nil {
			v.updateInputs()
		}
		return v, nil

	case configSavedMsg:
		if msg.err != nil {
			v.err = msg.err
			v.successMsg = ""
		} else {
			v.err = nil
			v.successMsg = "Configuration saved successfully!"
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			v.err = nil
			v.successMsg = ""
			if msg.String() == "tab" {
				v.focusIndex++
			} else {
				v.focusIndex--
			}

			inputsCount := 3
			if v.editingRunner {
				inputsCount = 6
			}

			if v.focusIndex < 0 {
				v.focusIndex = inputsCount - 1
			} else if v.focusIndex >= inputsCount {
				v.focusIndex = 0
			}

			for i := range v.inputs {
				if i == v.focusIndex {
					v.inputs[i].Focus()
				} else {
					v.inputs[i].Blur()
				}
			}
			return v, nil

		case "ctrl+s":
			return v, v.saveConfig

		case "r", "R":
			if !v.editingRunner {
				v.editingRunner = true
				v.focusIndex = 3
				v.updateRunnerInputs()
			}

		case "esc":
			if v.editingRunner {
				v.editingRunner = false
				v.focusIndex = 0
			}

		case "up", "down":
			if v.editingRunner && v.config != nil && len(v.config.Runners) > 0 {
				if msg.String() == "up" {
					v.selectedRunner--
					if v.selectedRunner < 0 {
						v.selectedRunner = len(v.config.Runners) - 1
					}
				} else {
					v.selectedRunner++
					if v.selectedRunner >= len(v.config.Runners) {
						v.selectedRunner = 0
					}
				}
				v.updateRunnerInputs()
			}
		}
	}

	for i := range v.inputs {
		var cmd tea.Cmd
		v.inputs[i], cmd = v.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *ConfigView) View() string {
	if v.config == nil {
		if v.err != nil {
			return ErrorBoxStyle.Render(fmt.Sprintf("Error loading config: %v", v.err))
		}
		return InfoBoxStyle.Render("Loading configuration...")
	}

	content := []string{
		HeaderStyle.Render("Configuration"),
		"",
	}

	if !v.editingRunner {
		content = append(content, TitleStyle.Render("Global Settings"))
		content = append(content, "")

		for i := 0; i < 3; i++ {
			style := InputStyle
			if i == v.focusIndex {
				style = FocusedInputStyle
			}
			content = append(content, style.Render(v.inputs[i].View()))
		}

		content = append(content, "")
		content = append(content, InfoBoxStyle.Render(fmt.Sprintf("Total Runners: %d", len(v.config.Runners))))
	} else {
		if len(v.config.Runners) == 0 {
			content = append(content, ErrorBoxStyle.Render("No runners configured"))
		} else {
			runner := v.config.Runners[v.selectedRunner]
			content = append(content, TitleStyle.Render(fmt.Sprintf("Runner: %s (%d/%d)", runner.Name, v.selectedRunner+1, len(v.config.Runners))))
			content = append(content, "")
			content = append(content, fmt.Sprintf("Executor: %s", runner.Executor))
			content = append(content, "")

			for i := 3; i < 6; i++ {
				style := InputStyle
				if i == v.focusIndex {
					style = FocusedInputStyle
				}
				content = append(content, style.Render(v.inputs[i].View()))
			}
		}
	}

	if v.err != nil {
		content = append(content, "", ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", v.err)))
	} else if v.successMsg != "" {
		content = append(content, "", SuccessBoxStyle.Render(v.successMsg))
	}

	// Help is now shown in the status bar

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (v *ConfigView) updateInputs() {
	if v.config == nil {
		return
	}

	v.inputs[inputConcurrent].SetValue(strconv.Itoa(v.config.Concurrent))
	v.inputs[inputCheckInterval].SetValue(strconv.Itoa(v.config.CheckInterval))
	v.inputs[inputLogLevel].SetValue(v.config.LogLevel)

	v.inputs[inputConcurrent].Focus()
	for i := 1; i < len(v.inputs); i++ {
		v.inputs[i].Blur()
	}
}

func (v *ConfigView) updateRunnerInputs() {
	if v.config == nil || len(v.config.Runners) == 0 {
		return
	}

	runner := v.config.Runners[v.selectedRunner]
	v.inputs[inputRunnerLimit].SetValue(strconv.Itoa(runner.Limit))
	v.inputs[inputRunnerMaxBuilds].SetValue(strconv.Itoa(runner.MaxBuilds))
	v.inputs[inputRunnerTags].SetValue(strings.Join(runner.TagList, ","))

	v.inputs[inputRunnerLimit].Focus()
	for i := 0; i < len(v.inputs); i++ {
		if i != inputRunnerLimit {
			v.inputs[i].Blur()
		}
	}
}

func (v *ConfigView) loadConfig() tea.Msg {
	if err := v.configMgr.Load(); err != nil {
		return configLoadedMsg{err: err}
	}
	return configLoadedMsg{config: v.configMgr.GetConfig()}
}

func (v *ConfigView) saveConfig() tea.Msg {
	if v.config == nil {
		return configSavedMsg{err: fmt.Errorf("no config loaded")}
	}

	if concurrent, err := strconv.Atoi(v.inputs[inputConcurrent].Value()); err == nil {
		v.configMgr.UpdateConcurrency(concurrent)
	}

	if interval, err := strconv.Atoi(v.inputs[inputCheckInterval].Value()); err == nil {
		v.configMgr.UpdateCheckInterval(interval)
	}

	if logLevel := v.inputs[inputLogLevel].Value(); logLevel != "" {
		v.configMgr.UpdateLogLevel(logLevel)
	}

	if v.editingRunner && len(v.config.Runners) > 0 {
		runner := v.config.Runners[v.selectedRunner]

		if limit, err := strconv.Atoi(v.inputs[inputRunnerLimit].Value()); err == nil {
			v.configMgr.UpdateRunnerLimit(runner.Name, limit)
		}

		if maxBuilds, err := strconv.Atoi(v.inputs[inputRunnerMaxBuilds].Value()); err == nil {
			v.configMgr.UpdateRunnerMaxBuilds(runner.Name, maxBuilds)
		}

		if tags := v.inputs[inputRunnerTags].Value(); tags != "" {
			tagList := strings.Split(tags, ",")
			for i := range tagList {
				tagList[i] = strings.TrimSpace(tagList[i])
			}
			v.configMgr.UpdateRunnerTags(runner.Name, tagList)
		}
	}

	if err := v.configMgr.Validate(); err != nil {
		return configSavedMsg{err: err}
	}

	if err := v.configMgr.Save(); err != nil {
		return configSavedMsg{err: err}
	}

	return configSavedMsg{}
}

type configLoadedMsg struct {
	config *runner.Config
	err    error
}

type configSavedMsg struct {
	err error
}
