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
	seen := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a PID
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Read process name from /proc/[pid]/comm
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

		// Read process stats
		statPath := filepath.Join("/proc", entry.Name(), "stat")
		statData, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		// Parse memory from /proc/[pid]/status
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

		// Generate agent ID
		agentID := fmt.Sprintf("%s-%d", procName, pid)

		// Determine status based on recent activity
		status := "working"

		// Upsert agent in database
		dbMu.Lock()
		_, err = db.Exec(`
			INSERT INTO agents (id, name, pid, cpu, ram, status, started_at, last_seen)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				cpu = excluded.cpu,
				ram = excluded.ram,
				status = excluded.status,
				last_seen = excluded.last_seen
		`, agentID, displayName, pid, 0.0, ramKB*1024, status, now, now)
		dbMu.Unlock()

		if err != nil {
			logger.Printf("Failed to upsert agent %s: %v", agentID, err)
		}

		seen[agentID] = true
	}

	// Remove stale agents (not seen in last 30 seconds)
	dbMu.Lock()
	threshold := now.Add(-30 * time.Second)
	_, err = db.Exec("DELETE FROM agents WHERE last_seen < ?", threshold)
	if err != nil {
		logger.Printf("Failed to clean stale agents: %v", err)
	}
	dbMu.Unlock()

	return nil
}
