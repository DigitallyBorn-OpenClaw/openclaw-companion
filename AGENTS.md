# AGENTS Guide

## Purpose
This repository is for **oc-companion**, a companion process for OpenClaw.

## Project Direction (Living Summary)
- `oc-companion` runs as its own Linux process and separate Linux user.
- It exposes carefully scoped capabilities through domain sockets for OpenClaw.
- It protects secrets by avoiding direct OpenClaw access to underlying providers/services.
- It sends event notifications back to OpenClaw through webhooks.

## Current Architecture Decisions
- IPC boundary: Unix domain sockets between OpenClaw and `oc-companion`.
- Privilege boundary: dedicated Linux user for this process.
- Provider integrations are owned by `oc-companion` and not directly exposed.
- OpenClaw receives asynchronous events through webhook callbacks.
- Primary implementation language: Go.
- Service bootstrap includes environment-based config and structured `slog` logging.

## Initial Tooling Scope
- Notify OpenClaw when a new Gmail message is received (via GCP-hosted Pub/Sub topic).
- Provide OpenClaw with requested Gmail message content.
- Provide OpenClaw with requested Google Calendar events.

## Maintenance Rules
- Keep this file updated for **large decisions** and **project direction changes**.
- Keep this file concise and durable (high-level decisions, not day-to-day logs).
- Put short-term or transient details into `./.notes/YYYY-MM-DD.md`.
- Track only pending and in-progress work in `TODO.md`.

## Agent Startup Context Rule
- At startup, an agent should read the two most recent files in `./.notes/` to quickly recover recent project context and progress.
