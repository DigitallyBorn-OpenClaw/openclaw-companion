# Socket Protocol

`oc-companion` exposes a Unix domain socket using a simple JSON request/response stream.

## Client Interaction Flow (Minimal)
1. Connect to the configured socket path.
2. Send `system.discover` to enumerate available methods and required params.
3. Call a method by name with JSON params.
4. Read one JSON response per request.

## Request Shape
```json
{
  "id": "req-1",
  "method": "system.discover",
  "params": {}
}
```

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
- Purpose: enumerate method metadata (name, description, usage, and params schema hints).

## Baseline System Method
- Method: `system.ping`
- Purpose: liveness check.
