package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Agent represents a detected AI agent process
type Agent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PID       int       `json:"pid"`
	CPU       float64   `json:"cpu"`
	RAM       int64     `json:"ram"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// Event represents an agent lifecycle event
type Event struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	AgentName string    `json:"agent_name"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	db     *sql.DB
	dbMu   sync.RWMutex
	logger = log.New(os.Stdout, "[AgentDock] ", log.LstdFlags)
)

// Known agent process names and their display names
var agentPatterns = map[string]string{
	"hermes":       "Hermes",
	"opencode":     "OpenCode",
	"kilo":         "Kilo",
	"claude":       "Claude Code",
	"goose":        "Goose",
	"qwen":         "Qwen Code",
	"codex":        "Codex",
	"cursor":       "Cursor Agent",
	"cline":        "Cline",
	"kiro":         "Kiro",
	"droid":        "Droid",
	"amp":          "Amp",
	"grok":         "Grok",
	"hermes-agent": "Hermes Agent",
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "/data/agentdock.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		pid INTEGER,
		cpu REAL DEFAULT 0,
		ram INTEGER DEFAULT 0,
		status TEXT DEFAULT 'unknown',
		started_at DATETIME,
		last_seen DATETIME
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		message TEXT,
		agent_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen);
	CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func handleAgents(w http.ResponseWriter, r *http.Request) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	rows, err := db.Query(`
		SELECT id, name, pid, cpu, ram, status, started_at, last_seen 
		FROM agents 
		ORDER BY last_seen DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var a Agent
		err := rows.Scan(&a.ID, &a.Name, &a.PID, &a.CPU, &a.RAM, &a.Status, &a.StartedAt, &a.LastSeen)
		if err != nil {
			continue
		}
		agents = append(agents, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	limit := 50
	l := r.URL.Query().Get("limit")
	if l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit > 200 {
			limit = 200
		}
	}

	rows, err := db.Query(`
		SELECT id, type, message, agent_name, created_at 
		FROM events 
		ORDER BY created_at DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.Type, &e.Message, &e.AgentName, &e.CreatedAt)
		if err != nil {
			continue
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	var agentCount, eventCount int
	db.QueryRow("SELECT COUNT(*) FROM agents WHERE last_seen > datetime('now', '-5 minutes')").Scan(&agentCount)
	db.QueryRow("SELECT COUNT(*) FROM events WHERE created_at > datetime('now', '-24 hours')").Scan(&eventCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"time":          time.Now().Format(time.RFC3339),
		"ntfy_enabled":  ntfy.Enabled,
		"agents_24h":    agentCount,
		"events_24h":    eventCount,
		"ntfy_url":      ntfy.BaseURL,
		"ntfy_topic":    ntfy.Topic,
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}

func handlePublishNtfy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}

	var req struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
		Title   string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.Topic == "" {
		req.Topic = ntfy.Topic
	}
	if req.Title == "" {
		req.Title = "AgentDock"
	}

	priority := getEnv("NTFY_PRIORITY_DONE", "default")
	tags := "robot,agentdock"

	sendNtfy(req.Title, req.Message, priority, tags)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func main() {
	initNtfy()

	if err := initDB(); err != nil {
		logger.Fatalf("Database init failed: %v", err)
	}
	defer db.Close()

	go scanLoop()
	go heartbeat()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/agents", handleAgents)
	http.HandleFunc("/api/events", handleEvents)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/ntfy/publish", handlePublishNtfy)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Printf("AgentDock starting on :%s", port)
	logger.Printf("ntfy: enabled=%v url=%s topic=%s", ntfy.Enabled, ntfy.BaseURL, ntfy.Topic)
	logger.Printf("ntfy config: URL=%s, Topic=%s, Token=%v", ntfy.BaseURL, ntfy.Topic, ntfy.Token != "")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
