#!/usr/bin/env bash
#
# binmate installer
# 
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.sh | bash
#
# Environment variables:
#   BINMATE_VERSION     - Specific version to install (e.g., "v1.0.0" or "latest", default: "latest")
#   BINMATE_INSTALL_DIR - Installation directory (default: "/usr/local/bin")
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="cturner8/copilot-cli-challenge"
BINARY_NAME="binmate"
VERSION="${BINMATE_VERSION:-latest}"
INSTALL_DIR="${BINMATE_INSTALL_DIR:-/usr/local/bin}"

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

detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        Windows*)    os="windows" ;;
        *)          
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     arch="amd64" ;;
        amd64)      arch="amd64" ;;
        aarch64)    arch="arm64" ;;
        arm64)      arch="arm64" ;;
        *)          
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

get_latest_version() {
    log_info "Fetching latest version..."
    local latest
    latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$latest" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    
    echo "$latest"
}

validate_version() {
    local version=$1
    
    # Check if version exists
    local status_code
    status_code=$(curl -o /dev/null -s -w "%{http_code}" "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${version}")
    
    if [ "$status_code" != "200" ]; then
        log_error "Version ${version} not found in releases"
        exit 1
    fi
}

download_and_install() {
    local version=$1
    local platform=$2
    local archive_name="${BINARY_NAME}_${version#v}_${platform}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive_name}"
    local checksum_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/checksums.txt"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    
    log_info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    
    # Download archive
    if ! curl -fsSL "${download_url}" -o "${tmp_dir}/${archive_name}"; then
        log_error "Failed to download ${archive_name}"
        log_error "URL: ${download_url}"
        rm -rf "${tmp_dir}"
        exit 1
    fi
    
    # Download checksums
    log_info "Downloading checksums..."
    if ! curl -fsSL "${checksum_url}" -o "${tmp_dir}/checksums.txt"; then
        log_warn "Failed to download checksums, skipping verification"
    else
        # Verify checksum
        log_info "Verifying checksum..."
        cd "${tmp_dir}"
        if command -v sha256sum >/dev/null 2>&1; then
            if ! grep "${archive_name}" checksums.txt | sha256sum -c --status; then
                log_error "Checksum verification failed"
                rm -rf "${tmp_dir}"
                exit 1
            fi
        elif command -v shasum >/dev/null 2>&1; then
            if ! grep "${archive_name}" checksums.txt | shasum -a 256 -c --status; then
                log_error "Checksum verification failed"
                rm -rf "${tmp_dir}"
                exit 1
            fi
        else
            log_warn "No checksum utility found, skipping verification"
        fi
        cd - >/dev/null
    fi
    
    # Extract archive
    log_info "Extracting archive..."
    if ! tar -xzf "${tmp_dir}/${archive_name}" -C "${tmp_dir}"; then
        log_error "Failed to extract archive"
        rm -rf "${tmp_dir}"
        exit 1
    fi
    
    # Install binary
    log_info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    
    # Create install directory if it doesn't exist
    if [ ! -d "${INSTALL_DIR}" ]; then
        if ! mkdir -p "${INSTALL_DIR}"; then
            log_error "Failed to create directory ${INSTALL_DIR}"
            log_error "Try running with sudo or set BINMATE_INSTALL_DIR to a writable location"
            rm -rf "${tmp_dir}"
            exit 1
        fi
    fi
    
    # Move binary to install directory
    if ! mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"; then
        log_error "Failed to install binary to ${INSTALL_DIR}"
        log_error "Try running with sudo or set BINMATE_INSTALL_DIR to a writable location"
        rm -rf "${tmp_dir}"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # Cleanup
    rm -rf "${tmp_dir}"
}

main() {
    log_info "binmate installer"
    echo
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: ${platform}"
    
    # Determine version
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
        log_info "Latest version: ${VERSION}"
    else
        log_info "Installing version: ${VERSION}"
        validate_version "$VERSION"
    fi
    
    # Download and install
    download_and_install "$VERSION" "$platform"
    
    echo
    log_info "Successfully installed ${BINARY_NAME} ${VERSION} to ${INSTALL_DIR}/${BINARY_NAME}"
    echo
    log_info "Run '${BINARY_NAME} --help' to get started"
    
    # Check if install directory is in PATH
    if ! echo "$PATH" | grep -q "${INSTALL_DIR}"; then
        echo
        log_warn "${INSTALL_DIR} is not in your PATH"
        log_warn "Add it to your PATH by adding this to your shell profile:"
        echo "    export PATH=\"\${PATH}:${INSTALL_DIR}\""
    fi
}

main
