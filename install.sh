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

ASSET="${BIN}-${OS}-${ARCH}"
BASE="https://github.com/${REPO}/releases/latest/download"

echo "installing zipp-nest (${OS}/${ARCH})..."

TMPDL=$(mktemp -d)
trap 'rm -rf "$TMPDL"' EXIT

curl -L --fail --progress-bar "$BASE/$ASSET" -o "$TMPDL/$ASSET" || {
  echo "download failed — check https://github.com/${REPO}/releases"
  exit 1
}

# verify checksum when the release ships one
if curl -sL --fail "$BASE/checksums.txt" -o "$TMPDL/checksums.txt" 2>/dev/null; then
  EXPECTED=$(awk -v a="$ASSET" '$2 == a {print $1}' "$TMPDL/checksums.txt")
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "$TMPDL/$ASSET" | awk '{print $1}')
  else
    ACTUAL=$(shasum -a 256 "$TMPDL/$ASSET" | awk '{print $1}')
  fi
  if [ -z "$EXPECTED" ] || [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "✗ checksum verification FAILED — refusing to install"
    echo "  expected: ${EXPECTED:-<missing>}"
    echo "  actual:   $ACTUAL"
    exit 1
  fi
  echo "✓ checksum verified"
else
  echo "! release has no checksums.txt — skipping verification"
fi

chmod +x "$TMPDL/$ASSET"
sudo mv "$TMPDL/$ASSET" "${INSTALL_DIR}/${BIN}"

echo "✓ installed to ${INSTALL_DIR}/${BIN}"
echo ""
echo "run:  zipp-nest"
