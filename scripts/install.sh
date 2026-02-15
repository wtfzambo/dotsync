#!/usr/bin/env bash
#
# dotsync installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/wtfzambo/dotsync/main/scripts/install.sh | bash
#
# ⚠️ IMPORTANT: This script must be EXECUTED, never SOURCED
# ❌ WRONG: source install.sh (will exit your shell on errors)
# ✅ CORRECT: bash install.sh
# ✅ CORRECT: curl -fsSL ... | bash
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

REPO="wtfzambo/dotsync"
BINARY_NAME="dotsync"

log_info() {
    echo -e "${BLUE}==>${NC} $1"
}

log_success() {
    echo -e "${GREEN}==>${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}==>${NC} $1"
}

log_error() {
    echo -e "${RED}Error:${NC} $1" >&2
}

# Detect OS and architecture
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Darwin)
            os="Darwin"
            ;;
        Linux)
            os="Linux"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)
            arch="x86_64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

# Download and install from GitHub releases
install_from_release() {
    log_info "Installing ${BINARY_NAME} from GitHub releases..."

    local platform=$1
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Get latest release version
    log_info "Fetching latest release..."
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
    local version
    local release_json

    if command -v curl &> /dev/null; then
        release_json=$(curl -fsSL "$latest_url")
    elif command -v wget &> /dev/null; then
        release_json=$(wget -qO- "$latest_url")
    else
        log_error "Neither curl nor wget found. Please install one of them."
        return 1
    fi

    version=$(echo "$release_json" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        log_error "Failed to fetch latest version"
        return 1
    fi

    log_info "Latest version: $version"

    # Construct download URL
    local archive_name="${BINARY_NAME}_${version}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

    log_info "Downloading ${archive_name}..."
    
    cd "$tmp_dir"
    if command -v curl &> /dev/null; then
        if ! curl -fsSL -o "$archive_name" "$download_url"; then
            log_warning "Download failed, trying fallback installation method..."
            cd - > /dev/null || cd "$HOME"
            return 1
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q -O "$archive_name" "$download_url"; then
            log_warning "Download failed, trying fallback installation method..."
            cd - > /dev/null || cd "$HOME"
            return 1
        fi
    fi

    # Extract archive
    log_info "Extracting archive..."
    if ! tar -xzf "$archive_name"; then
        log_error "Failed to extract archive"
        return 1
    fi

    # Determine install location
    local install_dir
    if [[ -w /usr/local/bin ]]; then
        install_dir="/usr/local/bin"
    else
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
    fi

    # Install binary
    log_info "Installing to ${install_dir}..."
    if [[ -w "$install_dir" ]]; then
        mv "$BINARY_NAME" "$install_dir/"
    else
        sudo mv "$BINARY_NAME" "$install_dir/"
    fi
    chmod +x "$install_dir/$BINARY_NAME"

    log_success "${BINARY_NAME} installed to ${install_dir}/${BINARY_NAME}"

    # Check if install_dir is in PATH
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        log_warning "$install_dir is not in your PATH"
        echo ""
        echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"\$PATH:$install_dir\""
        echo ""
    fi

    cd - > /dev/null || cd "$HOME"
    return 0
}

# Check if Go is installed and meets minimum version
check_go() {
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go detected: $(go version)"

        # Extract major and minor version numbers
        local major=$(echo "$go_version" | cut -d. -f1)
        local minor=$(echo "$go_version" | cut -d. -f2)

        # Check if Go version is 1.25 or later
        if [ "$major" -eq 1 ] && [ "$minor" -lt 25 ]; then
            log_error "Go 1.25 or later is required (found: $go_version)"
            echo ""
            echo "Please upgrade Go:"
            echo "  - Download from https://go.dev/dl/"
            echo "  - Or use your package manager to update"
            echo ""
            return 1
        fi

        return 0
    else
        return 1
    fi
}

# Install using go install (fallback)
install_with_go() {
    log_info "Installing ${BINARY_NAME} using 'go install'..."

    if go install "github.com/${REPO}/cmd/${BINARY_NAME}@latest"; then
        log_success "${BINARY_NAME} installed successfully via go install"

        # Determine where Go installed the binary
        local bin_dir
        local gobin
        gobin=$(go env GOBIN 2>/dev/null || true)
        if [ -n "$gobin" ]; then
            bin_dir="$gobin"
        else
            bin_dir="$(go env GOPATH)/bin"
        fi

        # Check if GOPATH/bin (or GOBIN) is in PATH
        if [[ ":$PATH:" != *":$bin_dir:"* ]]; then
            log_warning "$bin_dir is not in your PATH"
            echo ""
            echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
            echo "  export PATH=\"\$PATH:$bin_dir\""
            echo ""
        fi

        return 0
    else
        log_error "go install failed"
        return 1
    fi
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        log_success "${BINARY_NAME} is installed and ready!"
        echo ""
        "$BINARY_NAME" --version
        echo ""
        echo "Get started:"
        echo "  dotsync --help"
        echo ""
        return 0
    else
        log_error "${BINARY_NAME} was installed but is not in PATH"
        return 1
    fi
}

# Main installation flow
main() {
    echo ""
    echo "🔗 dotsync Installer"
    echo ""

    log_info "Detecting platform..."
    local platform
    platform=$(detect_platform)
    log_info "Platform: $platform"

    # Try downloading from GitHub releases first
    if install_from_release "$platform"; then
        verify_installation
        exit 0
    fi

    log_warning "Failed to install from releases, trying fallback method..."

    # Try go install as fallback
    if check_go; then
        if install_with_go; then
            verify_installation
            exit 0
        fi
    fi

    # All methods failed
    log_error "Installation failed"
    echo ""
    echo "Manual installation:"
    echo "  1. Download from https://github.com/${REPO}/releases/latest"
    echo "  2. Extract and move '${BINARY_NAME}' to your PATH"
    echo ""
    echo "Or install from source:"
    echo "  1. Install Go 1.25+ from https://go.dev/dl/"
    echo "  2. Run: go install github.com/${REPO}/cmd/${BINARY_NAME}@latest"
    echo ""
    exit 1
}

main "$@"
