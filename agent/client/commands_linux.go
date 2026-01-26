//go:build linux
// +build linux

package client

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CommandResult struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// DecommissionHandler is set by main to handle decommission commands
var DecommissionHandler func()

func executeCommands(commands []struct {
	ID      uint            `json:"id"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}) []CommandResult {
	out := make([]CommandResult, 0, len(commands))
	for _, cmd := range commands {
		if cmd.ID == 0 {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(cmd.Kind)) {
		case "restart_service":
			out = append(out, runRestartService(cmd.ID, cmd.Payload))
		case "decommission":
			out = append(out, runDecommission(cmd.ID))
		default:
			out = append(out, CommandResult{ID: cmd.ID, Status: "failed", Error: "unsupported command"})
		}
	}
	return out
}

func runDecommission(id uint) CommandResult {
	log.Println("Received decommission command, initiating cleanup...")

	// If a handler is set, call it (allows main to do cleanup)
	if DecommissionHandler != nil {
		DecommissionHandler()
	}

	// Return success before exiting
	result := CommandResult{ID: id, Status: "succeeded", Output: "Decommission initiated"}

	// Schedule exit after a short delay to allow result to be sent
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Agent decommissioned, exiting...")
		os.Exit(0)
	}()

	return result
}

func runRestartService(id uint, payload json.RawMessage) CommandResult {
	var p struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(payload, &p)
	service := strings.TrimSpace(p.Name)
	if service == "" {
		return CommandResult{ID: id, Status: "failed", Error: "missing service name"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "restart", service)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return CommandResult{ID: id, Status: "failed", Output: string(b), Error: err.Error()}
	}
	return CommandResult{ID: id, Status: "succeeded", Output: string(b)}
}
