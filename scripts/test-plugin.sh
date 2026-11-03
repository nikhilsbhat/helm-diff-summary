#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
trap 'exit 1' HUP INT TERM

export HELM_PLUGINS="$temp_dir/data/plugins"
export HELM_CONFIG_HOME="$temp_dir/config"
export HELM_CACHE_HOME="$temp_dir/cache"
export HELM_DATA_HOME="$temp_dir/data"
export HELM_PLUGIN_DIR="$HELM_PLUGINS/diff-summary"
export TEST_PLUGIN_RELEASE="$temp_dir/release"
export TEST_PLUGIN_VERSION
TEST_PLUGIN_VERSION=$(awk -F '"' '/^version:/ {print $2; exit}' "$root/plugin.yaml")
mkdir -p "$HELM_PLUGINS" "$temp_dir/mock" "$TEST_PLUGIN_RELEASE" "$temp_dir/source"
cp "$root/scripts/testdata/plugin/curl" "$root/scripts/testdata/plugin/uname" "$temp_dir/mock/"
chmod +x "$temp_dir/mock/curl" "$temp_dir/mock/uname"

# Build once, then exercise the real download hook against local release fixtures.
go build -o "$temp_dir/source/helm-diff-summary" "$root"
export PATH="$temp_dir/mock:$PATH"
export TEST_PLUGIN_OS=Linux TEST_PLUGIN_ARCH=x86_64
for platform in linux_x86_64 darwin_arm64 freebsd_i386 linux_arm; do
  archive="helm-diff-summary_${TEST_PLUGIN_VERSION}_${platform}.tar.gz"
  tar -czf "$TEST_PLUGIN_RELEASE/$archive" -C "$temp_dir/source" helm-diff-summary
done
checksums="$TEST_PLUGIN_RELEASE/helm-diff-summary_${TEST_PLUGIN_VERSION}_SHA256SUMS"
for archive in "$TEST_PLUGIN_RELEASE/"*.tar.gz; do
  if command -v sha256sum >/dev/null 2>&1; then
    digest=$(sha256sum "$archive" | awk '{print $1}')
  else
    digest=$(shasum -a 256 "$archive" | awk '{print $1}')
  fi
  printf '%s  %s\n' "$digest" "${archive##*/}" >> "$checksums"
done

mkdir -p "$temp_dir/package/diff-summary"
cp "$root/plugin.yaml" "$root/install-binary.sh" "$temp_dir/package/diff-summary/"
helm plugin install "$temp_dir/package/diff-summary"
helm plugin list | grep 'diff-summary'
helm diff-summary --help | grep 'fail-on-delete'
helm diff-summary --version

printf 'default, api, Deployment (apps) has been added:\n+ kind: Deployment\n' > "$temp_dir/input"
cd "$temp_dir"
"$temp_dir/source/helm-diff-summary" -o json < input > standalone.json
helm diff-summary -o json < input > plugin.json
cmp standalone.json plugin.json

printf 'default, api, Deployment (apps) has been removed:\n- kind: Deployment\n' > input
set +e
helm diff-summary --fail-on-delete < input > /dev/null 2> error.log
code=$?
set -e
[ "$code" -eq 2 ] || { echo "Expected policy exit 2, got $code" >&2; exit 1; }

# The update hook also uses the manifest's pinned release and archive naming.
export TEST_PLUGIN_OS=Darwin TEST_PLUGIN_ARCH=aarch64
"$HELM_PLUGIN_DIR/install-binary.sh"
export TEST_PLUGIN_OS=FreeBSD TEST_PLUGIN_ARCH=i686
"$HELM_PLUGIN_DIR/install-binary.sh"
export TEST_PLUGIN_OS=Linux TEST_PLUGIN_ARCH=armv7l
"$HELM_PLUGIN_DIR/install-binary.sh"

# Failed updates must leave the previously working binary in place.
cp "$HELM_PLUGIN_DIR/bin/helm-diff-summary" previous-binary
printf 'corrupt archive' > "$TEST_PLUGIN_RELEASE/helm-diff-summary_${TEST_PLUGIN_VERSION}_linux_arm.tar.gz"
if "$HELM_PLUGIN_DIR/install-binary.sh" > error.log 2>&1; then
  echo 'Expected checksum rejection' >&2; exit 1
fi
grep 'Checksum verification failed' error.log
cmp previous-binary "$HELM_PLUGIN_DIR/bin/helm-diff-summary"
export TEST_DOWNLOAD_FAIL=1
if "$HELM_PLUGIN_DIR/install-binary.sh" > error.log 2>&1; then
  echo 'Expected download failure' >&2; exit 1
fi
unset TEST_DOWNLOAD_FAIL
export TEST_PLUGIN_OS=Unsupported
if "$HELM_PLUGIN_DIR/install-binary.sh" > error.log 2>&1; then
  echo 'Expected unsupported OS rejection' >&2; exit 1
fi
grep 'Unsupported plugin OS' error.log
export TEST_PLUGIN_OS=Linux TEST_PLUGIN_ARCH=unsupported
if "$HELM_PLUGIN_DIR/install-binary.sh" > error.log 2>&1; then
  echo 'Expected unsupported architecture rejection' >&2; exit 1
fi
grep 'Unsupported plugin architecture' error.log
helm plugin uninstall diff-summary
make -C "$root" local/install HELM_PLUGIN_PATH="$HELM_PLUGIN_DIR"
helm diff-summary -o json < input > plugin.json
"$temp_dir/source/helm-diff-summary" -o json < input > standalone.json
cmp standalone.json plugin.json
helm plugin uninstall diff-summary
echo 'Plugin install, invocation, policy exit, update hook, and failure checks passed.'
