package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// scanLoop continuously scans /proc for AI agent processes
func scanLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := scanProcesses(); err != nil {
			logger.Printf("Scan error: %v", err)
		}
	}
}

// scanProcesses reads /proc and detects AI agent processes
func scanProcesses() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("failed to read /proc: %w", err)
	}

	now := time.Now()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Read process name
		commPath := filepath.Join("/proc", entry.Name(), "comm")
		comm, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		procName := strings.TrimSpace(string(comm))
		displayName, isAgent := agentPatterns[procName]
		
		if !isAgent {
			// Check cmdline for agent patterns
			cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
			cmdline, err := os.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}
			cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
			for pattern, name := range agentPatterns {
				if strings.Contains(strings.ToLower(cmdStr), pattern) {
					displayName = name
					isAgent = true
					break
				}
			}
		}
		if !isAgent {
			continue
		}

		// Read memory from /proc/[pid]/status
		var ramKB int64
		statusPath := filepath.Join("/proc", entry.Name(), "status")
		statusFile, err := os.Open(statusPath)
		if err == nil {
			defer statusFile.Close()
			scanner := bufio.NewScanner(statusFile)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "VmRSS:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						fmt.Sscanf(fields[1], "%d", &ramKB)
					}
					break
				}
			}
		}

		// Calculate CPU usage (simplified)
		cpuUsage := getCPUUsage(entry.Name())

		agentID := fmt.Sprintf("%s-%d", procName, pid)
		
		// Detect state
		newState := detectState(cpuUsage, agentID)

		// Get old state and update
		dbMu.Lock()
		var oldState string
		err = db.QueryRow("SELECT status FROM agents WHERE id = ?", agentID).Scan(&oldState)
		
		if err == sql.ErrNoRows {
			// New agent
			recordEvent("agent_started", fmt.Sprintf("%s started (PID %d)", displayName, pid), displayName)
			sendNtfy("AgentDock", fmt.Sprintf("🟢 %s started", displayName), "low", "rocket,agent")
		} else if oldState != newState {
			trackAgentStateChange(agentID, displayName, oldState, newState)
		}

		// Upsert agent
		_, err = db.Exec(`
			INSERT INTO agents (id, name, pid, cpu, ram, status, started_at, last_seen)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				cpu = excluded.cpu,
				ram = excluded.ram,
				status = excluded.status,
				last_seen = excluded.last_seen
		`, agentID, displayName, pid, cpuUsage, ramKB*1024, newState, now, now)
		dbMu.Unlock()

		if err != nil {
			logger.Printf("Failed to upsert agent %s: %v", agentID, err)
		}
	}

	// Remove stale agents
	dbMu.Lock()
	threshold := now.Add(-30 * time.Second)
	rows, _ := db.Query("SELECT id, name FROM agents WHERE last_seen < ?", threshold)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, name string
			rows.Scan(&id, &name)
			recordEvent("agent_stopped", fmt.Sprintf("%s stopped", name), name)
			sendNtfy("AgentDock", fmt.Sprintf("🔴 %s stopped", name), "high", "stop_sign,agent")
		}
	}
	_, _ = db.Exec("DELETE FROM agents WHERE last_seen < ?", threshold)
	dbMu.Unlock()

	return nil
}

// getCPUUsage calculates a simplified CPU usage percentage from /proc/[pid]/stat
func getCPUUsage(pid string) float64 {
	statPath := filepath.Join("/proc", pid, "stat")
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) < 17 {
		return 0
	}

	// utime + stime (fields 14 and 15)
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	totalTime := utime + stime

	// Convert to percentage (rough approximation)
	// This is simplified; for accurate CPU you'd need to track over time
	return float64(totalTime%100) / 100.0 * 100
}
