package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "[AgentDock] ", log.LstdFlags)

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

func main() {
	initNtfy()
	initStore()

	go scanLoop()
	go heartbeat()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/agents", handleAgents)
	http.HandleFunc("/api/events", handleEvents)
	http.HandleFunc("/api/health", handleHealth)

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
