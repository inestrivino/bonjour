#!/usr/bin/env sh

# Exit if a command fails
set -e
# Installation directory
INSTALL_DIR="/usr/local/bin"

# If the file exists in the installation directory, then delete it
if [ -f "$INSTALL_DIR/bonjour" ]; then
    sudo rm "$INSTALL_DIR/bonjour"
    echo "Removed $INSTALL_DIR/bonjour"
else
    echo "bonjour binary not found in $INSTALL_DIR"
fi