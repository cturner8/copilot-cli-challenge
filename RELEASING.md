# Releasing binmate

This document describes the release process for binmate.

## Prerequisites

- Write access to the GitHub repository
- All tests passing on the main branch

## Release Process

### 1. Prepare the Release

1. Ensure all changes are merged to the `main` or `dev` branch
2. Update the version number in relevant files (if needed)
3. Update CHANGELOG.md (if it exists) with release notes
4. Ensure all tests pass: `go test ./...`

### 2. Create a Release Tag

Create and push a version tag following semantic versioning (vMAJOR.MINOR.PATCH):

```bash
# Create a tag for the release
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Automated Release Process

Once the tag is pushed, the following happens automatically:

1. **GitHub Actions Triggers**: The release workflow (`.github/workflows/release.yml`) is triggered
2. **Tests Run**: All tests are executed with race detection enabled
3. **GoReleaser Builds**: If tests pass, GoReleaser builds binaries for:
   - Linux (amd64, arm64)
   - macOS/Darwin (amd64, arm64)
4. **Checksums Generated**: SHA256 checksums are computed for all binaries
5. **GitHub Release Created**: A new GitHub release is created with:
   - Release notes (auto-generated from commits)
   - Binary archives for each platform
   - Checksum file
   - Installation instructions

### 4. Verify the Release

1. Go to the [Releases page](https://github.com/cturner8/copilot-cli-challenge/releases)
2. Verify the release was created successfully
3. Run post-release verification (see [Post-Release Verification](#post-release-verification) section below)

## Post-Release Verification

After a release is published, comprehensive end-to-end testing should be performed to ensure the release works correctly across all supported platforms and architectures.

### Automated E2E Testing

The repository includes automated E2E tests that can be run via GitHub Actions:

1. Go to the [E2E Tests workflow](https://github.com/cturner8/copilot-cli-challenge/actions/workflows/e2e.yml)
2. Click "Run workflow"
3. Specify the version to test (e.g., `v1.0.0` or `latest`)
4. Select platforms and architectures to test (or use `all` for comprehensive testing)
5. Click "Run workflow" to start the tests

The workflow will:
- Test installation via install scripts (`install.sh` for Unix, `install.ps1` for Windows)
- Run 24 core functionality tests on each platform/architecture
- Upload test logs as artifacts
- Report pass/fail status for each combination

**Supported test combinations:**
- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)
- Windows: amd64, arm64

### Manual Local Testing

You can also run E2E tests locally on your machine:

#### Unix (Linux/macOS)

```bash
# Test latest version
./e2e-test.sh

# Test specific version
./e2e-test.sh v1.0.0

# Or use environment variable
BINMATE_VERSION=v1.0.0 ./e2e-test.sh
```

#### Windows (PowerShell)

```powershell
# Test latest version
.\e2e-test.ps1

# Test specific version
.\e2e-test.ps1 -Version v1.0.0

# Or use environment variable
$env:BINMATE_VERSION = "v1.0.0"
.\e2e-test.ps1
```

### Manual Installation Testing

For manual verification:

#### Unix (Linux/macOS)

```bash
# Test install.sh with latest version
curl -fsSL https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh | bash

# Test install.sh with specific version
curl -fsSL https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh | BINMATE_VERSION=v1.0.0 bash

# Test with custom install directory
curl -fsSL https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh | BINMATE_INSTALL_DIR=/tmp/binmate-test bash
```

#### Windows (PowerShell)

```powershell
# Test install.ps1 with latest version
irm https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1 | iex

# Test install.ps1 with specific version
$env:BINMATE_VERSION = "v1.0.0"
irm https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1 | iex

# Test with custom install directory
$env:BINMATE_INSTALL_DIR = "C:\Temp\binmate-test"
irm https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1 | iex
```

### Issue Tracking

To track verification progress, create a Post-Release Verification issue:

1. Go to [Issues → New Issue](https://github.com/cturner8/copilot-cli-challenge/issues/new/choose)
2. Select "Post-Release Verification" template
3. Fill in the version and release URL
4. Use the checklist to track testing progress for each platform
5. Link to automated E2E test results
6. Document any issues found
7. Close the issue once all verification is complete

The template includes comprehensive checklists for:
- All 6 platform/architecture combinations
- Installation testing
- Core functionality testing
- Error handling verification
- Additional release quality checks

## Release Workflow Details

### Test Workflow

The test workflow (`.github/workflows/test.yml`) runs on every push and pull request to main/dev branches:
- Runs all tests with race detection
- Generates code coverage reports
- Uploads coverage to Codecov (if configured)

### Release Workflow

The release workflow (`.github/workflows/release.yml`) runs on version tags:
- Runs all tests first (fails if tests fail)
- Uses GoReleaser to build cross-platform binaries
- Creates GitHub release with binaries and checksums
- Handles CGO requirements for SQLite3

## GoReleaser Configuration

The `.goreleaser.yml` file configures:
- **Platforms**: Linux and macOS (Windows disabled due to CGO complexity)
- **Architectures**: amd64 and arm64
- **Archive Format**: tar.gz
- **Checksums**: SHA256
- **Changelog**: Auto-generated from GitHub commits with SHA suppression and linked PR references
- **Build metadata**: version, commit, and build date injected via linker flags (`-X main.version`, `-X main.commit`, `-X main.date`)

## Troubleshooting

### Release Workflow Fails

1. Check the [Actions tab](https://github.com/cturner8/copilot-cli-challenge/actions) for error details
2. Common issues:
   - **Tests failing**: Fix tests before releasing
   - **CGO cross-compilation errors**: Ensure cross-compilation tools are installed
   - **GoReleaser errors**: Check `.goreleaser.yml` syntax

### Build Fails for Specific Platform

If a specific platform build fails:
1. Review the GoReleaser logs
2. Check the platform-specific environment variables in `.goreleaser.yml`
3. Verify cross-compilation tools are available

### Install Script Issues

If users report install script problems:
1. Test the script locally: `bash install.sh`
2. Verify the GitHub release exists and contains all binaries
3. Check checksum file is present and correct

## Version Numbering

binmate follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version: Breaking changes
- **MINOR** version: New features (backward compatible)
- **PATCH** version: Bug fixes (backward compatible)

Examples:
- `v1.0.0` - Initial release
- `v1.1.0` - New feature added
- `v1.1.1` - Bug fix
- `v2.0.0` - Breaking change

## Manual Release (Emergency)

If automated release fails and you need to release manually:

1. Build binaries locally:
   ```bash
   # Install GoReleaser
   go install github.com/goreleaser/goreleaser@latest
   
   # Build snapshot (test)
   goreleaser release --snapshot --clean
   
   # Build actual release (with tag)
   goreleaser release --clean
   ```

2. Create GitHub release manually:
   - Go to Releases → New Release
   - Upload binaries and checksums
   - Add release notes

## Manual Release Build

To verify a local build, run the following:

basic build

```bash
go build -o /tmp/binmate
```

build with additional metadata flags (normally set automatically by `goreleaser`)

```bash
go build -o /tmp/binmate \
   -ldflags "-X main.version=dev-local -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" .
```

## Post-Release

1. Announce the release in appropriate channels
2. Update documentation if needed
3. Monitor for issues reported by users
