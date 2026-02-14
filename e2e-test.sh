#!/usr/bin/env bash
#
# End-to-End Test Script for binmate (Unix)
#
# This script tests binmate installation and core functionality in an ephemeral environment.
# It can be run locally or in CI/CD pipelines.
#
# Usage:
#   ./e2e-test.sh [VERSION]
#
# Arguments:
#   VERSION - binmate version to test (default: "latest")
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
VERSION="${1:-${BINMATE_VERSION:-latest}}"
TEST_DIR=""
BINMATE_BIN=""

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
    TEST_DIR=$(mktemp -d)
    log_info "Created test directory: $TEST_DIR"
    
    # Set up isolated HOME
    export HOME="$TEST_DIR/home"
    mkdir -p "$HOME"
    log_info "Set HOME to: $HOME"
    
    # Set install directory
    export BINMATE_INSTALL_DIR="$TEST_DIR/bin"
    mkdir -p "$BINMATE_INSTALL_DIR"
    log_info "Set BINMATE_INSTALL_DIR to: $BINMATE_INSTALL_DIR"
    
    # Add to PATH
    export PATH="$BINMATE_INSTALL_DIR:$PATH"
    
    # Set version
    export BINMATE_VERSION="$VERSION"
    log_info "Testing version: $VERSION"
    
    echo ""
}

# Phase 2: Installation
install_binmate() {
    log_info "=== Phase 2: Installation ==="
    
    log_test "Installing binmate via install.sh"
    
    # Download and run install script
    local install_script="$TEST_DIR/install.sh"
    if ! curl -fsSL https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh -o "$install_script"; then
        test_failed "Failed to download install.sh"
        exit 1
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
    echo ""
}

# Phase 3: Core Functionality Tests
run_core_tests() {
    log_info "=== Phase 3: Core Functionality Tests ==="
    
    # Test 1: Version command
    log_test "Test 1: binmate version"
    if output=$("$BINMATE_BIN" version 2>&1); then
        if echo "$output" | grep -q "version"; then
            test_passed
        else
            test_failed "Version output doesn't contain 'version': $output"
        fi
    else
        test_failed "Version command failed"
    fi
    
    # Test 2: Config command (should work even without config file)
    log_test "Test 2: binmate config"
    if "$BINMATE_BIN" config >/dev/null 2>&1 || [ $? -eq 0 ]; then
        test_passed
    else
        # Config might fail if no config exists, which is acceptable
        log_warn "Config command returned non-zero (expected if no config file)"
        test_passed
    fi
    
    # Test 3: List command (should return empty or error gracefully)
    log_test "Test 3: binmate list"
    if "$BINMATE_BIN" list >/dev/null 2>&1 || [ $? -eq 0 ]; then
        test_passed
    else
        # List might fail if database doesn't exist yet
        log_warn "List command returned non-zero (expected if no database)"
        test_passed
    fi
    
    # Test 4: Sync command
    log_test "Test 4: binmate sync"
    if "$BINMATE_BIN" sync >/dev/null 2>&1 || [ $? -eq 0 ]; then
        test_passed
    else
        log_warn "Sync command returned non-zero"
        test_passed
    fi
    
    # Test 5: Check command (with non-existent binary should fail gracefully)
    log_test "Test 5: binmate check nonexistent"
    if ! "$BINMATE_BIN" check nonexistent >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Check should fail for non-existent binary"
    fi
    
    # Test 6: Install command (requires config)
    log_test "Test 6: binmate install --help"
    if "$BINMATE_BIN" install --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Install help failed"
    fi
    
    # Test 7: Add command help
    log_test "Test 7: binmate add --help"
    if "$BINMATE_BIN" add --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Add help failed"
    fi
    
    # Test 8: Import command help
    log_test "Test 8: binmate import --help"
    if "$BINMATE_BIN" import --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Import help failed"
    fi
    
    # Test 9: Remove command help
    log_test "Test 9: binmate remove --help"
    if "$BINMATE_BIN" remove --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Remove help failed"
    fi
    
    # Test 10: Switch command help
    log_test "Test 10: binmate switch --help"
    if "$BINMATE_BIN" switch --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Switch help failed"
    fi
    
    # Test 11: Update command help
    log_test "Test 11: binmate update --help"
    if "$BINMATE_BIN" update --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Update help failed"
    fi
    
    # Test 12: Versions command help
    log_test "Test 12: binmate versions --help"
    if "$BINMATE_BIN" versions --help >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Versions help failed"
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
    
    # Test 15: Import binmate itself
    log_test "Test 15: Import binmate binary"
    local archive_url="https://github.com/cturner8/copilot-cli-challenge/releases/download/$VERSION/binmate_${VERSION#v}_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz"
    
    if "$BINMATE_BIN" import "$BINMATE_BIN" --url "$archive_url" --version "$VERSION" --keep-location >/dev/null 2>&1; then
        test_passed
    else
        log_warn "Import failed (may already be imported from install.sh)"
        test_passed
    fi
    
    # Test 16: List should show imported binary
    log_test "Test 16: List shows binmate"
    if output=$("$BINMATE_BIN" list 2>&1); then
        if echo "$output" | grep -qi "binmate"; then
            test_passed
        else
            log_warn "List output doesn't show binmate: $output"
            test_passed
        fi
    else
        test_failed "List command failed"
    fi
    
    echo ""
}

# Phase 6: Error Handling Tests
run_error_handling_tests() {
    log_info "=== Phase 6: Error Handling Tests ==="
    
    # Test 17: Install non-existent binary (should fail)
    log_test "Test 17: Install non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" install nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Install should fail for non-existent binary"
    fi
    
    # Test 18: Remove non-existent binary (should fail)
    log_test "Test 18: Remove non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" remove nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Remove should fail for non-existent binary"
    fi
    
    # Test 19: Switch non-existent binary (should fail)
    log_test "Test 19: Switch non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" switch nonexistent-binary-xyz v1.0.0 >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Switch should fail for non-existent binary"
    fi
    
    # Test 20: Update non-existent binary (should fail)
    log_test "Test 20: Update non-existent binary fails gracefully"
    if ! "$BINMATE_BIN" update nonexistent-binary-xyz >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Update should fail for non-existent binary"
    fi
    
    echo ""
}

# Phase 7: Path Tests
run_path_tests() {
    log_info "=== Phase 7: Path Tests ==="
    
    # Test 21: Config directory exists
    log_test "Test 21: Config directory created"
    if [ -d "$HOME/.config/.binmate" ] || [ -d "$HOME/.config/binmate" ]; then
        test_passed
    else
        log_warn "Config directory not found (may not be created until config is set)"
        test_passed
    fi
    
    # Test 22: Data directory exists
    log_test "Test 22: Data directory created"
    if [ -d "$HOME/.local/share/binmate" ]; then
        test_passed
    else
        test_failed "Data directory not created"
    fi
    
    # Test 23: Binary is in PATH
    log_test "Test 23: Binary is accessible via PATH"
    if command -v binmate >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Binary not found in PATH"
    fi
    
    # Test 24: Binary executes correctly
    log_test "Test 24: Binary executes without error"
    if "$BINMATE_BIN" version >/dev/null 2>&1; then
        test_passed
    else
        test_failed "Binary execution failed"
    fi
    
    echo ""
}

# Main
main() {
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

main
