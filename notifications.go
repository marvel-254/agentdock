package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ntfyConfig struct {
	Enabled         bool
	BaseURL         string
	Topic           string
	Token           string
	PriorityDone    string
	PriorityBlocked string
	PriorityWorking string
}

var ntfy ntfyConfig

func initNtfy() {
	ntfy = ntfyConfig{
		Enabled:         os.Getenv("NTFY_ENABLED") == "1" || os.Getenv("NTFY_ENABLED") == "true",
		BaseURL:         getEnv("NTFY_URL", "https://ntfy.sh"),
		Topic:           getEnv("NTFY_TOPIC", "orionos-alerts"),
		Token:           os.Getenv("NTFY_TOKEN"),
		PriorityDone:    getEnv("NTFY_PRIORITY_DONE", "default"),
		PriorityBlocked: getEnv("NTFY_PRIORITY_BLOCKED", "high"),
		PriorityWorking: getEnv("NTFY_PRIORITY_WORKING", "low"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sendNtfy(title, message, priority, tags string) {
	if !ntfy.Enabled || ntfy.BaseURL == "" || ntfy.Topic == "" {
		return
	}

	url := fmt.Sprintf("%s/%s", ntfy.BaseURL, ntfy.Topic)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		logger.Printf("ntfy request error: %v", err)
		return
	}

	req.Header.Set("Title", title)
	if priority != "" {
		req.Header.Set("Priority", priority)
	}
	if tags != "" {
		req.Header.Set("Tags", tags)
	}
	if ntfy.Token != "" {
		req.Header.Set("Authorization", "Bearer "+ntfy.Token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("ntfy send failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Printf("ntfy sent: %s", title)
	} else {
		logger.Printf("ntfy failed with status: %d", resp.StatusCode)
	}
}

func trackAgentStateChange(agentID, name, oldStatus, newStatus string) {
	if oldStatus == newStatus {
		return
	}

	msg := fmt.Sprintf("%s is now %s", name, newStatus)
	var priority, tags string

	switch newStatus {
	case "blocked":
		priority = ntfy.PriorityBlocked
		tags = "rotating_light,agent"
		msg = fmt.Sprintf("🔴 %s needs attention!", name)
	case "done", "completed":
		priority = ntfy.PriorityDone
		tags = "white_check_mark,agent"
		msg = fmt.Sprintf("✅ %s finished", name)
	case "failed":
		priority = ntfy.PriorityBlocked
		tags = "x,agent"
		msg = fmt.Sprintf("❌ %s failed", name)
	case "working":
		priority = ntfy.PriorityWorking
		tags = "gear,agent"
	}

	sendNtfy("OrionOS", msg, priority, tags)
	store.addEvent("state_change", fmt.Sprintf("%s: %s → %s", name, oldStatus, newStatus), name)
}

func heartbeat() {
	if !ntfy.Enabled {
		return
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		agents := store.getAgents()
		if len(agents) > 0 {
			sendNtfy(
				"OrionOS Summary",
				fmt.Sprintf("%d agents currently working", len(agents)),
				"low",
				"chart,agent",
			)
		}
	}
}
