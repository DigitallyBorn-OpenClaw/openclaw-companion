#!/usr/bin/env sh

set -eu

DEFAULT_REPO="DigitallyBorn/oc-companion"
REPO="${OC_COMPANION_REPO:-$DEFAULT_REPO}"
INSTALL_DIR="${OC_COMPANION_INSTALL_DIR:-/usr/local/bin}"
VERSION="${OC_COMPANION_VERSION:-latest}"

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_asset() {
  os="$(uname -s)"
  arch="$(uname -m)"

  [ "$os" = "Linux" ] || fail "unsupported OS: $os (Linux required)"

  case "$arch" in
    x86_64|amd64)
      printf 'oc-companion\n'
      ;;
    aarch64|arm64)
      printf 'oc-companion-linux-arm64\n'
      ;;
    *)
      fail "unsupported architecture: $arch"
      ;;
  esac
}

build_download_url() {
  asset="$1"

  if [ "$VERSION" = "latest" ]; then
    printf 'https://github.com/%s/releases/latest/download/%s\n' "$REPO" "$asset"
    return
  fi

  printf 'https://github.com/%s/releases/download/%s/%s\n' "$REPO" "$VERSION" "$asset"
}

install_binary() {
  asset="$1"
  download_url="$2"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT INT TERM
  tmp_file="$tmp_dir/$asset"

  log "Downloading $asset from $download_url"
  curl -fsSL "$download_url" -o "$tmp_file" || fail "failed to download release asset"
  chmod +x "$tmp_file"

  target="$INSTALL_DIR/oc-companion"

  if [ -w "$INSTALL_DIR" ]; then
    install -m 0755 "$tmp_file" "$target"
  else
    require_cmd sudo
    sudo install -m 0755 "$tmp_file" "$target"
  fi

  log "Installed to $target"
  "$target" --help >/dev/null 2>&1 || true
  log "Run: oc-companion --help"
}

main() {
  require_cmd uname
  require_cmd curl
  require_cmd mktemp
  require_cmd install

  asset="$(detect_asset)"
  download_url="$(build_download_url "$asset")"

  install_binary "$asset" "$download_url"
}

main "$@"
