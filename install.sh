#!/usr/bin/env bash

set -e

INSTALL_DIR="$HOME/.deecli/bin"
mkdir -p "$INSTALL_DIR"

# Assume deecli binary is in current directory
cp ./deecli "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/deecli"

SHELL_RC=""

if [ -n "$ZSH_VERSION" ]; then
  SHELL_RC="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
  SHELL_RC="$HOME/.bashrc"
else
  echo "Unsupported shell. Please add $INSTALL_DIR to your PATH manually."
  exit 1
fi

# Add to PATH if not already present
if ! grep -q "$INSTALL_DIR" "$SHELL_RC"; then
  echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
  echo "Added $INSTALL_DIR to PATH in $SHELL_RC"
else
  echo "$INSTALL_DIR already in PATH in $SHELL_RC"
fi

echo "Installation complete. Please restart your terminal or run:"
echo "  source $SHELL_RC"