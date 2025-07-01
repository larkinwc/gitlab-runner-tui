#!/bin/bash
set -e

# GitLab Runner TUI Installer Script

REPO="larkin/gitlab-runner-tui"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="gitlab-runner-tui"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="armv7"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    linux)
        OS="Linux"
        ;;
    darwin)
        OS="Darwin"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Get latest release
echo "Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to get latest release"
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"

# Construct download URL
FILENAME="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${FILENAME}"

echo "Downloading from: $DOWNLOAD_URL"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -L -o "$FILENAME" "$DOWNLOAD_URL"; then
    echo "Failed to download release"
    exit 1
fi

echo "Extracting..."
tar -xzf "$FILENAME"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd -
rm -rf "$TMP_DIR"

echo "✓ GitLab Runner TUI installed successfully!"
echo "  Location: $INSTALL_DIR/$BINARY_NAME"
echo "  Version: $LATEST_RELEASE"
echo ""
echo "Run 'gitlab-runner-tui' to start"