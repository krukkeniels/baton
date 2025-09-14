#!/bin/bash
set -e

# Baton CLI Orchestrator Installation Script
# Similar to Claude Code CLI installation

VERSION="latest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="baton"
GITHUB_REPO="race-day/baton"
TEMP_DIR=$(mktemp -d)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}==> ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os arch

    case "$OSTYPE" in
        linux*)   os="linux" ;;
        darwin*)  os="darwin" ;;
        msys*|cygwin*) os="windows" ;;
        *)        print_error "Unsupported operating system: $OSTYPE" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        armv7l) arch="arm" ;;
        *) print_error "Unsupported architecture: $(uname -m)" ;;
    esac

    PLATFORM="${os}-${arch}"
    if [[ "$os" == "windows" ]]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    print_status "Detected platform: $PLATFORM"
}

# Check if running as root for system install
check_permissions() {
    if [[ "$EUID" -eq 0 ]]; then
        print_warning "Running as root - installing system-wide"
        INSTALL_DIR="/usr/local/bin"
    elif [[ -w "/usr/local/bin" ]]; then
        print_status "Installing to /usr/local/bin (system-wide)"
    elif [[ -d "$HOME/.local/bin" ]]; then
        print_status "Installing to $HOME/.local/bin (user-only)"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    else
        print_status "Creating $HOME/.local/bin for user installation"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi
}

# Download binary from GitHub releases
download_binary() {
    local download_url

    if [[ "$VERSION" == "latest" ]]; then
        download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-${PLATFORM}"
    else
        download_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    fi

    print_status "Downloading from: $download_url"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$download_url" -o "${TEMP_DIR}/${BINARY_NAME}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "${TEMP_DIR}/${BINARY_NAME}"
    else
        print_error "Neither curl nor wget is available. Please install one of them."
    fi

    if [[ ! -f "${TEMP_DIR}/${BINARY_NAME}" ]]; then
        print_error "Failed to download baton binary"
    fi

    print_success "Downloaded baton binary"
}

# Verify binary
verify_binary() {
    chmod +x "${TEMP_DIR}/${BINARY_NAME}"

    # Basic verification - check if it's executable
    if ! "${TEMP_DIR}/${BINARY_NAME}" --version >/dev/null 2>&1; then
        print_warning "Binary verification failed, but continuing with installation"
    else
        print_success "Binary verification passed"
    fi
}

# Install binary
install_binary() {
    print_status "Installing to $INSTALL_DIR"

    if [[ -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
        print_warning "Existing installation found - backing up"
        mv "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/${BINARY_NAME}.backup"
    fi

    cp "${TEMP_DIR}/${BINARY_NAME}" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    print_success "Installed baton to $INSTALL_DIR"
}

# Update PATH if needed
update_path() {
    local shell_rc
    local path_updated=false

    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_warning "$INSTALL_DIR is not in your PATH"

        # Detect shell and update appropriate RC file
        case "$SHELL" in
            */bash)
                if [[ -f "$HOME/.bashrc" ]]; then
                    shell_rc="$HOME/.bashrc"
                elif [[ -f "$HOME/.bash_profile" ]]; then
                    shell_rc="$HOME/.bash_profile"
                fi
                ;;
            */zsh)
                shell_rc="$HOME/.zshrc"
                ;;
            */fish)
                # Fish uses a different syntax
                if command -v fish >/dev/null 2>&1; then
                    fish -c "set -U fish_user_paths $INSTALL_DIR \\$fish_user_paths"
                    path_updated=true
                fi
                ;;
        esac

        if [[ -n "$shell_rc" && ! "$path_updated" ]]; then
            echo "export PATH=\"$INSTALL_DIR:\\$PATH\"" >> "$shell_rc"
            print_success "Added $INSTALL_DIR to PATH in $shell_rc"
            print_warning "Please restart your terminal or run: source $shell_rc"
        fi
    fi
}

# Cleanup
cleanup() {
    rm -rf "$TEMP_DIR"
}

# Main installation process
main() {
    echo "🚀 Baton CLI Orchestrator Installation"
    echo "======================================"
    echo ""

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --version VERSION    Install specific version (default: latest)"
                echo "  --dir DIR           Install to specific directory"
                echo "  --help              Show this help message"
                echo ""
                echo "Examples:"
                echo "  curl -fsSL https://raw.githubusercontent.com/race-day/baton/main/install.sh | bash"
                echo "  curl -fsSL https://raw.githubusercontent.com/race-day/baton/main/install.sh | bash -s -- --version v1.0.0"
                echo ""
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                ;;
        esac
    done

    # Installation steps
    detect_platform
    check_permissions
    download_binary
    verify_binary
    install_binary
    update_path

    echo ""
    echo "🎉 Installation completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Open a new terminal or run: source ~/.bashrc (or ~/.zshrc)"
    echo "2. Verify installation: baton --version"
    echo "3. Get started: baton init"
    echo "4. Read the docs: baton --help"
    echo ""
    echo "Requirements for full functionality:"
    echo "• Claude Code CLI (for LLM integration)"
    echo "• Git (for version control features)"
    echo ""

    # Try to run baton --version
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        echo "✅ Installation verified:"
        "$BINARY_NAME" --version 2>/dev/null || echo "Baton installed successfully"
    else
        echo "⚠️  Please restart your terminal to use baton"
    fi
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"