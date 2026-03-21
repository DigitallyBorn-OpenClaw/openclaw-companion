# oc-companion

`oc-companion` is a companion service for OpenClaw.

It runs as a separate Linux process and user, exposes a narrow tool API over a Unix domain socket, and protects provider credentials by keeping integrations isolated from OpenClaw. It also serves as the event bridge that reports external changes back to OpenClaw through webhooks.

## Current Status
- Go service bootstrap is implemented.
- JSON socket protocol is implemented.
- Tool discovery endpoint is implemented.
- Initial tool methods are implemented with placeholder provider integrations:
  - `gmail.getMessage`
  - `calendar.listEvents`

## Documentation
- Usage guide: `docs/usage.md`
- Development guide: `docs/development.md`
- Architecture guide (with diagram): `docs/architecture.md`
- Socket protocol reference: `docs/socket-protocol.md`

## Make Targets
- `make build`: build native binary to `build/oc-companion`.
- `make build-arm64`: build Linux arm64 binary to `build/oc-companion-linux-arm64`.
- `make test`: run test suite.
- `make fmt`: format Go packages.

## GitHub Actions CI/CD
- Push and pull request workflows run `make test`, `make build`, and `make build-arm64`.
- Tag pushes matching `v*` create a GitHub Release.
- Release assets include `build/oc-companion` and `build/oc-companion-linux-arm64`.

## Install from GitHub Release
Use the install script (supports Linux `x86_64` and `arm64`, including Amazon Linux):

```bash
curl -fsSL https://raw.githubusercontent.com/DigitallyBorn/oc-companion/main/scripts/install.sh | sh
```

The script automatically:
- Detects CPU architecture.
- Downloads the matching release asset.
- Installs `oc-companion` to `/usr/local/bin/oc-companion` (uses `sudo` when needed).

Optional environment overrides:

```bash
# Install a specific release tag instead of latest.
curl -fsSL https://raw.githubusercontent.com/DigitallyBorn/oc-companion/main/scripts/install.sh | \
  OC_COMPANION_VERSION=v0.1.0 sh

# Install to a different directory.
curl -fsSL https://raw.githubusercontent.com/DigitallyBorn/oc-companion/main/scripts/install.sh | \
  OC_COMPANION_INSTALL_DIR="$HOME/.local/bin" sh

# Use a different GitHub repo (owner/repo).
curl -fsSL https://raw.githubusercontent.com/DigitallyBorn/oc-companion/main/scripts/install.sh | \
  OC_COMPANION_REPO="owner/repo" sh
```

## Quick Start
1. Ensure Go is installed.
2. Set required environment variables.
3. Run the service.

Example:

```bash
export OC_OPENCLAW_WEBHOOK_BASE_URL="http://127.0.0.1:8080/webhooks"
go run ./cmd/oc-companion
```

Then connect to the configured Unix socket and call `system.discover` first to enumerate available tools and metadata.
