#!/bin/sh
set -eu

REPO="sylophi/terrier"
DEST="${TERRIER_INSTALL_DIR:-$HOME/.local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64) ARCH=arm64 ;;
  x86_64) ARCH=x64 ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="terrier-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

mkdir -p "$DEST"
# Staged inside the destination so the install is a same-filesystem
# rename. Downloading to $TMPDIR would make the final step a copy over the
# live binary, which an interruption could leave truncated.
TMP=$(mktemp "$DEST/.terrier.XXXXXX")
trap 'rm -f "$TMP"' EXIT

echo "Downloading $URL..." >&2
curl -fsSL "$URL" -o "$TMP"
chmod 755 "$TMP"
mv "$TMP" "$DEST/terrier"
echo "Installed terrier to $DEST/terrier" >&2

# `ter` is the same binary under a shorter name. A symlink rather than a
# copy, so `terrier update` moves both at once. Anything already sitting
# at that name belongs to something else and is left alone.
if [ -e "$DEST/ter" ] && [ "$(readlink "$DEST/ter" || true)" != "terrier" ]; then
  echo "Note: $DEST/ter already exists and is not terrier's alias. Leaving it alone." >&2
else
  ln -sf terrier "$DEST/ter"
  echo "Installed alias $DEST/ter" >&2
fi

# There is nothing to configure: the registry is created on the first
# `terrier add`.
CONFIG_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/terrier/projects.json"
if [ -f "$CONFIG_FILE" ]; then
  echo "Using the existing registry at $CONFIG_FILE" >&2
else
  echo "Projects will be registered in $CONFIG_FILE" >&2
fi

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "Note: $DEST is not in \$PATH. Add it to your shell profile to use terrier." >&2 ;;
esac
