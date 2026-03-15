package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualPublicKeys(t *testing.T) {
	tests := []struct {
		name string
		a    map[string]string
		b    map[string]string
		want bool
	}{
		{
			name: "identical maps",
			a:    map[string]string{"node1": "pubkey1", "node2": "pubkey2"},
			b:    map[string]string{"node1": "pubkey1", "node2": "pubkey2"},
			want: true,
		},
		{
			name: "different values for same key",
			a:    map[string]string{"node1": "pubkey1"},
			b:    map[string]string{"node1": "pubkey_other"},
			want: false,
		},
		{
			name: "different keys",
			a:    map[string]string{"node1": "pubkey1"},
			b:    map[string]string{"node2": "pubkey1"},
			want: false,
		},
		{
			name: "both empty maps",
			a:    map[string]string{},
			b:    map[string]string{},
			want: true,
		},
		{
			name: "one nil one empty",
			a:    nil,
			b:    map[string]string{},
			want: true,
		},
		{
			name: "extra key in first map",
			a:    map[string]string{"node1": "pubkey1", "node2": "pubkey2"},
			b:    map[string]string{"node1": "pubkey1"},
			want: false,
		},
		{
			name: "extra key in second map",
			a:    map[string]string{"node1": "pubkey1"},
			b:    map[string]string{"node1": "pubkey1", "node2": "pubkey2"},
			want: false,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EqualPublicKeys(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
