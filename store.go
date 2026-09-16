package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Simple file-based storage to avoid CGO dependency
type Store struct {
	mu       sync.RWMutex
	agents   []Agent
	events   []Event
	filename string
}

var store *Store

func initStore() {
	store = &Store{
		agents:   make([]Agent, 0),
		events:   make([]Event, 0),
		filename: "/data/agentdock.json",
	}
	store.load()
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filename)
	if err != nil {
		return // File doesn't exist yet
	}

	var saved struct {
		Agents []Agent `json:"agents"`
		Events []Event `json:"events"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}
	s.agents = saved.Agents
	s.events = saved.Events
}

func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(map[string]interface{}{
		"agents": s.agents,
		"events": s.events,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}

func (s *Store) getAgents() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Agent, len(s.agents))
	copy(result, s.agents)
	return result
}

func (s *Store) getEvents(limit int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit > len(s.events) {
		limit = len(s.events)
	}
	result := make([]Event, limit)
	copy(result, s.events[:limit])
	return result
}

func (s *Store) upsertAgent(agent Agent) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find existing
	for i, a := range s.agents {
		if a.ID == agent.ID {
			oldStatus := a.Status
			s.agents[i] = agent
			return oldStatus != agent.Status, oldStatus
		}
	}

	// New agent
	s.agents = append(s.agents, agent)
	return true, ""
}

func (s *Store) removeStale(threshold time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := make([]string, 0)
	remaining := make([]Agent, 0)
	for _, a := range s.agents {
		if a.LastSeen.Before(threshold) {
			removed = append(removed, a.Name)
		} else {
			remaining = append(remaining, a)
		}
	}
	s.agents = remaining
	return removed
}

func (s *Store) addEvent(eventType, message, agentName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := Event{
		ID:        len(s.events) + 1,
		Type:      eventType,
		Message:   message,
		AgentName: agentName,
		CreatedAt: time.Now(),
	}
	s.events = append(s.events, event)
}
