# Copilot Instructions for binmate

## Project Overview

**binmate** is a binary version manager CLI tool written in Go that allows users to install and manage multiple versions of command-line binaries from remote repositories (currently GitHub releases). It features an interactive TUI (Terminal User Interface) built with Bubble Tea for browsing and managing installations.

## Language and Localisation

- Always use UK English rather than US English when generating documentation or code snippets (e.g., "organisation" not "organization", "colour" not "color")
- Use context7 MCP to get up-to-date library and package documentation when necessary

## Architecture

### Project Structure

```
.
├── cmd/                      # Command entry point
│   └── main.go              # Root command setup and initialisation
├── internal/                # Internal packages
│   ├── cli/                 # CLI command definitions
│   │   ├── root/           # Root command (launches TUI)
│   │   └── install/        # Install command (CLI-based installation)
│   ├── core/               # Core business logic
│   │   ├── config/         # Configuration management
│   │   ├── install/        # Archive extraction (tar.gz, zip)
│   │   └── version/        # Version management and symlink handling
│   ├── providers/          # External provider integrations
│   │   └── github/         # GitHub releases API integration
│   └── tui/                # Terminal UI (Bubble Tea)
├── config.json             # Development config file
├── schema.json             # JSON schema for config validation
├── main.go                 # Application entry point
└── .github/
    └── agents/             # Custom Copilot agents
```

### Key Components

#### 1. Configuration (`internal/core/config/`)

- `config.go`: Defines the configuration structures for binaries
- `read_config.go`: Reads and parses `config.json`
- `get_binary.go`: Retrieves specific binary configurations

#### 2. CLI Commands (`internal/cli/`)

- **Root Command**: Launches the interactive TUI by default
- **Install Command**: CLI-based installation with flags `--binary` and `--version`
  - Supports aliases: `i`, `add`
  - Downloads from GitHub releases
  - Extracts archives and manages versioned installations

#### 3. Providers (`internal/providers/github/`)

- `fetch_release_asset.go`: Fetches release information from GitHub API
- `filter_assets.go`: Filters assets based on OS, architecture, and regex patterns
- `download_asset.go`: Downloads release assets
- `filter_assets_test.go`: Unit tests for asset filtering

#### 4. Core Installation (`internal/core/install/`)

- `extract.go`: Orchestrates extraction based on format
- `tar.go`: Handles `.tar.gz` extraction
- `zip.go`: Handles `.zip` extraction

#### 5. Version Management (`internal/core/version/`)

- `get_install_path.go`: Determines installation paths for versioned binaries
- `set_active_version.go`: Creates symlinks to activate specific versions

#### 6. TUI (`internal/tui/`)

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- `program.go`: Initialises the Bubble Tea program
- `model.go`: Defines the TUI model and state
- `init.go`: Initial command for the TUI
- `update.go`: Handles TUI updates and user interactions
- `view.go`: Renders the TUI

## Development Guidelines

### Go Version

- **Go 1.25.5** (specified in `go.mod`)

### Key Dependencies

- **Cobra**: CLI framework for command structure
- **Viper**: Configuration management
- **Bubble Tea**: Terminal UI framework
- **Lipgloss**: Terminal styling (transitive dependency)

### Code Style

- Follow standard Go conventions and idioms
- Use UK English in all documentation and user-facing messages
- Keep packages focused and single-purpose
- Prefer small, composable functions
- Use meaningful variable and function names

### Configuration Schema

- All binary configurations must conform to `schema.json`
- Required fields: `id`, `name`, `provider`, `path`, `format`
- Optional: `releaseRegex` for filtering release assets
- Supported providers: `github` (currently only provider)
- Supported formats: `.tar.gz`, `.zip`

### Provider Implementation

- Currently only GitHub is supported
- Provider logic is isolated in `internal/providers/github/`
- Asset filtering considers OS, architecture, and format
- Uses GitHub API v3 (REST)

### Testing

- Unit tests exist for critical functionality (e.g., `filter_assets_test.go`)
- Run tests with `go test ./...`
- Write tests for new provider logic and core business functions

### Building

- When running a test build to verify changes, write to the `/tmp` directory rather than the repository root directory

### Version Management

- Binaries are installed to versioned directories
- Symlinks are used to activate specific versions
- Installation path structure: `<install_path>/<name>/<version>/`

### Error Handling

- Use `log.Panicf()` for critical errors in CLI commands
- Return errors from internal functions for graceful handling
- Provide clear, actionable error messages

## Future Considerations

- Support for additional providers (e.g., GitLab, direct URLs)
- Enhanced TUI features (version switching, uninstallation)
- Configuration file customisation via `--config` flag (currently defined but not implemented)
- Shell integration for PATH management
- Automatic binary discovery and configuration generation
