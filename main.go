package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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

type Event struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	AgentName string    `json:"agent_name"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	logger = log.New(os.Stdout, "[AgentDock] ", log.LstdFlags)
)

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

func detectState(cpu float64, agentID string) string {
	if cpu > 5.0 {
		return "working"
	}
	if cpu > 0.5 {
		return "active"
	}
	return "waiting"
}

func handleAgents(w http.ResponseWriter, r *http.Request) {
	agents := store.getAgents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	l := r.URL.Query().Get("limit")
	if l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit > 200 {
			limit = 200
		}
	}
	events := store.getEvents(limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	agents := store.getAgents()
	events := store.getEvents(200)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"time":         time.Now().Format(time.RFC3339),
		"ntfy_enabled": ntfy.Enabled,
		"agents_count": len(agents),
		"events_count": len(events),
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
	initStore()

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

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
