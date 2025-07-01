# GitLab Runner TUI

[![CI](https://github.com/larkinwc/gitlab-runner-tui/actions/workflows/ci.yml/badge.svg)](https://github.com/larkinwc/gitlab-runner-tui/actions/workflows/ci.yml)
[![Release](https://github.com/larkinwc/gitlab-runner-tui/actions/workflows/release.yml/badge.svg)](https://github.com/larkinwc/gitlab-runner-tui/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/larkinwc/gitlab-runner-tui)](https://goreportcard.com/report/github.com/larkinwc/gitlab-runner-tui)
[![Go Reference](https://pkg.go.dev/badge/github.com/larkinwc/gitlab-runner-tui.svg)](https://pkg.go.dev/github.com/larkinwc/gitlab-runner-tui)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Terminal User Interface (TUI) for managing GitLab runners on Debian hosts. This tool provides an interactive way to monitor runners, view logs, and update configurations.

## Features

- **Runner Management**: View all configured GitLab runners with their status
- **Log Viewer**: Real-time log viewing with filtering and auto-scroll
- **Job History**: View recent job runs with runner information, status, and duration
- **Configuration Editor**: Update runner concurrency, limits, and other settings
- **System Monitor**: View service status, CPU/memory usage, and restart services
- **Debug Mode**: Enable verbose logging for troubleshooting
- **Keyboard Navigation**: Easy tab-based navigation between views

## Prerequisites

- Go 1.19 or higher
- GitLab Runner installed on the system
- Appropriate permissions to:
  - Read GitLab Runner configuration files
  - Execute `gitlab-runner` commands
  - Access system services (systemctl/service)
  - Read system logs (journalctl)

## Installation

### Quick Install (Recommended)

```bash
# Download and install latest release
curl -sSL https://raw.githubusercontent.com/larkinwc/gitlab-runner-tui/main/install.sh | bash

# Or with wget
wget -qO- https://raw.githubusercontent.com/larkinwc/gitlab-runner-tui/main/install.sh | bash
```

### Manual Download

Download the latest release for your platform from the [releases page](https://github.com/larkinwc/gitlab-runner-tui/releases).

```bash
# Example for Linux x64
curl -L https://github.com/larkinwc/gitlab-runner-tui/releases/latest/download/gitlab-runner-tui_Linux_x86_64.tar.gz | tar xz
sudo mv gitlab-runner-tui /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/larkinwc/gitlab-runner-tui
cd gitlab-runner-tui

# Build the binary
go build -o gitlab-runner-tui cmd/gitlab-runner-tui/main.go

# Optional: Install to system path
sudo cp gitlab-runner-tui /usr/local/bin/
```

## Usage

```bash
# Run with default config path (/etc/gitlab-runner/config.toml)
gitlab-runner-tui

# Run with custom config path
gitlab-runner-tui -config /path/to/config.toml

# Run in debug mode for verbose logging
gitlab-runner-tui -debug

# Show help and default paths
gitlab-runner-tui -help

# If running without sudo, it will check ~/.gitlab-runner/config.toml
```

### Default Config Paths

The tool checks for configuration files in this order:
1. `/etc/gitlab-runner/config.toml` (system-wide)
2. `~/.gitlab-runner/config.toml` (user-specific)

You can override with the `-config` flag.

## Keyboard Shortcuts

### Global
- `Tab` / `Shift+Tab`: Navigate between tabs
- `1-5`: Jump to specific tab (Runners, Logs, Config, System, History)
- `q`: Quit (or go back from logs view)
- `Ctrl+C`: Force quit

### Runners View
- `↑/↓`: Navigate runner list
- `Enter`: View logs for selected runner
- `r`: Refresh runner list

### Logs View
- `↑/↓` / `PgUp/PgDn`: Scroll logs
- `g` / `G`: Go to top/bottom
- `a`: Toggle auto-scroll
- `c`: Clear logs
- `r`: Refresh logs

### Config View
- `Tab`: Navigate between fields
- `Ctrl+S`: Save configuration
- `r`: Edit runner-specific settings
- `↑/↓`: Select different runner (in runner edit mode)
- `Esc`: Exit runner edit mode

### System View
- `r`: Refresh system status
- `s`: Restart GitLab Runner service

### History View
- `r`: Refresh job history
- `↑/↓`: Navigate job list

## Configuration

The tool reads and modifies the standard GitLab Runner configuration file (usually `/etc/gitlab-runner/config.toml`).

### Editable Settings

**Global:**
- Concurrent job limit
- Check interval
- Log level

**Per-Runner:**
- Job limit
- Max concurrent builds
- Tags
- Run untagged jobs
- Locked status

## Security Considerations

- The tool requires read/write access to the GitLab Runner configuration
- Service management commands may require sudo privileges
- Runner tokens are displayed but can be masked in future versions

## Development

### Building from Source

```bash
# Get dependencies
go mod download

# Build
make build

# Install locally
make install
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Run tests with race detection (recommended)
go test -race ./...

# Generate local coverage badge
./scripts/coverage-badge.sh
```

#### Coverage Goals

- Current coverage: ~22.3%
- Target coverage: 70%
- Critical paths covered: Terminal size handling, config validation, runner parsing

The project uses both Codecov and Coveralls for tracking test coverage over time. Coverage reports are automatically generated on each push to the main branch.

### Code Quality

```bash
# Run linter
make lint

# Format code
go fmt ./...

# Run security scan
gosec ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make test`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.