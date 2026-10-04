#!/usr/bin/env sh

set -e

INSTALL_DIR="/usr/local/bin"

if [ -f "$INSTALL_DIR/bonjour" ]; then
    sudo rm "$INSTALL_DIR/bonjour"
    echo "Removed $INSTALL_DIR/bonjour"
else
    echo "bonjour binary not found in $INSTALL_DIR"
fi