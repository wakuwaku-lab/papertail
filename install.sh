#!/bin/bash

set -e

REPO="wakuwaku-lab/papertail"

get_latest_tag() {
    curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/'
}

TAG=$(get_latest_tag)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

if [ "$OS" = "windows" ]; then
    EXT=".exe"
else
    EXT=""
fi

FILENAME="papertail-${TAG}-${OS}-${ARCH}${EXT}"
URL="https://github.com/${REPO}/releases/download/v${TAG}/${FILENAME}"

echo "Downloading ${FILENAME}..."
curl -fSL "$URL" -o "papertail${EXT}"
chmod +x "papertail${EXT}"

echo "Done! Run ./papertail${EXT} to start."
