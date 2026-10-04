#!/us/bin/env sh
set -e

OWNER="inestrivino"
REPO="bonjour"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

LATEST_TAG=$(curl -s https://api.github.com/repos/$OWNER/$REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_TAG#v}

URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST_TAG/${REPO}_${VERSION}_${OS}_${ARCH}.tar.gz"

echo "Downloading $REPO v$VERSION for $OS/$ARCH..."
curl -fsSL "$URL" -o /tmp/bonjour.tar.gz
tar -xzf /tmp/bonjour.tar.gz -C /tmp
sudo mv /tmp/bonjour "$INSTALL_DIR/bonjour"
rm /tmp/bonjour.tar.gz

echo "Successfully installed $REPO to $INSTALL_DIR/bonjour"