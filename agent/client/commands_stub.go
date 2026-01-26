//go:build !linux
// +build !linux

package client

// DecommissionHandler is set by main to handle decommission commands
var DecommissionHandler func()

type CommandResult struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}
