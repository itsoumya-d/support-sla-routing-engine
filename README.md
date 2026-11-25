# Support SLA Routing Engine

Low-latency support ticket routing, skill-matched dispatching, and automated SLA breach escalation engine written in Go.

## Features
- Dynamic SLA deadline calculation based on ticket urgency and customer tier.
- Concurrent skill-matched agent pool dispatcher with load balancing.
- Thread-safe breach monitor and escalation triggers.
- Sub-millisecond routing latency with race-tested concurrency.

## Usage
```bash
go test -v -race ./...
```
