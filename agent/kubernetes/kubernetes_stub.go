//go:build !linux
// +build !linux

package kubernetes

import (
	"context"
	"gluon-agent/client"
)


func Sync(_ context.Context, _ *client.Client, _ string) {}
