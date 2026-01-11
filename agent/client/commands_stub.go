//go:build !linux
// +build !linux

package client


type CommandResult struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}
