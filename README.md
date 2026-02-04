# DEV GitHub Copilot CLI Challenge

Submission for the [2026 DEV GitHub Copilot CLI challenge](https://dev.to/devteam/join-the-github-copilot-cli-challenge-win-github-universe-tickets-copilot-pro-subscriptions-and-50af).

## About binmate

**binmate** is a CLI/TUI application for managing binary installations from GitHub releases. It provides an easy way to install, manage, and switch between different versions of command-line tools.

### Key Features

- **Interactive TUI**: Browse and manage binaries with a Terminal User Interface
- **CLI Commands**: Automate binary management with command-line interface
- **Version Management**: Install multiple versions and switch between them
- **GitHub Integration**: Automatically fetch releases from GitHub repositories
- **Database Tracking**: SQLite database tracks all installations and versions
- **Checksum Verification**: Ensures integrity of downloaded binaries

## Installation

<!-- TODO: Update this section once builds are published to GitHub releases -->

```bash
go build -o binmate .
```

## Usage

### Interactive Mode

Launch the TUI for interactive management:

```bash
binmate
```

### CLI Commands

#### Add a Binary

Add a binary from a GitHub release URL or from config:

```bash
# Add from URL
binmate add https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz

# Add from config
binmate add gh
```

#### List Binaries

List all registered binaries:

```bash
binmate list
```

List versions of a specific binary:

```bash
binmate list gh
```

#### Install a Binary

Install a specific version of a binary:

```bash
binmate install --binary gh --version v2.30.0
```

Install the latest version:

```bash
binmate install --binary gh --version latest
```

#### Switch Versions

Switch to a different installed version:

```bash
binmate switch gh v2.29.0
```

#### Update to Latest

Update a binary to the latest version:

```bash
binmate update gh
```

#### Remove a Binary

Remove a binary from the database:

```bash
binmate remove gh
```

Remove a binary and its files:

```bash
binmate remove gh --files
```

#### View Configuration

Display the current configuration:

```bash
binmate config
```

Display configuration as JSON:

```bash
binmate config --json
```

#### Sync Configuration

Sync the configuration file with the database:

```bash
binmate sync
```

## Configuration

Configuration is stored in `~/.config/.binmate/config.json`:

```json
{
  "version": 1,
  "binaries": [
    {
      "id": "gh",
      "name": "gh",
      "provider": "github",
      "path": "cli/cli",
      "format": ".tar.gz"
    }
  ]
}
```

### Configuration Fields

- `id`: Unique identifier for the binary
- `name`: Display name of the binary
- `provider`: Provider type (currently only "github" supported)
- `path`: Repository path (e.g., "owner/repo")
- `format`: Archive format (.tar.gz, .zip, .tgz)
- `installPath`: (optional) Custom installation path
- `assetRegex`: (optional) Regex to filter release assets
- `releaseRegex`: (optional) Regex to filter releases

## Database

binmate uses SQLite to track installations:

- Location: `~/.local/share/binmate/user.db`
- Tables: binaries, installations, versions, downloads, logs

## Architecture

The project follows a layered architecture:

```
cmd/                    # Command entry point
internal/
  cli/                  # CLI command definitions
    add/                # Add binary command
    config/             # Config command
    import/             # Import command
    install/            # Install command
    list/               # List command
    remove/             # Remove command
    switch/             # Switch version command
    sync/               # Sync config command
    update/             # Update command
  core/                 # Core business logic
    binary/             # Binary management service
    config/             # Configuration management
    crypto/             # Checksum verification
    install/            # Installation and extraction
    url/                # GitHub URL parsing
    version/            # Version management service
  database/             # SQLite data layer
    repository/         # Data access repositories
  providers/            # External provider integrations
    github/             # GitHub releases API
  tui/                  # Terminal UI (Bubble Tea)
```

## Copilot Agents

The project utilises custom copilot agents from the [awesome-copilot](https://github.com/github/awesome-copilot/blob/main/docs/README.agents.md) repository:
- Context7 Expert
- Critical Thinking
- Devils Advocate
- Gilfoyle
- Go MCP Server Development Expert (revised slightly to be a generic go agent, see [go-expert.agent.md](./.github/agents/go-expert.agent.md))
- SQLite Database Administrator (revised from PostgreSQL Database Administrator agent, see [sqlite-dba.agent.md](./.github/agents/sqlite-dba.agent.md))

## Models

This project utilised the following GitHub Copilot models:

- Claude Opus 4.5
- Claude Sonnet 4.5
- Claude Haiku 4.5
- OpenAI GPT 5.1 Codex Max (xhigh reasoning mode)