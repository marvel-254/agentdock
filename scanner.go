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

func scanLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := scanProcesses(); err != nil {
			logger.Printf("Scan error: %v", err)
		}
	}
}

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

		commPath := filepath.Join("/proc", entry.Name(), "comm")
		comm, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		procName := strings.TrimSpace(string(comm))
		displayName, isAgent := agentPatterns[procName]

		if !isAgent {
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

		cpuUsage := getCPUUsage(entry.Name())
		agentID := fmt.Sprintf("%s-%d", procName, pid)
		newState := detectState(cpuUsage, agentID)

		agent := Agent{
			ID:        agentID,
			Name:      displayName,
			PID:       pid,
			CPU:       cpuUsage,
			RAM:       ramKB * 1024,
			Status:    newState,
			StartedAt: now,
			LastSeen:  now,
		}

		changed, oldState := store.upsertAgent(agent)
		if changed {
			if oldState == "" {
				store.addEvent("agent_started", fmt.Sprintf("%s started (PID %d)", displayName, pid), displayName)
				sendNtfy("AgentDock", fmt.Sprintf("🟢 %s started", displayName), "low", "rocket,agent")
			} else {
				trackAgentStateChange(agentID, displayName, oldState, newState)
			}
		}
	}

	// Remove stale
	threshold := now.Add(-30 * time.Second)
	removed := store.removeStale(threshold)
	for _, name := range removed {
		store.addEvent("agent_stopped", fmt.Sprintf("%s stopped", name), name)
		sendNtfy("AgentDock", fmt.Sprintf("🔴 %s stopped", name), "high", "stop_sign,agent")
	}

	return nil
}

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

	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	totalTime := utime + stime

	return float64(totalTime%100) / 100.0 * 100
}
