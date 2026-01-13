#!/bin/sh
# fluxid installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/steinmann321/fluxid-cli/main/install.sh | sh
# Usage: curl -fsSL https://raw.githubusercontent.com/steinmann321/fluxid-cli/main/install.sh | sh -s -- v0.1.4

set -e

# Configuration
REPO="steinmann321/fluxid-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="fluxid"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version (default to latest if not specified)
VERSION="${1:-latest}"

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Darwin*)
            OS="darwin"
            ;;
        Linux*)
            OS="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            OS="windows"
            ;;
        *)
            echo "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    echo "${GREEN}Detected platform: $OS-$ARCH${NC}"
}

# Check prerequisites
check_prerequisites() {
    echo "${YELLOW}Checking prerequisites...${NC}"

    # Check if Claude CLI is installed
    if ! command -v claude >/dev/null 2>&1; then
        echo "${RED}Error: Claude CLI is not installed${NC}"
        echo "${RED}fluxid requires Claude CLI to function${NC}"
        echo ""
        echo "Install Claude CLI from: https://github.com/anthropics/claude-cli"
        exit 1
    fi

    # Test if Claude CLI is working
    if ! claude -p "say hi" >/dev/null 2>&1; then
        echo "${RED}Error: Claude CLI is installed but not working${NC}"
        echo "${RED}Please ensure Claude CLI is properly configured${NC}"
        echo ""
        echo "Test with: claude -p \"say hi\""
        exit 1
    fi

    echo "${GREEN}✓ Claude CLI is installed and working${NC}"
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        echo "${YELLOW}Fetching latest version...${NC}"
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo "${RED}Error: Could not fetch latest version${NC}"
            exit 1
        fi
        echo "${GREEN}Latest version: $VERSION${NC}"
    else
        echo "${GREEN}Installing version: $VERSION${NC}"
    fi
}

# Download binary
download_binary() {
    BINARY_EXT=""
    if [ "$OS" = "windows" ]; then
        BINARY_EXT=".exe"
    fi

    BINARY_FILENAME="${BINARY_NAME}-${OS}-${ARCH}${BINARY_EXT}"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_FILENAME"

    echo "${YELLOW}Downloading from: $DOWNLOAD_URL${NC}"

    TEMP_DIR=$(mktemp -d)
    TEMP_FILE="$TEMP_DIR/$BINARY_NAME"

    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo "${RED}Error: Failed to download binary${NC}"
        echo "${RED}URL: $DOWNLOAD_URL${NC}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    chmod +x "$TEMP_FILE"
    echo "${GREEN}Download complete${NC}"
}

# Install binary
install_binary() {
    echo "${YELLOW}Installing to $INSTALL_DIR/$BINARY_NAME...${NC}"

    # Check if install directory is writable
    if [ ! -w "$INSTALL_DIR" ]; then
        echo "${YELLOW}$INSTALL_DIR requires sudo access${NC}"
        if ! sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"; then
            echo "${RED}Error: Failed to install binary${NC}"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
    else
        if ! mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"; then
            echo "${RED}Error: Failed to install binary${NC}"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
    fi

    rm -rf "$TEMP_DIR"
    echo "${GREEN}✓ Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME${NC}"
}

# Verify installation
verify_installation() {
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        echo "${YELLOW}Warning: $BINARY_NAME is not in your PATH${NC}"
        echo "${YELLOW}Add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/$BINARY_NAME${NC}"
    else
        INSTALLED_VERSION=$("$BINARY_NAME" version 2>/dev/null | head -n1 || echo "unknown")
        echo "${GREEN}✓ Installation verified${NC}"
        echo "${GREEN}$INSTALLED_VERSION${NC}"
    fi
}

# Main installation flow
main() {
    echo "${GREEN}Installing fluxid...${NC}"
    detect_platform
    check_prerequisites
    get_latest_version
    download_binary
    install_binary
    verify_installation
    echo "${GREEN}✓ Installation complete!${NC}"
    echo ""
    echo "Get started:"
    echo "  fluxid init"
    echo "  fluxid --claude --file=/path/to/task.md"
}

main
