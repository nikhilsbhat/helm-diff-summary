#!/bin/sh
set -eu

: "${HELM_PLUGIN_DIR:?HELM_PLUGIN_DIR must point to the plugin directory}"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  FreeBSD) os=freebsd ;;
  *) echo "Unsupported plugin OS: $(uname -s). Use the standalone binary instead." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=x86_64 ;;
  aarch64|arm64) arch=arm64 ;;
  i386|i486|i586|i686) arch=i386 ;;
  armv6l|armv7l) arch=arm ;;
  *) echo "Unsupported plugin architecture: $(uname -m)" >&2; exit 1 ;;
esac

# Keep the download pinned to the installed plugin checkout, including updates.
version=$(awk -F '"' '/^version:/ {print $2; exit}' "$HELM_PLUGIN_DIR/plugin.yaml")
if [ -z "$version" ]; then
  echo "Missing version in plugin.yaml" >&2
  exit 1
fi

archive="helm-diff-summary_${version}_${os}_${arch}.tar.gz"
checksums="helm-diff-summary_${version}_SHA256SUMS"
release="https://github.com/nikhilsbhat/helm-diff-summary/releases/download/v${version}"
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
trap 'exit 1' HUP INT TERM

echo "Installing helm diff-summary ${version} for ${os}/${arch}..."
curl --fail --silent --show-error --location "$release/$archive" -o "$temp_dir/$archive"
curl --fail --silent --show-error --location "$release/$checksums" -o "$temp_dir/$checksums"

expected=$(awk -v archive="$archive" '$2 == archive {print $1}' "$temp_dir/$checksums")
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$temp_dir/$archive" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$temp_dir/$archive" | awk '{print $1}')
else
  echo "Need sha256sum or shasum to verify the release." >&2
  exit 1
fi
if [ -z "$expected" ] || [ "$actual" != "$expected" ]; then
  echo "Checksum verification failed for $archive" >&2
  exit 1
fi

tar -xzf "$temp_dir/$archive" -C "$temp_dir" helm-diff-summary
chmod +x "$temp_dir/helm-diff-summary"
mkdir -p "$HELM_PLUGIN_DIR/bin"
mv "$temp_dir/helm-diff-summary" "$HELM_PLUGIN_DIR/bin/helm-diff-summary"
echo "Installed. Pipe helm diff output into helm diff-summary."
