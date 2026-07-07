# Event Bus Go

A lightweight Go event bus server with HTTP endpoints for registering event types, publishing events, and streaming events over Server-Sent Events (SSE).

## API Routes

| Method | Route | Description |
| --- | --- | --- |
| GET | /health | Health check endpoint returning service status and version. |
| POST | /events | Add a new event type. Expects a JSON body with an `event-type` field. |
| GET | /events | List all registered event types. |
| DELETE | /events/{event} | Remove an event type and notify subscribers. |
| POST | /events/{event} | Publish an event for the given event type. Expects a JSON body containing event data. |
| GET | /event-stream | Subscribe to one or more event types and receive events as SSE. Expects one or more `event` query parameters. |

## Environment Variables
| Name | Description |
| --- | --- |
| APP_VERSION | Version of running server, reported by health check. Defaults to 0.0 if unset. |
| SERVER_PORT | Port for server to listen on. Defaults to 8080 if unset. |
| BASE_PATH | Base path to prepend to above API routes. Defaults to /event-bus if unset. |
| EVENT_BUS | Flag to configure underlying event bus implementation. Currently, only "channel" is supported and is the default. |
| TLS_CERT_PATH | Path to .crt cert file used for TLS. If unset, server will not use TLS or HTTP/2. |
| TLS_KEY_PATH | Patk to .key keyfile used for TLS. If unset, server will not use TLS or HTTP/2. |


### Example Requests

#### Health check

```bash
curl http://localhost:8080/health
```

#### Add an event type

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"event-type":"user.created"}'
```

#### List event types

```bash
curl http://localhost:8080/events
```

#### Delete an event type

```bash
curl -X DELETE http://localhost:8080/events/user.created
```

#### Publish an event

```bash
curl -X POST http://localhost:8080/events/user.created \
  -H "Content-Type: application/json" \
  -d '{"data":{"id":1}}'
```

#### Stream events

```bash
curl -N "http://localhost:8080/event-stream?event=user.created&event=user.deleted"
```

## Running the Server

Use the following command to start the server:

```bash
go run ./event-bus.go
```

## Testing

Run the test suite with:

```bash
go test ./...
```

## Building

Use the following command to build an executable:

```bash
go build -o event-bus.exe ./event-bus.go
```