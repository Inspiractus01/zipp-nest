#!/bin/bash
set -e

REPO="Inspiractus01/zipp-nest"
BIN="zipp-nest"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH"; exit 1 ;;
esac

URL="https://github.com/${REPO}/releases/latest/download/${BIN}-${OS}-${ARCH}"

echo "installing zipp-nest (${OS}/${ARCH})..."
curl -L --fail --progress-bar "$URL" -o "/tmp/${BIN}" || {
  echo "download failed — check https://github.com/${REPO}/releases"
  exit 1
}
chmod +x "/tmp/${BIN}"
sudo mv "/tmp/${BIN}" "${INSTALL_DIR}/${BIN}"

echo "✓ installed to ${INSTALL_DIR}/${BIN}"
echo ""
echo "run:  zipp-nest serve"
