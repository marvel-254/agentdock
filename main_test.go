package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if resp["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", resp["status"])
	}
}

func TestAgentsEmpty(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/agents", nil)
	w := httptest.NewRecorder()
	handleAgents(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var agents []Agent
	if err := json.Unmarshal(w.Body.Bytes(), &agents); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(agents) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(agents))
	}
}

func TestEventsEmpty(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/events?limit=5", nil)
	w := httptest.NewRecorder()
	handleEvents(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var events []Event
	if err := json.Unmarshal(w.Body.Bytes(), &events); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestRecordEvent(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	if err := recordEvent("test", "test message", "TestAgent"); err != nil {
		t.Fatalf("recordEvent failed: %v", err)
	}

	// Verify event was recorded
	req := httptest.NewRequest("GET", "/api/events?limit=10", nil)
	w := httptest.NewRecorder()
	handleEvents(w, req)

	var events []Event
	json.Unmarshal(w.Body.Bytes(), &events)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != "test" {
		t.Fatalf("expected type=test, got %s", events[0].Type)
	}

	if events[0].Message != "test message" {
		t.Fatalf("expected message='test message', got %s", events[0].Message)
	}
}

func TestPublishNtfyMethod(t *testing.T) {
	os.Setenv("PORT", "18080")
	os.Setenv("NTFY_ENABLED", "0")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	// POST with valid body
	body := `{"topic":"test","title":"Hello","message":"World"}`
	req := httptest.NewRequest("POST", "/api/ntfy/publish", strings.NewReader(body))
	w := httptest.NewRecorder()
	handlePublishNtfy(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPublishNtfyGET(t *testing.T) {
	os.Setenv("PORT", "18080")
	os.Setenv("NTFY_ENABLED", "0")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	// GET should fail
	req := httptest.NewRequest("GET", "/api/ntfy/publish", nil)
	w := httptest.NewRecorder()
	handlePublishNtfy(w, req)

	if w.Code != 405 {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestIndex(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handleIndex(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "AgentDock") {
		t.Fatalf("response doesn't contain AgentDock")
	}
}

func TestDetectState(t *testing.T) {
	os.Setenv("PORT", "18080")
	initNtfy()
	if err := initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	tests := []struct {
		cpu      float64
		expected string
	}{
		{10.0, "working"},
		{5.0, "working"},
		{1.0, "active"},
		{0.3, "waiting"},
	}

	for _, tt := range tests {
		state := detectState(tt.cpu, "nonexistent")
		if state != tt.expected {
			t.Fatalf("detectState(%f) = %s, want %s", tt.cpu, state, tt.expected)
		}
	}
}

func TestSendNtfyDisabled(t *testing.T) {
	os.Setenv("NTFY_ENABLED", "0")
	initNtfy()

	// Should not panic or error when disabled
	sendNtfy("Test", "Message", "high", "rotating_light")
}

func TestSendNtfyEnabledNoURL(t *testing.T) {
	os.Setenv("NTFY_ENABLED", "1")
	os.Setenv("NTFY_URL", "")
	initNtfy()

	// Should handle gracefully when URL is empty
	sendNtfy("Test", "Message", "high", "rotating_light")
}
