# TODO

## In Progress
- Define initial service skeleton (startup, config loading, structured logging).

## Pending
- Define Unix domain socket API contract for OpenClaw tool requests.
- Design authn/authz and file-permission model for socket access.
- Implement Gmail Pub/Sub consumer for new message notifications.
- Implement OpenClaw webhook notifier with retry/backoff strategy.
- Implement Gmail message retrieval tool endpoint.
- Implement Google Calendar events retrieval tool endpoint.
- Add provider credential management strategy (least privilege, rotation path).
- Add operational docs for deployment as separate Linux user/process.
