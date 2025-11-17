# Synthesizer log tester

This small service emits synthetic JSON logs to help test log pipelines and collectors.

## Run the service

With Go installed:

```bash
go run .
```

The server listens on `:4000`.

## Trigger the log tester

The log worker is triggered via a POST to `/` with a small JSON body specifying how long it should emit logs and at what rate.

- `duration`: Go duration string (e.g. `"10s"`, `"2m"`)
- `rate`: number of log entries per second

Example: run the log tester for 30 seconds at 5 logs/second:

```bash
curl -X POST http://localhost:4000/ \
  -H 'Content-Type: application/json' \
  -d '{"duration":"30s","rate":5}'
```

You should then see log entries like `"some random fake data for log testing"` printed to stdout until the duration has elapsed.

## Test JSON responses

You can also hit the JSON endpoints directly:

- Fast response: `GET /`
- Slow response with artificial latency: `GET /slow`
- Always-500 endpoint: `GET /500`
- Request timeout endpoint: `GET /timeout`

Example:

```bash
curl http://localhost:4000/
```