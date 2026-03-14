# Socket Protocol

`oc-companion` exposes a Unix domain socket using a simple JSON request/response stream.

## Client Interaction Flow (Minimal)
1. Connect to the configured socket path.
2. Send `system.discover` to enumerate available methods and required params.
3. Call a method by name with JSON params from discovery metadata.
4. Read one JSON response per request.
5. Reuse the same socket connection for additional requests.

## Request Shape
```json
{
  "id": "req-1",
  "method": "system.discover",
  "params": {}
}
```

`id` may be a string or number and is echoed in the response.

## Response Shape
Success:
```json
{
  "id": "req-1",
  "result": {
    "service": "oc-companion"
  }
}
```

Error:
```json
{
  "id": "req-1",
  "error": {
    "code": -32601,
    "message": "method not found"
  }
}
```

## Discovery Method
- Method: `system.discover`
- Purpose: enumerate method metadata (name, description, usage, and params schema hints), plus event-delivery metadata.

## Initial Tool Methods
- Method: `gmail.getMessage`
  - Required params: `message_id`
  - Returns normalized message fields (`id`, `thread_id`, `from`, `to`, `subject`, `snippet`, `received_at`).

- Method: `calendar.listEvents`
  - Required params: `start`, `end` (RFC3339)
  - Optional params: `calendar_id` (default `primary`), `max_results` (default `20`, max `100`)
  - Returns `events` plus resolved query window metadata.

## Example Sequence
`system.discover` request:
```json
{"id":"1","method":"system.discover"}
```

`gmail.getMessage` request:
```json
{"id":"2","method":"gmail.getMessage","params":{"message_id":"18c2b"}}
```

`calendar.listEvents` request:
```json
{"id":"3","method":"calendar.listEvents","params":{"start":"2026-03-14T00:00:00Z","end":"2026-03-15T00:00:00Z","max_results":10}}
```

## Baseline System Method
- Method: `system.ping`
- Purpose: liveness check.

## Event Metadata
Discovery also returns event metadata for webhook-delivered asynchronous notifications, including the initial `gmail.new_message` event contract.
