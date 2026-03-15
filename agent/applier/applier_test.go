package applier

import (
	"gluon-agent/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeedsUpdate(t *testing.T) {
	tests := []struct {
		name   string
		bundle *client.ConfigBundle
		state  *ConfigState
		want   bool
	}{
		{
			name:   "higher version returns true",
			bundle: &client.ConfigBundle{Version: 2, Hash: "abc"},
			state:  &ConfigState{Version: 1, Hash: "abc"},
			want:   true,
		},
		{
			name:   "same version but different hash returns true",
			bundle: &client.ConfigBundle{Version: 1, Hash: "new-hash"},
			state:  &ConfigState{Version: 1, Hash: "old-hash"},
			want:   true,
		},
		{
			name:   "same version and same hash returns false",
			bundle: &client.ConfigBundle{Version: 1, Hash: "abc"},
			state:  &ConfigState{Version: 1, Hash: "abc"},
			want:   false,
		},
		{
			name:   "zero state with bundle version > 0 returns true",
			bundle: &client.ConfigBundle{Version: 1, Hash: "abc"},
			state:  &ConfigState{Version: 0, Hash: ""},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsUpdate(tt.bundle, tt.state)
			assert.Equal(t, tt.want, got)
		})
	}
}
