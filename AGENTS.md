# OrionOS — Agent Instructions

## Project Overview
Self-hosted Go web dashboard for monitoring AI agents (OpenCode, Hermes, Claude Code, etc.). Single binary, no CGO, file-based JSON storage. Runs in Docker with <50MB RAM.

## Commands
| Task | Command |
|------|---------|
| Local dev | `go mod download && go run .` |
| Build binary | `go build -o orionos .` |
| Docker build | `docker build -t orionos .` |
| Docker run | `docker run -d --name orionos -p 8080:8080 -v orionos-data:/data -v /proc:/host/proc:ro -e HOST_PROC=/host/proc orionos` |
| Docker Compose | `docker compose up -d` |
| Run tests | `./test.sh` (runs `go test -v -count=1 ./...` in golang:1.22-alpine) |
| Lint/format | `gofmt -l .` (no linter configured) |

## Architecture
```
main.go       → HTTP server, routing, startup
scanner.go    → /proc scanner (5s interval), detects agents by process name/cmdline
store.go      → In-memory + JSON file persistence (/data/orionos.json), no external DB
notifications.go → ntfy.sh push notifications (optional)
dashboard.go  → Embedded HTML + JS for real-time UI
```

## Key Conventions
- **Single package** (`main`) — all `.go` files in root
- **No tests exist** — `test.sh` runs but finds no `_test.go` files
- **File storage** — JSON at `/data/orionos.json` (mounted volume in Docker)
- **Process detection** — Scans `/proc`, matches `comm` + `cmdline` against `agentPatterns` map
- **HOST_PROC** — Set to `/host/proc` in Docker to read host processes
- **No CGO** — Pure Go, uses `wget` for healthcheck in Docker

## Environment Variables
| Variable | Default | Purpose |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `TZ` | `UTC` | Timezone |
| `HOST_PROC` | `/proc` | Procfs path (use `/host/proc` in Docker) |
| `NTFY_ENABLED` | `0` | Enable ntfy (1/true) |
| `NTFY_URL` | `https://ntfy.sh` | ntfy server |
| `NTFY_TOPIC` | `orionos-alerts` | Notification topic |
| `NTFY_TOKEN` | — | Bearer token for private topics |
| `NTFY_PRIORITY_*` | varies | Priority per state |

## API Endpoints
- `GET /api/agents` — List detected agents
- `GET /api/events?limit=50` — Recent events
- `GET /api/health` — Health + ntfy status
- `POST /api/ntfy/publish` — Manual notification

## Development Notes
- Dashboard is embedded Go string (`dashboardHTML` const), not separate files
- Frontend polls `/api/agents` (5s) and `/api/events` (5s) via fetch
- Scanner interval hardcoded to 5s in `scanLoop()` — change in `scanner.go:14`
- Stale agent removal: 30s threshold in `scanner.go:120`
- CPU calculation in `getCPUUsage()` is approximate (utime+stime % 100)

## CI/CD
- GitHub Actions: `.github/workflows/docker.yml` builds on push to main
- Pushes to `ghcr.io/marvel-254/orionos:latest`
- Uses Docker Buildx with GHA cache

## Common Issues
| Symptom | Fix |
|---------|-----|
| No agents detected | Ensure agents running; check `docker logs orionos`; verify `/proc` mount |
| ntfy not working | Check `/api/health` for `ntfy_enabled`; verify topic subscription |
| High CPU | Scanner runs every 5s; increase interval in `scanner.go:14` and rebuild |
