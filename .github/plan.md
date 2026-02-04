# binmate Implementation Plan

## Project Overview

**binmate** is a CLI/TUI application for managing binary installations from remote repositories, with initial focus on GitHub releases. It provides a user-friendly interface to download, track, and
manage different versions of binaries.

**Problem Statement**: Developers need an easy way to install and manage precompiled binaries from remote sources without manual download/installation steps. Similar to nvm for Node.js, binmate should
simplify version management with rollback capabilities.

**Proposed Approach**: Build a modular Go application with layered architecture:
1. **Data Layer** – SQLite database for persistence
2. **Core Business Logic** – Installation, versioning, checksum verification
3. **Provider Layer** – Pluggable architecture for GitHub and future providers
4. **UI Layer** – TUI (Bubble Tea/Lip Gloss) for interactive use, CLI (Cobra) for automation

---

## Work Plan

### Phase 1: Project Foundation & Core Structure
- [ ] **Setup Go project** - Initialize module structure, dependencies (Cobra, Viper, Bubble Tea, Lip Gloss, SQLite)
- [ ] **Create directory structure** - cmd/, pkg/, internal/, config/ layout
- [ ] **Implement configuration management** - Viper configuration files (config.yaml, defaults)
- [ ] **Setup database schema** - SQLite database initialization, migrations, schema definition

**Considerations**:
- Database location: `$HOME/.config/.binmate/binmate.db`
- Cache location: `$HOME/.config/.binmate/cache/`
- Destination symlink directory: `$HOME/.local/bin/` (configurable)

---

### Phase 2: Core Domain Models & Database Layer
- [ ] **Define domain models**
    - `Binary` – Name, provider, current version, available versions
    - `Installation` – Binary name, version, provider, installed path, checksum, timestamp
    - `Provider` – Type (GitHub), configuration, API credentials
- [ ] **Implement database operations**
    - Create tables (binaries, installations, providers)
    - CRUD operations for each model
    - Query operations (list installed, check version, etc.)
- [ ] **Write database layer tests** - Ensure persistence works correctly

---

### Phase 3: Provider Interface & GitHub Implementation
- [ ] **Define provider interface** - Abstract interface for future extensibility
    - `FetchReleases()` – Get available versions
    - `DownloadAsset()` – Download specific binary asset
    - `GetChecksum()` – Verify asset integrity
- [ ] **Implement GitHub provider**
    - Use GitHub REST API (no authentication initially, support token later)
    - Parse release assets
    - Download binary artifacts
    - Support both release and pre-release versions
- [ ] **Write provider tests** - Mock GitHub API responses

**Design Decision Needed**: How to handle GitHub API rate limiting for unauthenticated requests? Should we support `GITHUB_TOKEN` environment variable?

---

### Phase 4: Installation & File Management
- [ ] **Implement installation logic**
    - Download binary to cache folder
    - Verify checksum (SHA256, MD5 support)
    - Extract if needed (tar.gz, zip support)
    - Create symlinks in destination directory
- [ ] **Implement version switching** - Change symlink to different installed version
- [ ] **Implement rollback** - Revert to previous version via symlink swap
- [ ] **Implement resume capability** - Track incomplete downloads, allow resumption
- [ ] **Handle cross-platform paths** - Test on Linux (primary), plan for macOS/Windows

---

### Phase 5: CLI Interface (Cobra)
- [ ] **Implement core commands**
    - `binmate` – Start interactive TUI session
    - `binmate add <binary> --provider github --repo owner/repo` – Add new binary
    - `binmate list` – Show installed binaries and versions
    - `binmate install <binary> <version>` – Install specific version
    - `binmate remove <binary>` – Remove binary and symlinks
    - `binmate switch <binary> <version>` – Switch between versions
    - `binmate update <binary>` – Upgrade to latest version
    - `binmate import <path>` – Import existing binary
    - `binmate config` – Show/edit configuration
- [ ] **Implement flags** - `--global`, `--local`, `--provider`, etc.
- [ ] **Add help documentation** - Clear descriptions for all commands
- [ ] **Write CLI tests** - Command parsing and execution

---

### Phase 6: TUI Interface (Bubble Tea/Lip Gloss)
- [ ] **Design TUI layout**
    - Main dashboard showing installed binaries
    - Navigation to install new, switch version, view details
    - Installation progress display
- [ ] **Implement main navigation view** - Menu-driven interface
- [ ] **Implement list/search view** - Browse available binaries
- [ ] **Implement installation view** - Interactive installation workflow
- [ ] **Implement version switcher view** - Select and switch versions
- [ ] **Add styling** - Use Lip Gloss for consistent colours/formatting
- [ ] **Write TUI tests** - Component interaction and state management

---

### Phase 7: Cross-Cutting Concerns
- [ ] **Implement logging** - Structured logging for debugging
- [ ] **Error handling** - Comprehensive error types and messages
- [ ] **Input validation** - Sanitise user input (binary names, versions, paths)
- [ ] **Security**
    - Checksum verification (mandatory)
    - Symlink safety (prevent directory traversal)
    - Secure temporary file handling
    - API credential handling (tokens)
- [ ] **Performance optimisation** - Cache API responses, parallel downloads if needed

---

### Phase 8: Testing & Quality
- [ ] **Unit tests** – Models, database, providers (target: >80% coverage)
- [ ] **Integration tests** – Full workflows (install → switch → remove)
- [ ] **Mock GitHub API** – Use httptest for provider tests
- [ ] **Manual testing** – Install real binaries, test switching/rollback
- [ ] **Documentation** – README, install instructions, usage guide

---

### Phase 9: Documentation & Release
- [ ] **Write comprehensive README** - Project overview, installation, quick start
- [ ] **Create examples** - Common use cases (installing gh-cli, jq, etc.)
- [ ] **Document configuration** - config.yaml options and defaults
- [ ] **Release preparation** - Build script, publish to GitHub releases

---

## Key Design Decisions (Awaiting Confirmation)

1. **GitHub API Authentication**: Support personal access tokens from environment variable or config?
2. **Binary Formats**: Which formats to support initially (ELF, Mach-O, PE)? Or auto-detect?
3. **Path Handling**: Should destination directory be configurable via config file?
4. **Database Migrations**: Simple append-only schema or support schema evolution?
5. **Error Recovery**: Auto-retry failed downloads or manual retry?
6. **Symlink Strategy**: How many versions to keep cached before cleanup?

---

## Tech Stack Summary
- **Language**: Go 1.21+
- **CLI Framework**: Cobra 1.7+
- **Config Management**: Viper 1.17+
- **TUI Framework**: Bubble Tea v0.24+
- **Styling**: Lip Gloss v0.9+
- **Database**: SQLite (sqlite3 Go driver)
- **Testing**: Go testing, httptest for mocks
- **Logging**: stdlib log package (or slog in Go 1.21+)

---

## Success Criteria
- ✅ CLI commands functional (install, list, switch, remove, import)
- ✅ TUI fully interactive and responsive
- ✅ Database properly tracks installations
- ✅ Checksum verification working
- ✅ Version switching via symlinks operational
- ✅ GitHub provider retrieving releases correctly
- ✅ Comprehensive test coverage (>70%)
- ✅ Documentation complete

---

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|-----------|
| Symlink handling on different OS | High | Early testing on Linux, Windows, macOS |
| GitHub API rate limiting | Medium | Cache responses, support auth tokens |
| Database schema changes | Medium | Plan v1 schema carefully, avoid migrations |
| Large binary downloads | Medium | Support resume capability, progress UI |
| Security of downloaded binaries | High | Mandatory checksum verification |