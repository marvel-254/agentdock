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
- 🔔 ntfy notifications when agents block/finish
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

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/agents` | List all detected agents |
| `GET /api/events` | Recent events (add `?limit=50`) |
| `GET /api/health` | Health check |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Web server port |
| `TZ` | `UTC` | Timezone |

## Detected Agents

- Hermes
- OpenCode
- Kilo
- Claude Code
- Goose
- Qwen Code
- Codex
- Cursor Agent
- Cline
- Kiro
- Droid
- Amp
- Grok

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
```

## License

MIT
