# Usage Guide

## Purpose
`oc-companion` exposes a local Unix socket API for OpenClaw and keeps provider access isolated behind this boundary.

## Prerequisites
- Linux environment with Unix domain sockets.
- Go toolchain available for local runs.
- A reachable OpenClaw Gmail webhook URL.
- A GCP project with an existing Gmail Pub/Sub topic.

## Configuration
`oc-companion` reads configuration from environment variables.

Required:
- `OC_OPENCLAW_GMAIL_WEBHOOK_URL`
- `OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN`
- `OC_GCP_PROJECT_ID`
- `OC_GCP_GMAIL_PUBSUB_TOPIC_ID`

Optional:
- `OC_GCP_CREDENTIALS_FILE` (if omitted, Application Default Credentials are used)
- `OC_GCP_PUBSUB_SUBSCRIPTION_PREFIX` (default: `oc-companion-gmail`)
- `OC_COMPANION_SOCKET_PATH` (default: `/run/oc-companion/companion.sock`)
- `OC_COMPANION_LOG_LEVEL` (`debug`, `info`, `warn`, `error`; default: `info`)
- `OC_COMPANION_LOG_FORMAT` (`text` or `json`; default: `text`)
- `OC_COMPANION_SHUTDOWN_TIMEOUT` (duration like `10s` or integer seconds; default: `10s`)
- `OC_OPENCLAW_WEBHOOK_BASE_URL` (legacy alias for `OC_OPENCLAW_GMAIL_WEBHOOK_URL`)

## Running

```bash
export OC_OPENCLAW_GMAIL_WEBHOOK_URL="http://127.0.0.1:8080/hooks/gmail"
export OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN="replace-me"
export OC_GCP_PROJECT_ID="my-gcp-project"
export OC_GCP_GMAIL_PUBSUB_TOPIC_ID="gmail-notifications"
export OC_COMPANION_SOCKET_PATH="/tmp/oc-companion.sock"
go run ./cmd/oc-companion
```

## Client Interaction (Simplest Path)
1. Connect to the Unix socket.
2. Send `system.discover`.
3. Select a method from discovery output.
4. Send method request with valid params.
5. Reuse connection for additional calls.

See full protocol details in `docs/socket-protocol.md`.

## Example Requests
Discovery:

```json
{"id":"1","method":"system.discover"}
```

Ping:

```json
{"id":"2","method":"system.ping"}
```

Gmail message lookup (contract in place; provider integration pending):

```json
{"id":"3","method":"gmail.getMessage","params":{"message_id":"18c2b"}}
```

Calendar events lookup (contract in place; provider integration pending):

```json
{"id":"4","method":"calendar.listEvents","params":{"start":"2026-03-14T00:00:00Z","end":"2026-03-15T00:00:00Z","max_results":10}}
```

## Runtime Behavior Notes
- The socket directory is created if missing.
- Stale socket files are cleaned up when safe.
- Socket permissions are set to `0660`.
- Invalid JSON requests return a protocol parse error.
- A dedicated Pub/Sub subscription is created on startup and deleted during shutdown.
- Pub/Sub messages are acknowledged only after the OpenClaw webhook returns a `2xx` response.
