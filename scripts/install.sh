#!/us/bin/env sh

# Exit if a command fails
set -e

# Variables: Repo owner, repo name, installation directory
OWNER="inestrivino"
REPO="bonjour"
INSTALL_DIR="/usr/local/bin"
# We determine the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# We check that the architecture is supported by the project
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Obtain the repo's latest tag (software's latest version)
LATEST_TAG=$(curl -s https://api.github.com/repos/$OWNER/$REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_TAG#v}

# URL of the latest release using the tag and computer architecture
URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST_TAG/${REPO}_${VERSION}_${OS}_${ARCH}.tar.gz"

# Using the release's URL, download it, decompress into the installation directory and then delete it from the tmp folder
echo "Downloading $REPO v$VERSION for $OS/$ARCH..."
curl -fsSL "$URL" -o /tmp/bonjour.tar.gz
tar -xzf /tmp/bonjour.tar.gz -C /tmp
sudo mv /tmp/bonjour "$INSTALL_DIR/bonjour"
rm /tmp/bonjour.tar.gz

# If all went well, inform the user
echo "Successfully installed $REPO to $INSTALL_DIR/bonjour"