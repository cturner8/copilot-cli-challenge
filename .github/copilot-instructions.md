# Copilot Instructions for binmate

## Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

### Starting a new work session

**MANDATORY WORKFLOW:**

1. **Checkout dev branch** - Switch back to the `dev` branch
2. **Pull from remote**:
   ```bash
   git pull
   git status  # MUST show "up to date with origin"
   ```
3. **Create a new session branch** - Create a new session branch for the current task. Branch name should be in the form `{type}/{issue id}`. For example:

- `bug/bm-nhs` - bug task
- `feature/bm-nhs` - feature task
- `task/bm-nhs` - generic task

### Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds, formatter
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** 
  - Provide context for next session
  - Provide user with a GitHub PR link to start the PR process

**CRITICAL RULES:**

- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

### Project Overview

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
│   ├── database/           # SQLite data layer
│   │   └── repository/     # Data access repositories
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

#### 6. Database (`internal/database/`)

- SQLite3-based persistence layer for installations and metadata
- `connection.go`: Database connection management
- `migrations.go`: Schema migration system
- `models.go`: Core data models (Binary, Installation, Version, Download, etc.)
- `schema.go`: Table definitions and schema versioning
- `repository/`: Data access layer with repositories for each entity
  - `binaries.go`: Binary definitions CRUD
  - `installations.go`: Installation tracking
  - `versions.go`: Active version management
  - `downloads.go`: Download cache tracking
  - `service.go`: High-level business operations

#### 7. TUI (`internal/tui/`)

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Follows the Elm Architecture (Model-Update-View)
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
- **Bubble Tea**: Terminal UI framework (Elm Architecture)
- **SQLite3**: Embedded database for state persistence
- **Lipgloss**: Terminal styling (transitive dependency via Bubble Tea)

### Code Style

- Follow standard Go conventions and idioms
- Use UK English in all documentation and user-facing messages
- Keep packages focused and single-purpose
- Prefer small, composable functions
- Use meaningful variable and function names
- When changing or adding new `go` files, run `go fmt` on changes files

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

- Unit tests exist for critical functionality (e.g., `filter_assets_test.go`, database tests)
- Run all tests: `go test ./...`
- Run specific test: `go test -run TestName ./path/to/package`
- Run with verbose output: `go test -v ./...`
- Write tests for new provider logic, core business functions, and database operations
- Database tests use in-memory SQLite (`:memory:`) for isolation

### Building

- Build the binary: `go build -o binmate .`
- Build to specific location: `go build -o /tmp/binmate .`
- Run without building: `go run . [args]`
- When running test builds to verify changes, write to the `/tmp` directory rather than the repository root directory

### Version Management

- Binaries are installed to versioned directories
- Symlinks are used to activate specific versions
- Installation path structure: `<install_path>/<name>/<version>/`
- Version state is tracked in SQLite database alongside filesystem state
- The `Version` table maintains active version references per binary

### Database Conventions

- Use SQLite for all persistent state (installations, versions, downloads, logs)
- Database location: User's config directory or specified via configuration
- Migrations are versioned and run automatically on connection
- All timestamps stored as Unix epoch (int64)
- Use prepared statements for all queries (already done in repositories)
- In-memory databases (`:memory:`) for testing isolation

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
