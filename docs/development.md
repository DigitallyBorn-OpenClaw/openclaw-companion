# Development Guide

## Project Layout
- `cmd/oc-companion`: executable entrypoint.
- `internal/app`: lifecycle and socket listener bootstrap.
- `internal/config`: environment configuration loading/validation.
- `internal/logging`: structured logger construction.
- `internal/protocol`: request/response and error envelope types.
- `internal/api`: method registry, dispatch, and connection serving.
- `internal/tools`: tool method registration and service contracts.
- `docs`: project documentation.

## Local Setup
1. Install Go.
2. Clone repo.
3. Set required env vars.

```bash
export OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN="replace-me"
export OC_GCP_PROJECT_ID="my-gcp-project"
export OC_GCP_GMAIL_PUBSUB_TOPIC_ID="gmail-notifications"
export OC_GCP_CREDENTIALS_FILE="/var/lib/oc-companion/gcp-credentials.json"
```

By default, the companion sends Gmail event callbacks to `http://127.0.0.1:18789/hooks/gmail`.
Those credentials must be valid for both Pub/Sub access and Gmail readonly message retrieval if you want `gmail.getMessage` to work locally.

## Build and Test
Build:

```bash
go build ./...
```

Test:

```bash
go test ./...
```

Format:

```bash
gofmt -w ./cmd ./internal
```

## Adding a New Tool Method
1. Define request/response behavior and simplest client interaction.
2. Add service contract types in `internal/tools`.
3. Register the method in `internal/tools/register.go`.
4. Add strong param validation and protocol-level errors.
5. Ensure discovery metadata includes usage and params.
6. Add tests for validation and success/error behavior.
7. Update `docs/socket-protocol.md` and `docs/usage.md`.

## Integration Pattern
- Keep external provider logic behind service interfaces.
- Keep tool handlers thin: parse/validate/map errors/return normalized output.
- Return deterministic structures so OpenClaw clients can rely on stable contracts.

## Deployment Notes (Current)
- Run as a dedicated Linux user.
- Keep socket path in a restricted directory.
- Grant socket access only to authorized local principals.
- Use webhook URL routing controlled by trusted local/network boundaries.
- Store the OpenClaw webhook token and GCP credentials file with `0600` permissions.
