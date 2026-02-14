#!/usr/bin/env bash
#
# End-to-End Test Script for binmate (Unix)
#
# This script tests binmate installation and core functionality in an ephemeral environment.
# It can be run locally or in CI/CD pipelines.
#
# Usage:
#   ./e2e-test.sh [--version VERSION]
#   ./e2e-test.sh [VERSION]
#
# Environment variables:
#   BINMATE_VERSION - Alternative way to specify version
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Configuration
VERSION="${BINMATE_VERSION:-latest}"
TEST_DIR=""
BINMATE_BIN=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESOLVED_VERSION=""
RELEASE_TAG=""
BINMATE_ARCHIVE_URL=""
IMPORTED_BINARY_ID="binmate"
FZF_LATEST_VERSION=""
FZF_PREVIOUS_VERSION=""

# Functions
log_info() {
    echo -e "${GREEN}==>${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

log_error() {
    echo -e "${RED}Error:${NC} $1" >&2
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

test_passed() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${GREEN}✓ PASS${NC}"
}

test_failed() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${RED}✗ FAIL${NC}: $1"
}

github_curl() {
    local token="${GITHUB_TOKEN:-${GH_TOKEN:-}}"

    if [ -n "$token" ]; then
        curl -fsSL \
            -H "Authorization: Bearer ${token}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            "$@"
        return
    fi

    curl -fsSL "$@"
}

parse_args() {
    local positional_version_set=0

    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                if [ $# -lt 2 ]; then
                    log_error "--version requires a value"
                    exit 1
                fi
                VERSION="$2"
                shift 2
                ;;
            --version=*)
                VERSION="${1#*=}"
                shift
                ;;
            -h|--help)
                cat <<EOF
Usage: ./e2e-test.sh [--version VERSION]
       ./e2e-test.sh [VERSION]
EOF
                exit 0
                ;;
            *)
                if [ "$positional_version_set" -eq 0 ]; then
                    VERSION="$1"
                    positional_version_set=1
                    shift
                else
                    log_error "Unexpected argument: $1"
                    exit 1
                fi
                ;;
        esac
    done
}

download_test_config() {
    local config_dir="$HOME/.config/.binmate"
    local config_file="$config_dir/config.json"
    local local_config="$SCRIPT_DIR/config.json"
    local remote_config="https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/config.json"

    mkdir -p "$config_dir"

    if [ -f "$local_config" ]; then
        cp "$local_config" "$config_file"
    else
        if ! github_curl "$remote_config" -o "$config_file"; then
            return 1
        fi
    fi

    export BINMATE_CONFIG_PATH="$config_file"
    return 0
}

detect_release_platform() {
    local os_name=""
    local arch_name=""

    case "$(uname -s)" in
        Linux*)  os_name="linux" ;;
        Darwin*) os_name="darwin" ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            return 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)    arch_name="amd64" ;;
        aarch64|arm64)   arch_name="arm64" ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            return 1
            ;;
    esac

    echo "${os_name}_${arch_name}"
}

fetch_fzf_release_versions() {
    local releases_output=""
    local tags=""
    local candidate=""

    if ! releases_output=$(github_curl "https://api.github.com/repos/junegunn/fzf/releases?per_page=5"); then
        return 1
    fi

    tags=$(echo "$releases_output" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    FZF_PREVIOUS_VERSION=""
    while IFS= read -r candidate; do
        if [ -n "$candidate" ] && [ "$candidate" != "$FZF_LATEST_VERSION" ]; then
            FZF_PREVIOUS_VERSION="$candidate"
            break
        fi
    done <<EOF
$tags
EOF

    if [ -z "$FZF_LATEST_VERSION" ] || [ -z "$FZF_PREVIOUS_VERSION" ]; then
        return 1
    fi

    return 0
}

cleanup() {
    if [ -n "$TEST_DIR" ] && [ -d "$TEST_DIR" ]; then
        log_info "Cleaning up test environment..."
        rm -rf "$TEST_DIR"
    fi
}

# Setup trap for cleanup
trap cleanup EXIT INT TERM

# Phase 1: Environment Setup
setup_environment() {
    log_info "=== Phase 1: Environment Setup ==="
    
    # Create ephemeral test directory
    TEST_DIR=$(mktemp -d "${TMPDIR:-/tmp}/binmate-e2e-XXXXXX")
    log_info "Created test directory: $TEST_DIR"
    
    # Set up isolated HOME
    export HOME="$TEST_DIR/home"
    mkdir -p "$HOME/.config/.binmate" "$HOME/.local/bin"
    log_info "Set HOME to: $HOME"
    
    # Set install directory
    export BINMATE_INSTALL_DIR="$TEST_DIR/bin"
    mkdir -p "$BINMATE_INSTALL_DIR"
    log_info "Set BINMATE_INSTALL_DIR to: $BINMATE_INSTALL_DIR"

    # Skip installer auto-import so import tests can exercise explicit import behaviour
    export BINMATE_SKIP_AUTO_IMPORT=true
    
    # Add to PATH
    export PATH="$BINMATE_INSTALL_DIR:$HOME/.local/bin:$PATH"
    
    # Set version
    export BINMATE_VERSION="$VERSION"
    log_info "Testing version: $VERSION"
    if [ -n "${GITHUB_TOKEN:-${GH_TOKEN:-}}" ]; then
        log_info "Using GitHub authentication header for API requests"
    fi

    # Ensure config is available in the isolated environment
    if ! download_test_config; then
        test_failed "Failed to prepare config.json in isolated HOME"
        exit 1
    fi
    log_info "Config file prepared at: $BINMATE_CONFIG_PATH"
    
    echo ""
}

# Phase 2: Installation
install_binmate() {
    log_info "=== Phase 2: Installation ==="
    
    log_test "Installing binmate via install.sh"
    
    # Prefer local install.sh when available, fallback to remote
    local install_script="$SCRIPT_DIR/install.sh"
    if [ ! -f "$install_script" ]; then
        install_script="$TEST_DIR/install.sh"
        if ! github_curl "https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh" -o "$install_script"; then
            test_failed "Failed to download install.sh"
            exit 1
        fi
    fi

    if ! bash "$install_script"; then
        test_failed "Installation failed"
        exit 1
    fi
    
    test_passed
    
    # Verify binary exists
    log_test "Verifying binary installation"
    BINMATE_BIN="$BINMATE_INSTALL_DIR/binmate"
    
    if [ ! -f "$BINMATE_BIN" ]; then
        test_failed "Binary not found at $BINMATE_BIN"
        exit 1
    fi
    
    if [ ! -x "$BINMATE_BIN" ]; then
        test_failed "Binary is not executable"
        exit 1
    fi
    
    test_passed

    log_test "Resolving installed binmate version"
    if version_output=$("$BINMATE_BIN" version 2>&1); then
        RESOLVED_VERSION=$(echo "$version_output" | awk '{print $2}')
        if [ -z "$RESOLVED_VERSION" ]; then
            test_failed "Unable to resolve installed version from output: $version_output"
            exit 1
        fi
        RELEASE_TAG="$RESOLVED_VERSION"
        if [[ "$RELEASE_TAG" != v* ]]; then
            RELEASE_TAG="v${RELEASE_TAG}"
        fi
    else
        test_failed "Failed to resolve installed version"
        exit 1
    fi

    local platform=""
    if ! platform=$(detect_release_platform); then
        test_failed "Failed to detect release platform"
        exit 1
    fi

    log_test "Resolving release archive URL for ${platform}"
    local release_json=""
    if ! release_json=$(github_curl "https://api.github.com/repos/cturner8/copilot-cli-challenge/releases/tags/${RELEASE_TAG}"); then
        test_failed "Failed to fetch release metadata for ${RELEASE_TAG}"
        exit 1
    fi

    BINMATE_ARCHIVE_URL=$(echo "$release_json" \
        | grep '"browser_download_url":' \
        | sed -E 's/.*"([^"]+)".*/\1/' \
        | grep "_${platform}\.tar\.gz$" \
        | head -n 1)

    if [ -z "$BINMATE_ARCHIVE_URL" ]; then
        test_failed "Could not resolve archive URL for platform ${platform}"
        exit 1
    fi

    local asset_name=""
    asset_name="${BINMATE_ARCHIVE_URL##*/}"
    asset_name="${asset_name%.tar.gz}"
    asset_name="${asset_name%.zip}"
    asset_name="${asset_name%.tgz}"
    IMPORTED_BINARY_ID=$(echo "$asset_name" | sed -E 's/[-_].*$//')

    if [ -z "$IMPORTED_BINARY_ID" ]; then
        test_failed "Failed to derive imported binary ID from asset ${asset_name}"
        exit 1
    fi

    test_passed

    echo ""
}

# Phase 3: Core Functionality Tests
run_core_tests() {
    log_info "=== Phase 3: Core Functionality Tests ==="
    
    # Test 1: Version command
    log_test "Test 1: binmate version"
    if output=$("$BINMATE_BIN" version 2>&1); then
        if echo "$output" | grep -q "binmate"; then
            test_passed
        else
            test_failed "Version output doesn't contain 'binmate': $output"
        fi
    else
        test_failed "Version command failed"
    fi
    
    # Test 2: Verbose version command
    log_test "Test 2: binmate version --verbose"
    if output=$("$BINMATE_BIN" version --verbose 2>&1); then
        if echo "$output" | grep -q "version:" && echo "$output" | grep -q "commit:"; then
            test_passed
        else
            test_failed "Verbose version output missing expected fields: $output"
        fi
    else
        test_failed "Verbose version command failed"
    fi
    
    # Test 3: Config command
    log_test "Test 3: binmate config"
    if output=$("$BINMATE_BIN" config 2>&1); then
        if echo "$output" | grep -q "binmate Configuration"; then
            test_passed
        else
            test_failed "Config output missing header: $output"
        fi
    else
        test_failed "Config command failed"
    fi
    
    # Test 4: Config JSON command
    log_test "Test 4: binmate config --json"
    if output=$("$BINMATE_BIN" config --json 2>&1); then
        if (echo "$output" | grep -q '"Binaries"' || echo "$output" | grep -q '"binaries"') && echo "$output" | grep -q '{'; then
            test_passed
        else
            test_failed "Config JSON output missing expected fields: $output"
        fi
    else
        test_failed "Config --json command failed"
    fi
    
    # Test 5: List command before sync
    log_test "Test 5: binmate list (pre-sync)"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -q "No binaries installed"; then
            test_passed
        else
            test_failed "Expected no binaries before sync: $output"
        fi
    else
        test_failed "List command failed before sync"
    fi

    # Test 6: Sync command
    log_test "Test 6: binmate sync"
    if output=$("$BINMATE_BIN" sync 2>&1); then
        if echo "$output" | grep -q "Sync complete"; then
            test_passed
        else
            test_failed "Sync output missing completion message: $output"
        fi
    else
        test_failed "Sync command failed"
    fi

    # Test 7: List command after sync
    log_test "Test 7: binmate list (post-sync)"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -qi "fzf"; then
            test_passed
        else
            test_failed "Expected fzf in list output after sync: $output"
        fi
    else
        test_failed "List command failed after sync"
    fi

    # Test 8: Check all command
    log_test "Test 8: binmate check --all"
    if "$BINMATE_BIN" check --all >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Check --all command failed"
    fi
    
    # Test 9: Import installed binmate binary
    log_test "Test 9: binmate import (self)"
    if "$BINMATE_BIN" import "$BINMATE_BIN" --url "$BINMATE_ARCHIVE_URL" --version "$RELEASE_TAG" --keep-location >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Import command failed"
    fi
    
    # Test 10: Versions command for imported binary
    log_test "Test 10: binmate versions --binary ${IMPORTED_BINARY_ID}"
    if output=$("$BINMATE_BIN" versions --binary "$IMPORTED_BINARY_ID" 2>&1); then
        if echo "$output" | grep -q "$RELEASE_TAG"; then
            test_passed
        else
            test_failed "binmate version list does not include $RELEASE_TAG: $output"
        fi
    else
        test_failed "versions command failed for imported binary (${IMPORTED_BINARY_ID})"
    fi

    # Test 11: Install fzf latest
    log_test "Test 11: binmate install --binary fzf --version latest"
    if "$BINMATE_BIN" install --binary fzf --version latest >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Failed to install fzf latest"
    fi
    
    # Test 12: Capture fzf latest installed version
    log_test "Test 12: binmate versions --binary fzf"
    if output=$("$BINMATE_BIN" versions --binary fzf 2>&1); then
        FZF_LATEST_VERSION=$(echo "$output" | awk '/^\*/{print $2; exit}')
        if [ -n "$FZF_LATEST_VERSION" ]; then
            test_passed
        else
            test_failed "Unable to determine active fzf version: $output"
        fi
    else
        test_failed "versions command failed for fzf"
    fi

    # Test 13: List shows fzf
    log_test "Test 13: binmate list includes fzf"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -qi "fzf"; then
            test_passed
        else
            test_failed "fzf missing from list output: $output"
        fi
    else
        test_failed "List command failed"
    fi

    # Test 14: fzf executable runs
    log_test "Test 14: fzf --version"
    local fzf_path=""
    if fzf_path=$(command -v fzf 2>/dev/null); then
        if [[ "$fzf_path" != "$HOME/.local/bin/"* ]]; then
            test_failed "fzf resolved outside isolated HOME: $fzf_path"
        elif output=$("$fzf_path" --version 2>&1); then
            if [ -n "$output" ]; then
                test_passed
            else
                test_failed "fzf --version returned empty output"
            fi
        else
            test_failed "fzf binary did not execute: $fzf_path"
        fi
    else
        test_failed "fzf command is not executable from PATH"
    fi

    # Test 15: Update fzf
    log_test "Test 15: binmate update --binary fzf"
    if "$BINMATE_BIN" update --binary fzf >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Update command failed for fzf"
    fi

    # Test 16: Update imported binary
    log_test "Test 16: binmate update --binary ${IMPORTED_BINARY_ID}"
    if "$BINMATE_BIN" update --binary "$IMPORTED_BINARY_ID" >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Update command failed for imported binary (${IMPORTED_BINARY_ID})"
    fi

    # Test 17: Fetch previous fzf version
    log_test "Test 17: Resolve latest and previous fzf releases"
    if fetch_fzf_release_versions; then
        if [ "$FZF_LATEST_VERSION" = "$FZF_PREVIOUS_VERSION" ]; then
            test_failed "Unable to determine a distinct previous fzf version"
        else
            test_passed
        fi
    else
        test_failed "Failed to fetch fzf release versions from GitHub API"
    fi

    # Test 18: Install previous fzf version
    log_test "Test 18: binmate install --binary fzf --version previous"
    if "$BINMATE_BIN" install --binary fzf --version "$FZF_PREVIOUS_VERSION" >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Failed to install previous fzf version: $FZF_PREVIOUS_VERSION"
    fi

    # Test 19: Switch fzf to latest
    log_test "Test 19: binmate switch --binary fzf --version latest-installed"
    if "$BINMATE_BIN" switch --binary fzf --version "$FZF_LATEST_VERSION" >/dev/null 2>&1; then
        if output=$("$BINMATE_BIN" versions --binary fzf 2>&1); then
            if echo "$output" | grep -q "^\* ${FZF_LATEST_VERSION}"; then
                test_passed
            else
                test_failed "Active fzf version is not ${FZF_LATEST_VERSION}: $output"
            fi
        else
            test_failed "Failed to read fzf versions after switch"
        fi
    else
        test_failed "Failed to switch fzf to ${FZF_LATEST_VERSION}"
    fi

    # Test 20: Switch fzf to previous
    log_test "Test 20: binmate switch --binary fzf --version previous"
    if "$BINMATE_BIN" switch --binary fzf --version "$FZF_PREVIOUS_VERSION" >/dev/null 2>&1; then
        if output=$("$BINMATE_BIN" versions --binary fzf 2>&1); then
            if echo "$output" | grep -q "^\* ${FZF_PREVIOUS_VERSION}"; then
                test_passed
            else
                test_failed "Active fzf version is not ${FZF_PREVIOUS_VERSION}: $output"
            fi
        else
            test_failed "Failed to read fzf versions after switch"
        fi
    else
        test_failed "Failed to switch fzf to ${FZF_PREVIOUS_VERSION}"
    fi

    # Test 21: Switch back to latest
    log_test "Test 21: binmate switch --binary fzf --version latest-installed (restore)"
    if "$BINMATE_BIN" switch --binary fzf --version "$FZF_LATEST_VERSION" >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Failed to switch fzf back to ${FZF_LATEST_VERSION}"
    fi

    # Test 22: Remove fzf binary and files
    log_test "Test 22: binmate remove --binary fzf --files"
    if "$BINMATE_BIN" remove --binary fzf --files >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Remove command failed for fzf"
    fi

    # Test 23: List no longer includes fzf
    log_test "Test 23: binmate list excludes fzf after removal"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -qi "fzf"; then
            test_failed "fzf still present in list after removal: $output"
        else
            test_passed
        fi
    else
        test_failed "List command failed after removing fzf"
    fi

    # Test 24: fzf no longer executable from PATH
    log_test "Test 24: fzf binary removed from PATH"
    if [ -e "$HOME/.local/bin/fzf" ]; then
        test_failed "fzf symlink still exists at $HOME/.local/bin/fzf"
    else
        test_passed
    fi
    
    echo ""
}

# Phase 4: Database Tests
run_database_tests() {
    log_info "=== Phase 4: Database Tests ==="
    
    # Test 13: Database file creation
    log_test "Test 13: Database file exists"
    local db_path="$HOME/.local/share/binmate/user.db"
    
    # Trigger database creation by running a command
    "$BINMATE_BIN" list >/dev/null 2>&1 || true
    
    if [ -f "$db_path" ]; then
        test_passed
    else
        test_failed "Database file not created at $db_path"
    fi
    
    # Test 14: Database is readable
    log_test "Test 14: Database is readable"
    if [ -r "$db_path" ]; then
        test_passed
    else
        test_failed "Database file is not readable"
    fi
    
    echo ""
}

# Phase 5: Import Tests
run_import_tests() {
    log_info "=== Phase 5: Import Tests ==="
    
    # Test 27: Re-importing binmate is idempotent
    log_test "Test 27: Re-import binmate binary (idempotency)"
    if "$BINMATE_BIN" import "$BINMATE_BIN" --url "$BINMATE_ARCHIVE_URL" --version "$RELEASE_TAG" --keep-location >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Re-import command failed"
    fi
    
    # Test 28: List should show imported binary entry
    log_test "Test 28: binmate list shows ${IMPORTED_BINARY_ID}"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -qi "$IMPORTED_BINARY_ID"; then
            test_passed
        else
            test_failed "List output doesn't show ${IMPORTED_BINARY_ID}: $output"
        fi
    else
        test_failed "List command failed"
    fi
    
    echo ""
}

# Phase 6: Error Handling Tests
run_error_handling_tests() {
    log_info "=== Phase 6: Error Handling Tests ==="
    
    # Test 29: Install non-existent binary (should fail)
    log_test "Test 29: Install non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" install --binary nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Install should fail for non-existent binary"
    fi
    
    # Test 30: Remove non-existent binary (should fail)
    log_test "Test 30: Remove non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" remove --binary nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Remove should fail for non-existent binary"
    fi
    
    # Test 31: Switch non-existent binary (should fail)
    log_test "Test 31: Switch non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" switch --binary nonexistent-binary-xyz --version v1.0.0 >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Switch should fail for non-existent binary"
    fi
    
    # Test 32: Update non-existent binary (should fail)
    log_test "Test 32: Update non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" update --binary nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Update should fail for non-existent binary"
    fi
    
    echo ""
}

# Phase 7: Path Tests
run_path_tests() {
    log_info "=== Phase 7: Path Tests ==="
    
    # Test 33: Config file exists
    log_test "Test 33: Isolated config.json exists"
    if [ -f "$BINMATE_CONFIG_PATH" ]; then
        test_passed
    else
        test_failed "Config file not found at $BINMATE_CONFIG_PATH"
    fi
    
    # Test 34: Data directory exists
    log_test "Test 34: Data directory created"
    if [ -d "$HOME/.local/share/binmate" ]; then
        test_passed
    else
        test_failed "Data directory not created"
    fi
    
    # Test 35: Binary is in PATH
    log_test "Test 35: Binary is accessible via PATH"
    if command -v binmate >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Binary not found in PATH"
    fi
    
    # Test 36: Binary executes correctly
    log_test "Test 36: Binary executes without error"
    if "$BINMATE_BIN" version >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Binary execution failed"
    fi
    
    echo ""
}

# Main
main() {
    parse_args "$@"

    log_info "========================================="
    log_info "  binmate End-to-End Test Suite (Unix)"
    log_info "========================================="
    echo ""
    
    setup_environment
    install_binmate
    run_core_tests
    run_database_tests
    run_import_tests
    run_error_handling_tests
    run_path_tests
    
    # Summary
    log_info "========================================="
    log_info "  Test Summary"
    log_info "========================================="
    echo "Total tests: $TESTS_TOTAL"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_info "All tests passed! ✓"
        exit 0
    else
        log_error "$TESTS_FAILED test(s) failed!"
        exit 1
    fi
}

main "$@"
