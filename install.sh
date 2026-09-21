#!/bin/bash
echo "==============================="
echo " Say Less Language Installer"
echo "==============================="
echo ""

INSTALL_DIR="$HOME/.sayless"
BIN_DIR="$INSTALL_DIR/bin"

# Detect platform
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Darwin)
        case "$ARCH" in
            arm64) BINARY="sale-macos-arm64" ;;
            *) BINARY="sale-macos-amd64" ;;
        esac
        ;;
    Linux)
        case "$ARCH" in
            aarch64|arm64) BINARY="sale-linux-arm64" ;;
            *) BINARY="sale-linux-amd64" ;;
        esac
        ;;
    *)
        echo "Unsupported platform: $OS"
        exit 1
        ;;
esac

# Check if binary exists in current directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ ! -f "$SCRIPT_DIR/$BINARY" ]; then
    echo "Error: $BINARY not found in current directory."
    echo "Make sure you run this script from the distribution folder."
    exit 1
fi

echo "Installing Say Less to $INSTALL_DIR..."

# Create installation directory
mkdir -p "$BIN_DIR"

# Copy binary
cp "$SCRIPT_DIR/$BINARY" "$BIN_DIR/sale"
chmod +x "$BIN_DIR/sale"

# Add to PATH
SHELL_RC=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_RC="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_RC="$HOME/.zshrc"
fi

if [ -n "$SHELL_RC" ]; then
    if ! grep -q '$HOME/.sayless/bin' "$SHELL_RC" 2>/dev/null; then
        echo '' >> "$SHELL_RC"
        echo '# Say Less' >> "$SHELL_RC"
        echo 'export PATH="$HOME/.sayless/bin:$PATH"' >> "$SHELL_RC"
        echo "Added to PATH in $SHELL_RC"
    else
        echo "PATH already configured in $SHELL_RC"
    fi
else
    echo ""
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$HOME/.sayless/bin:\$PATH\""
fi

echo ""
echo "==============================="
echo " Installation Complete!"
echo "==============================="
echo ""
echo "Start a new terminal or run:"
echo "  source ${SHELL_RC:-~/.bashrc}"
echo ""
echo "Quick start:"
echo "  sale new my-project"
echo "  cd my-project"
echo "  sale run src/main.sl"
echo ""
