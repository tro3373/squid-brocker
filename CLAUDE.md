# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

squid-brocker is a Go-based Squid proxy helper that enforces per-device, per-domain-group daily cumulative access time limits. Designed for Raspberry Pi deployment to manage children's internet usage. Squid's `external_acl_type` with `ttl=60` queries the Go helper every 60 seconds; each query counts as 1 minute of usage.

## Build & Development Commands

```bash
make build        # go build -o bin/squid-helper ./cmd/squid-helper
make test         # go test ./... -v -count=1
make lint         # golangci-lint run ./...
make dev          # pipe sample input to helper via go run
make up           # docker compose up -d + show logs
make down         # docker compose down + cleanup
make restart      # restart squid container
make build-images # docker compose build
```

Single test: `go test ./internal/tracker/ -run TestCheckAccess_ExceedsLimit -v`

## Architecture

```
cmd/squid-helper/main.go  → Entry point: flags(--config, --data), signal handling
internal/
  config/   → YAML rules loading & validation (DomainGroup, Rule, Limit)
  handler/  → Squid stdin/stdout protocol (reads "IP DOMAIN", writes "OK"/"ERR")
  tracker/  → Core logic: cumulative time tracking per (device, group, date)
              store.go: Store interface with MemoryStore (test) / FileStore (prod, JSON)
squid/      → Squid config, Docker entrypoint, error page HTML
```

**Data flow**: Squid → stdin → `handler.Run()` → `tracker.CheckAccess()` → config lookup → state update → stdout → Squid

**Key design decisions**:
- **Fail-open**: Unknown devices/domains are allowed (not blocked)
- **State persistence**: JSON file at `--data` path, protected by `sync.Mutex`
- **Date rollover**: Usage resets daily via date-keyed state entries
- **Minimal dependencies**: Only `gopkg.in/yaml.v3`

## Configuration

- `rules.yaml` (gitignored) - production config mounted into container
- `testdata/rules_test.yaml` - test fixture with sample devices/groups
- `.golangci.yaml` - linters: gosec, govet, staticcheck (auto-fix enabled)
- Go version: see `.tool-versions` (asdf)

## Docker

Multi-stage build (golang alpine → alpine + squid). Container exposes port 3128. Volumes: `/etc/squid-brocker/rules.yaml` (config, ro), `/var/lib/squid-brocker` (state data).
