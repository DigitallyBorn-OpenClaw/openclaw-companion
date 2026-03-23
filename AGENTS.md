This repository contains `oc-companion`, the companion service for OpenClaw.

# Scope and Purpose

- Scope is `oc-companion/main/`.
- This repository focuses on companion service implementation, local IPC protocol behavior, and provider integration logic.
- It does not own the overall AMI deployment objective; that is a workspace/system-level concern led by gateway infrastructure.

# Runtime Role

- `oc-companion` runs as a separate process under a dedicated Linux user: `oc-companion`.
- It exposes scoped MCP-tool-like capabilities over a local Unix domain socket.
- It owns integrations and credentials needed to serve those capabilities.
- It may emit asynchronous events/callbacks to gateway components when required.

# Security Boundary

- Treat the gateway as an unprivileged client of companion endpoints.
- Do not require or assume direct gateway access to companion config/secrets on disk.
- Keep socket permissions and process ownership consistent with Linux user isolation.
- Minimize exposed method surface and return only data required for each tool call.

# Protocol and Implementation Guidance

- Keep protocol behavior explicit and stable (discovery, method invocation, error semantics).
- Favor backward-compatible protocol changes where possible.
- Keep service startup, logging, and configuration deterministic and auditable.
- Update interface documentation whenever endpoints, payloads, or event contracts change.

# Documentation Hygiene

- Keep this file and related docs safe for public consumption.
- Do not include secrets, credentials, private tenant identifiers, or environment-specific confidential details.
