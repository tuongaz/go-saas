#!/bin/bash
set -e

MOCKERY_VERSION="v2.53.3"
TOOLS_DIR=$(dirname "$0")
BIN_DIR="$TOOLS_DIR/bin"
MOCKERY_BIN="$BIN_DIR/mockery"

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

# Check if mockery is already installed with the correct version
if [ -x "$MOCKERY_BIN" ]; then
  INSTALLED_VERSION=$("$MOCKERY_BIN" --version 2>/dev/null | grep -o "v[0-9]*\.[0-9]*\.[0-9]*" || echo "")
  if [ "$INSTALLED_VERSION" = "$MOCKERY_VERSION" ]; then
    echo "mockery $MOCKERY_VERSION is already installed, skipping installation."
    exit 0
  else
    echo "Updating mockery from $INSTALLED_VERSION to $MOCKERY_VERSION..."
  fi
fi

# Determine OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Map OS to mockery's naming convention
case $OS in
  Darwin)
    OS="Darwin"
    ;;
  Linux)
    OS="Linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Map architecture to mockery's naming convention
case $ARCH in
  x86_64)
    ARCH="x86_64"
    ;;
  arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Construct URL for mockery binary
DOWNLOAD_URL="https://github.com/vektra/mockery/releases/download/${MOCKERY_VERSION}/mockery_2.53.3_${OS}_${ARCH}.tar.gz"
TEMP_DIR=$(mktemp -d)

echo "Downloading mockery ${MOCKERY_VERSION} for ${OS}_${ARCH}..."
curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/mockery.tar.gz"

if [ ! -s "$TEMP_DIR/mockery.tar.gz" ]; then
  echo "Error: Failed to download mockery, file is empty."
  exit 1
fi

echo "Extracting mockery..."
tar -xzf "$TEMP_DIR/mockery.tar.gz" -C "$TEMP_DIR"

echo "Installing mockery to $BIN_DIR/mockery..."
mv "$TEMP_DIR/mockery" "$BIN_DIR/"
chmod +x "$BIN_DIR/mockery"

# Clean up
rm -rf "$TEMP_DIR"

# Verify installation
if [ -x "$BIN_DIR/mockery" ]; then
  echo "mockery $MOCKERY_VERSION has been installed successfully!"
else
  echo "Error: Failed to install mockery."
  exit 1
fi 