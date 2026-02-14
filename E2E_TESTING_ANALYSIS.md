# binmate E2E Testing Analysis

## Overview
This document provides a comprehensive analysis of the binmate CLI tool to inform the implementation of end-to-end testing.

## 1. CLI Commands Available

### Core Commands

#### `binmate` (TUI)
- **Purpose**: Launch interactive Terminal User Interface
- **Usage**: `binmate`
- **Pre-run**: Syncs config to database
- **Flags**: 
  - `--version, -v`: Show version
  - `--config, -c`: Custom config path
  - `--log-level, -l`: Control logging verbosity

#### `binmate add [binary-id|url]`
- **Purpose**: Add a new binary from GitHub release URL or config
- **Usage**: 
  - `binmate add https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz`
  - `binmate add gh` (from config)
- **Flags**:
  - `--url, -u`: GitHub release URL
  - `--authenticated, -a`: Use GitHub token authentication
- **Output**: `✓ Binary <id> added successfully`

#### `binmate list` (alias: `ls`)
- **Purpose**: List all installed binaries
- **Usage**: `binmate list`
- **Output**: Table format with Binary, Active Version, Installed count, Provider

#### `binmate install` (alias: `i`)
- **Purpose**: Install a specific version of a binary
- **Usage**: `binmate install --binary gh --version v2.30.0`
- **Flags**:
  - `--binary, -b`: Binary to install (required)
  - `--version, -v`: Version to install (default: "latest")
- **Pre-run**: Auto-syncs from config if binary not in DB
- **Output**: `✓ Successfully installed <binary> version <version>`

#### `binmate switch`
- **Purpose**: Switch to a different installed version
- **Usage**: `binmate switch --binary gh --version v2.30.0`
- **Flags**:
  - `--binary, -b`: Binary ID (required)
  - `--version, -v`: Version to switch to (required)
- **Output**: `✓ Switched <binary> to version <version>`

#### `binmate update`
- **Purpose**: Update a binary to latest version
- **Usage**: 
  - `binmate update --binary gh`
  - `binmate update --all`
- **Flags**:
  - `--binary, -b`: Binary ID to update
  - `--all, -a`: Update all binaries
- **Output**: `✓ Updated <binary> to version <version>`

#### `binmate remove` (aliases: `rm`, `delete`)
- **Purpose**: Remove a binary from database
- **Usage**: `binmate remove --binary gh --files`
- **Flags**:
  - `--binary, -b`: Binary ID (required)
  - `--files, -f`: Also remove physical files
- **Output**: 
  - `✓ Binary <id> removed from database`
  - `✓ Binary <id> removed (including files)`

#### `binmate import <path>`
- **Purpose**: Import an existing binary from filesystem
- **Usage**: 
  - `binmate import /usr/local/bin/gh --name gh`
  - `binmate import /usr/local/bin/gh --url <release-url>`
- **Flags**:
  - `--name, -n`: Name for the binary
  - `--version, -v`: Version string
  - `--url, -u`: GitHub release URL to associate
  - `--authenticated, -a`: Use GitHub auth
  - `--keep-location, -k`: Keep in original location
- **Output**: `✓ Binary <name> imported successfully`

#### `binmate sync`
- **Purpose**: Sync config file with database
- **Usage**: `binmate sync`
- **Output**: `✓ Sync complete`

#### `binmate config`
- **Purpose**: Display current configuration
- **Usage**: `binmate config [--json]`
- **Flags**:
  - `--json`: Output as JSON
- **Output**: Human-readable table or JSON

#### `binmate versions`
- **Purpose**: List installed versions of a binary
- **Usage**: `binmate versions --binary gh`
- **Flags**:
  - `--binary, -b`: Binary ID (required)
- **Output**: List of versions with active marker (*)

#### `binmate version` (alias: `v`)
- **Purpose**: Show binmate version information
- **Usage**: 
  - `binmate version`
  - `binmate --version`
  - `binmate version --verbose`
- **Flags**:
  - `--verbose, -V`: Show detailed build metadata
- **Output**: 
  - `binmate <version>`
  - Verbose: version, commit, date, modified

#### `binmate check`
- **Purpose**: Check for updates without installing
- **Usage**: 
  - `binmate check --binary gh`
  - `binmate check --all`
- **Flags**:
  - `--binary, -b`: Binary ID to check
  - `--all, -a`: Check all binaries
- **Output**: Update availability status

### Global Flags
- `--config, -c`: Path to config file
- `--log-level, -l`: Logging verbosity
- `--version, -v`: Show version (root command only)

## 2. install.sh Script Analysis

### Environment Variables
1. **BINMATE_VERSION**: Version to install (default: "latest")
2. **BINMATE_INSTALL_DIR**: Installation directory (default: "/usr/local/bin")
3. **BINMATE_SKIP_AUTO_IMPORT**: Skip automatic self-import (default: disabled)

### Key Functions

#### `detect_platform()`
- Detects OS: Linux, Darwin (macOS), Windows
- Detects architecture: amd64, arm64
- Returns: `${os}_${arch}` (e.g., "linux_amd64")
- Exits on unsupported platforms

#### `get_latest_version()`
- Queries GitHub API: `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`
- Extracts tag_name from JSON
- Returns: Version tag (e.g., "v1.0.0")

#### `validate_version(version)`
- Checks if version exists via GitHub API
- Returns HTTP status code 200 for valid versions
- Exits if version not found

#### `download_and_install()`
- **Downloads**: Archive from GitHub releases
- **Verifies**: SHA256 checksum from checksums.txt
  - Uses `sha256sum` (Linux) or `shasum` (macOS)
  - Warns if checksum tool unavailable
- **Extracts**: tar.gz archive
- **Installs**: Moves binary to install directory
- **Sets permissions**: `chmod +x`
- **Cleanup**: Removes temporary files

#### `run_auto_import()`
- **Post-install hook**: Automatically imports installed binmate for self-management
- **Command**: `binmate import <path> --url <release-url> --version <version> --keep-location`
- **Skippable**: Via BINMATE_SKIP_AUTO_IMPORT=1
- **Error handling**: Warns on failure, provides manual command

### Archive Naming Convention
```
binmate_${version#v}_${platform}.tar.gz
```
Examples:
- `binmate_1.0.0_linux_amd64.tar.gz`
- `binmate_1.0.0_darwin_arm64.tar.gz`

### Checksum Verification
- File: `checksums.txt` from release assets
- Algorithm: SHA256
- Format: Standard checksum file (compatible with sha256sum/shasum -c)

### PATH Detection
- Checks if install directory is in PATH
- Warns user if not in PATH with instructions

## 3. Test Scenarios for E2E Tests

### Installation Tests

#### Unix (Bash)
1. **Install Latest Version**
   - Run: `curl -fsSL <script> | bash`
   - Verify: Binary exists at install location
   - Verify: Binary is executable
   - Verify: Version command works
   - Verify: Auto-import succeeded (database entry exists)

2. **Install Specific Version**
   - Run: `BINMATE_VERSION=v1.0.0 bash install.sh`
   - Verify: Correct version installed
   - Verify: Version command shows correct version

3. **Custom Install Directory**
   - Run: `BINMATE_INSTALL_DIR=/tmp/bintest bash install.sh`
   - Verify: Binary installed to custom location
   - Verify: Binary works from custom location

4. **Skip Auto-Import**
   - Run: `BINMATE_SKIP_AUTO_IMPORT=1 bash install.sh`
   - Verify: Installation succeeds
   - Verify: No database entry for binmate (auto-import skipped)

5. **Platform Detection**
   - Test on: Linux (amd64, arm64)
   - Test on: macOS (amd64, arm64)
   - Verify: Correct platform binary downloaded

6. **Checksum Verification**
   - Verify: Checksum validation passes
   - Test: Corrupted checksum (should fail)

#### Windows (PowerShell)
1. **Install Latest Version**
   - Run: PowerShell install script
   - Verify: Binary exists at install location
   - Verify: Version command works
   - Verify: Auto-import succeeded

2. **Install Specific Version**
   - Run with version parameter
   - Verify correct version

3. **Custom Install Directory**
   - Test with custom path (Windows format)
   - Verify installation

4. **Skip Auto-Import**
   - Test skip flag
   - Verify no database entry

5. **Platform Detection**
   - Test on: Windows (amd64, arm64 if applicable)
   - Verify correct binary downloaded

6. **Checksum Verification**
   - Use PowerShell Get-FileHash
   - Verify SHA256 validation

### CLI Command Tests

#### Basic Operations
1. **Version Check**
   - `binmate --version` → outputs version
   - `binmate version` → outputs version
   - `binmate version --verbose` → outputs detailed info

2. **Config Display**
   - `binmate config` → shows config
   - `binmate config --json` → outputs JSON

3. **List Empty State**
   - `binmate list` → "No binaries installed"

4. **Sync Config**
   - `binmate sync` → "✓ Sync complete"
   - Verify database updated

#### Binary Management Workflow
1. **Add from Config**
   - `binmate add gh` → Binary added
   - `binmate list` → Shows gh (not installed)

2. **Install Latest**
   - `binmate install --binary gh --version latest`
   - Verify: gh binary downloaded and installed
   - Verify: Symlink/binary created in install path
   - `binmate list` → Shows gh with version

3. **List Versions**
   - `binmate versions --binary gh`
   - Verify: Shows installed version with * marker

4. **Install Second Version**
   - `binmate install --binary gh --version v2.29.0`
   - Verify: Second version installed
   - `binmate versions --binary gh` → Shows both versions

5. **Switch Version**
   - `binmate switch --binary gh --version v2.29.0`
   - Verify: Active version changed
   - `binmate list` → Shows new active version

6. **Check for Updates**
   - `binmate check --binary gh`
   - Verify: Shows update status
   - `binmate check --all` → Checks all binaries

7. **Update to Latest**
   - `binmate update --binary gh`
   - Verify: Updated to latest version
   - `binmate update --all` → Updates all

8. **Remove Binary (DB Only)**
   - `binmate remove --binary gh`
   - Verify: Removed from database
   - Verify: Files still exist

9. **Remove Binary (With Files)**
   - Re-add and install gh
   - `binmate remove --binary gh --files`
   - Verify: Removed from database
   - Verify: Files deleted

#### Add from URL
1. **Add via URL**
   - `binmate add https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz`
   - Verify: Binary added
   - `binmate list` → Shows binary

2. **Add with Flag**
   - `binmate add --url <release-url>`
   - Verify: Binary added

#### Import Binary
1. **Import with Name**
   - Create test binary in /tmp
   - `binmate import /tmp/testbin --name testbin`
   - Verify: Binary imported
   - `binmate list` → Shows testbin

2. **Import with URL**
   - `binmate import /tmp/testbin --url <release-url>`
   - Verify: Binary imported with GitHub association
   - Verify: Version extracted from URL

3. **Import with Keep Location**
   - `binmate import /tmp/testbin --name testbin --keep-location`
   - Verify: Binary referenced from original location
   - Verify: No copy made

#### Error Cases
1. **Install Non-existent Binary**
   - `binmate install --binary nonexistent --version latest`
   - Verify: Error message
   - Exit code: Non-zero

2. **Switch to Non-existent Version**
   - `binmate switch --binary gh --version v99.99.99`
   - Verify: Error message
   - Exit code: Non-zero

3. **Remove Non-existent Binary**
   - `binmate remove --binary nonexistent`
   - Verify: Error message
   - Exit code: Non-zero

4. **Missing Required Flags**
   - `binmate install` (no --binary)
   - Verify: Usage error
   - Exit code: Non-zero

### Database Tests
1. **Database Creation**
   - Run first command
   - Verify: Database created at `~/.local/share/binmate/user.db`
   - Verify: Tables exist

2. **Config Sync**
   - Modify config.json
   - Run `binmate sync`
   - Verify: Database updated

3. **Database Persistence**
   - Install binary
   - Restart (close DB)
   - Verify: Data persists

### Configuration Tests
1. **Default Config**
   - No config file
   - Verify: Uses defaults

2. **Custom Config Path**
   - `binmate --config /tmp/custom-config.json list`
   - Verify: Uses custom config

3. **Global Config Settings**
   - Test global.installPath
   - Test global.providers.github.authenticated
   - Verify: Settings applied

4. **Binary Override**
   - Binary with custom installPath
   - Verify: Overrides global setting

### GitHub Integration Tests
1. **Fetch Latest Release**
   - `binmate install --binary gh --version latest`
   - Verify: Latest version detected and installed

2. **Authenticated Requests**
   - Set GITHUB_TOKEN
   - `binmate add --url <private-repo-url> --authenticated`
   - Verify: Uses authentication

3. **Rate Limiting**
   - Make multiple unauthenticated requests
   - Verify: Handles rate limiting gracefully

## 4. Path Resolution and Environment Variables

### Directory Structure
```
~/.config/.binmate/          # Config directory
    └── config.json          # Configuration file

~/.local/share/binmate/      # Data directory
    └── user.db              # SQLite database

<installPath>/<binary>       # Binary installation (symlink or copy)

~/.local/share/binmate/binaries/  # Managed binaries location (inferred)
    └── <binary>/
        └── <version>/
            └── <binary>     # Actual binary
```

### Key Paths

#### Config Path Resolution
1. Flag: `--config <path>`
2. Default: `~/.config/.binmate/config.json`

#### Database Path
- Location: `~/.local/share/binmate/user.db`
- Resolved via: `database.GetDefaultDBPath()`

#### Install Path Resolution
1. Config: `global.installPath` or binary-specific `installPath`
2. Install script: `BINMATE_INSTALL_DIR` (default: `/usr/local/bin`)
3. Default: `/usr/local/bin` (typical)

#### Binary Storage Path
- Managed location for actual binaries (not symlinks)
- Likely: `~/.local/share/binmate/binaries/<binary>/<version>/<binary>`

### Environment Variables

#### For binmate
- `GITHUB_TOKEN`: GitHub personal access token for authenticated API calls
- `HOME`: Used for default path resolution
- `XDG_CONFIG_HOME`: Could affect config location (if supported)
- `XDG_DATA_HOME`: Could affect data location (if supported)

#### For install.sh
- `BINMATE_VERSION`: Version to install
- `BINMATE_INSTALL_DIR`: Installation directory
- `BINMATE_SKIP_AUTO_IMPORT`: Skip auto-import

#### For install.ps1 (to be created)
- `$env:BINMATE_VERSION`: Version to install
- `$env:BINMATE_INSTALL_DIR`: Installation directory
- `$env:BINMATE_SKIP_AUTO_IMPORT`: Skip auto-import

### Platform-Specific Considerations

#### Unix (Linux/macOS)
- Default shell: bash
- Path separator: `/`
- Home directory: `$HOME`
- Executable permission: Required (`chmod +x`)
- Archive format: `.tar.gz`
- Checksum tool: `sha256sum` (Linux) or `shasum` (macOS)

#### Windows
- Default shell: PowerShell
- Path separator: `\`
- Home directory: `$env:USERPROFILE` or `$env:HOME`
- Executable permission: Not required (implicit)
- Archive format: `.zip` (per goreleaser config)
- Checksum tool: `Get-FileHash`
- Binary extension: `.exe` (may need handling)

### Cross-Platform Path Handling
- Use forward slashes in Go (automatically converted on Windows)
- Use `filepath.Join()` for path construction
- Use `os.UserHomeDir()` for home directory
- Use `os.PathSeparator` for platform-specific separator

## 5. Test Data and Fixtures

### Test Binaries
For E2E tests, we can use:
1. **binmate itself** (self-referential testing)
2. **Small, stable binaries** from known repositories:
   - `cli/cli` (GitHub CLI)
   - `junegunn/fzf` (fuzzy finder)
   - Known old versions for version testing

### Test Config Files
```json
{
  "version": 1,
  "binaries": [
    {
      "id": "test-binary",
      "name": "test-binary",
      "provider": "github",
      "path": "cli/cli",
      "format": ".tar.gz"
    }
  ]
}
```

### Mock Release URLs
For testing URL parsing and installation:
```
https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz
https://github.com/junegunn/fzf/releases/download/v0.55.0/fzf-0.55.0-linux_amd64.tar.gz
```

## 6. Success Criteria

### Installation Success
- Binary exists at expected location
- Binary is executable
- Version command returns expected version
- Auto-import creates database entry (unless skipped)
- Checksum verification passes

### Command Success
- Exit code 0
- Expected output message (✓ prefix)
- Database state updated correctly
- Files created/modified/deleted as expected

### Error Handling
- Non-zero exit code
- Descriptive error message
- No partial state changes (atomicity where applicable)

## 7. Test Environment Requirements

### Dependencies
- Go 1.21+ (from go.mod)
- `curl` (for install script)
- `tar` (for extraction)
- `sha256sum` or `shasum` (for checksums)
- GitHub API access (rate limits apply)

### Test Isolation
- Use temporary directories for:
  - Installation directory
  - Config directory
  - Database location
- Clean up after each test
- Mock or use test GitHub releases

### CI/CD Considerations
- Run on multiple platforms (Linux, macOS, Windows)
- Test multiple architectures (amd64, arm64)
- Handle GitHub API rate limits (use GITHUB_TOKEN)
- Parallel test execution where possible
- Timeout protection (network operations)

## 8. Implementation Notes

### install.ps1 Considerations
1. **Archive Format**: Windows uses `.zip` (not `.tar.gz`)
2. **Binary Name**: May need `.exe` extension
3. **Extraction**: Use `Expand-Archive`
4. **Checksum**: Use `Get-FileHash -Algorithm SHA256`
5. **Platform Detection**: 
   - OS: `$env:OS` or `[System.Environment]::OSVersion`
   - Arch: `$env:PROCESSOR_ARCHITECTURE`
6. **Error Handling**: Use `$ErrorActionPreference = "Stop"`
7. **Auto-Import**: Same concept as bash version

### E2E Test Scripts
1. **Bash Version** (e2e-test.sh):
   - Test install.sh with various configurations
   - Test CLI commands in sequence
   - Clean up test artifacts
   - Exit with appropriate code

2. **PowerShell Version** (e2e-test.ps1):
   - Test install.ps1
   - Test CLI commands (same as bash)
   - Windows-specific path handling
   - Use `$LASTEXITCODE` for exit codes

### GitHub Actions Workflow
1. **Matrix Strategy**:
   - OS: ubuntu-latest, macos-latest, windows-latest
   - Architecture: Could test arm64 via emulation or runners
   
2. **Steps**:
   - Checkout code
   - Download release artifacts (or use latest release)
   - Run E2E test script for platform
   - Upload test logs as artifacts
   
3. **Triggers**:
   - On release creation (post-release verification)
   - Manual dispatch (for testing)
   - Scheduled (periodic verification)

### Issue Template
- Title: "Post-Release Verification for vX.Y.Z"
- Checklist of manual verification steps
- Links to automatic E2E test results
- Platform-specific verification items
- Sign-off section

## 9. Potential Issues and Mitigations

### Network Failures
- **Issue**: GitHub API or download failures
- **Mitigation**: Retry logic, timeout handling, fallback to cached releases

### Permission Issues
- **Issue**: Cannot write to `/usr/local/bin`
- **Mitigation**: Test with custom directory, document sudo requirement

### Platform-Specific Bugs
- **Issue**: Behavior differs across platforms
- **Mitigation**: Run E2E on all platforms, platform-specific test cases

### Database Locking
- **Issue**: SQLite database locked during tests
- **Mitigation**: Ensure proper cleanup, sequential test execution

### Version Parsing
- **Issue**: Inconsistent version formats
- **Mitigation**: Test with various version formats (v1.0.0, 1.0.0, etc.)

### PATH Issues
- **Issue**: Binary not in PATH after installation
- **Mitigation**: Test with absolute paths, document PATH setup

## 10. Next Steps

1. **Create install.ps1** - Windows equivalent of install.sh
2. **Create e2e-test.sh** - Unix E2E test script
3. **Create e2e-test.ps1** - Windows E2E test script
4. **Create .github/workflows/e2e.yml** - GitHub Actions workflow
5. **Create issue template** - Post-release verification checklist
6. **Update RELEASING.md** - Add E2E testing section

Each artifact should be comprehensive, well-documented, and follow the patterns established in the existing codebase.
