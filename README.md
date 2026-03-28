# go-propagator

Webhook proxy for Kubernetes environments. Receives incoming webhooks
and propagates them concurrently to multiple configured targets.

Built with Go's standard library — no frameworks, no external HTTP routers,
no DI containers.

## Motivation

I needed a way to route Grafana alerts to multiple endpoints in my FluxCD k3s
homelab. Rather than reaching for an existing tool, I used it as an opportunity
to learn Go by building something I'd actually deploy and operate.

## What it does

```
POST /webhook  →  target A
                  target B
                  target C
```

Accepts any payload, preserves the original `Content-Type`, and fans out
to all configured endpoints with bounded concurrency. Returns:

- `200` — all targets succeeded
- `207` — some targets failed, some succeeded
- `500` — all targets failed

## Learnings

**Concurrency** — Used a channel as a semaphore to bound fan-out, understanding the importance of this mechanism in Go's concurrency model.

**Error handling** — Practiced some error handling and error propagation.

**Implicit interfaces** — Types satisfy interfaces by having the right methods, which is something I enjoyed.

**Testing** — Learned about the common approaches to testing on Go, such as table tests.

## AI usage

It made easier to translate my theoretical knowledge and NodeJS experience to Go idioms, by sparring my ideas
with it to understand how problems are approached in this language and how professional code is developed in OSS.

## Local setup

```bash
# Prerequisites: Go 1.22+
git clone https://github.com/pausic/go-propagator.git
cd go-propagator

# Configure targets
cp config.example.yaml config.yaml
# Edit config.yaml with your webhook URLs

# Run
make run

# Test
make test

# Build
make build
```

## Configuration

```yaml
addr: "localhost:8080"
timeout: 10
concurrent: 5
webhooks:
  - "https://example.com/hook1"
  - "https://example.com/hook2"
  - "https://example.com/hook3"
```

| Field | Description |
|---|---|
| `addr` | Listen address |
| `timeout` | Shutdown timeout in seconds |
| `concurrent` | Max parallel outbound requests |
| `webhooks` | Target URLs to propagate to |

Override config path via `CONFIG_PATH` env var.

