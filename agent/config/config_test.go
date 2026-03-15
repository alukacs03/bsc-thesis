package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEnrolled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "both NodeID and APIKey set returns true",
			config: &Config{NodeID: "42", APIKey: "key-123"},
			want:   true,
		},
		{
			name:   "only NodeID set returns false",
			config: &Config{NodeID: "42", APIKey: ""},
			want:   false,
		},
		{
			name:   "only APIKey set returns false",
			config: &Config{NodeID: "", APIKey: "key-123"},
			want:   false,
		},
		{
			name:   "both empty returns false",
			config: &Config{NodeID: "", APIKey: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnrolled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasPendingEnrollment(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "RequestID and EnrollmentSecret set without APIKey returns true",
			config: &Config{
				RequestID:        "req-1",
				EnrollmentSecret: "secret-abc",
				APIKey:           "",
			},
			want: true,
		},
		{
			name: "with APIKey set returns false",
			config: &Config{
				RequestID:        "req-1",
				EnrollmentSecret: "secret-abc",
				APIKey:           "key-123",
			},
			want: false,
		},
		{
			name: "missing RequestID returns false",
			config: &Config{
				RequestID:        "",
				EnrollmentSecret: "secret-abc",
				APIKey:           "",
			},
			want: false,
		},
		{
			name: "missing EnrollmentSecret returns false",
			config: &Config{
				RequestID:        "req-1",
				EnrollmentSecret: "",
				APIKey:           "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.HasPendingEnrollment()
			assert.Equal(t, tt.want, got)
		})
	}
}
