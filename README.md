# AgentDock 🐑

Lightweight, self-hosted web dashboard for monitoring AI agents.

![Docker](https://img.shields.io/badge/Docker-29.7.2-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![License](https://img.shields.io/badge/License-MIT-green)

## What is AgentDock?

AgentDock is a Dockerized service that auto-detects running AI agents on your machine and displays them in a clean, real-time web dashboard. Think of it as a control plane for your AI coding agents.

**Key features:**
- 🔍 Auto-detects agents (Hermes, OpenCode, Kilo, Claude Code, Goose, etc.)
- 📊 Real-time CPU/RAM monitoring per agent
- 📅 Event timeline
- 🔔 ntfy notifications when agents block/finish/change state
- 🐳 Runs in Docker with <50MB RAM
- 💻 Works on low-resource machines (2 cores, 8GB RAM)

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/marvel-254/agentdock/main/install.sh | sh
```

Then open http://localhost:8080

## Manual Install

```bash
# Clone
git clone https://github.com/marvel-254/agentdock.git
cd agentdock

# Build and run with Docker Compose
docker compose up -d

# Or build and run manually
docker build -t agentdock .
docker run -d --name agentdock -p 8080:8080 -v agentdock-data:/data agentdock
```

## Configuration

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Web server port |
| `TZ` | `UTC` | Timezone |

### ntfy Notifications

| Variable | Default | Description |
|----------|---------|-------------|
| `NTFY_ENABLED` | `0` | Enable notifications (1/true to enable) |
| `NTFY_URL` | `https://ntfy.sh` | ntfy base URL |
| `NTFY_TOPIC` | `agentdock-alerts` | ntfy topic |
| `NTFY_TOKEN` | (empty) | Bearer token for private topics |
| `NTFY_PRIORITY_DONE` | `default` | Priority for completed agents |
| `NTFY_PRIORITY_BLOCKED` | `high` | Priority for blocked agents |
| `NTFY_PRIORITY_WORKING` | `low` | Priority for working agents |

### Enable ntfy

**Option 1: Docker Compose**

Add to your `docker-compose.yml`:
```yaml
environment:
  - NTFY_ENABLED=1
  - NTFY_URL=https://ntfy.sh
  - NTFY_TOPIC=my-agent-alerts
```

Then restart:
```bash
docker compose down
docker compose up -d
```

**Option 2: Docker run**

```bash
docker stop agentdock
docker rm agentdock
docker run -d \
  --name agentdock \
  --restart unless-stopped \
  -p 8080:8080 \
  -v agentdock-data:/data \
  -e NTFY_ENABLED=1 \
  -e NTFY_URL=https://ntfy.sh \
  -e NTFY_TOPIC=my-agent-alerts \
  ghcr.io/marvel-254/agentdock:latest
```

**Option 3: Self-hosted ntfy**

Set `NTFY_URL` to your own ntfy server:
```bash
-e NTFY_URL=http://localhost:10081
```

### Subscribing to Notifications

Install the [ntfy app](https://ntfy.sh/app) on your phone and subscribe to your topic. You'll get push notifications when:
- An agent starts
- An agent changes state (working/blocked/idle)
- An agent stops

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/agents` | List all detected agents |
| `GET /api/events` | Recent events (add `?limit=50`) |
| `GET /api/health` | Health check + ntfy status |
| `POST /api/ntfy/publish` | Send manual ntfy notification |

### Example

```bash
# Get agents
curl http://localhost:8080/api/agents | jq

# Get events
curl http://localhost:8080/api/events?limit=20 | jq

# Send manual notification
curl -X POST http://localhost:8080/api/ntfy/publish \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","message":"Hello from AgentDock","topic":"my-agent-alerts"}'
```

## Detected Agents

| Agent | Process Name |
|-------|--------------|
| Hermes | `hermes` |
| OpenCode | `opencode` |
| Kilo | `kilo` |
| Claude Code | `claude` |
| Goose | `goose` |
| Qwen Code | `qwen` |
| Codex | `codex` |
| Cursor Agent | `cursor` |
| Cline | `cline` |
| Kiro | `kiro` |
| Droid | `droid` |
| Amp | `amp` |
| Grok | `grok` |

## Architecture

```
┌─────────────────────────────────────┐
│  Docker Container (AgentDock)       │
│                                     │
│  ┌─────────┐  ┌──────────────────┐  │
│  │ Go API  │  │ Process Scanner  │  │
│  │ server  │──│ (reads /proc)    │  │
│  └────┬────┘  └──────────────────┘  │
│       │                             │
│  ┌────▼────┐  ┌──────────────────┐  │
│  │ SQLite  │  │ ntfy notifier   │  │
│  │ store   │  │                  │  │
│  └─────────┘  └──────────────────┘  │
│                                     │
└─────────────────────────────────────┘
         │
    Browser UI
    (htmx dashboard)
```

## Resource Usage

- Container RAM: ~30-50 MB
- Container CPU: <1% idle, <5% scanning
- Disk: ~20MB image + SQLite growth

## Development

```bash
# Run locally (requires Go 1.22+)
go mod download
go run .

# Build Docker image
docker build -t agentdock .

# Run with ntfy enabled locally
NTFY_ENABLED=1 NTFY_URL=https://ntfy.sh NTFY_TOPIC=test go run .
```

## Troubleshooting

**"No agents detected"**
- Ensure AI agents are running
- Check `docker logs agentdock` for scan errors

**ntfy notifications not received**
- Check `/api/health` to see ntfy status
- Verify ntfy topic subscription on your phone
- Check `docker logs agentdock` for ntfy errors

**High CPU on host**
- Increase scanner interval (rebuild with custom code)
- The scanner runs every 5s; adjust in `scanner.go`

## License

MIT
