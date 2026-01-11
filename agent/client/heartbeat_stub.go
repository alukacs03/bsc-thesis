//go:build !linux
// +build !linux

package client

import "errors"


func (c *Client) Heartbeat(_ string, _ string) error {
	return errors.New("heartbeat is only supported on linux")
}
