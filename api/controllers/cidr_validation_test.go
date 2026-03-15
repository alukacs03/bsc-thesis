package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireCIDR(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		field     string
		wantValue string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid CIDR passes and returns trimmed value",
			value:     "10.0.0.0/24",
			field:     "test_field",
			wantValue: "10.0.0.0/24",
			wantErr:   false,
		},
		{
			name:    "invalid string fails with error",
			value:   "not-a-cidr",
			field:   "test_field",
			wantErr: true,
			errMsg:  "test_field must be a valid CIDR",
		},
		{
			name:      "whitespace is trimmed and passes",
			value:     "  10.0.0.0/24  ",
			field:     "test_field",
			wantValue: "10.0.0.0/24",
			wantErr:   false,
		},
		{
			name:    "empty string fails with error",
			value:   "",
			field:   "test_field",
			wantErr: true,
			errMsg:  "test_field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := requireCIDR(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestCheckCIDROverlaps(t *testing.T) {
	tests := []struct {
		name    string
		pools   map[string]string
		wantErr bool
		errMust []string // substrings the error must contain
	}{
		{
			name: "overlapping CIDRs detected",
			pools: map[string]string{
				"big_pool":   "10.0.0.0/8",
				"small_pool": "10.1.0.0/16",
			},
			wantErr: true,
			errMust: []string{"big_pool", "small_pool"},
		},
		{
			name: "non-overlapping CIDRs pass",
			pools: map[string]string{
				"pool_a": "10.0.0.0/24",
				"pool_b": "192.168.0.0/24",
			},
			wantErr: false,
		},
		{
			name: "single pool passes",
			pools: map[string]string{
				"only_pool": "10.0.0.0/24",
			},
			wantErr: false,
		},
		{
			name: "identical CIDRs detected as overlap",
			pools: map[string]string{
				"pool_x": "172.16.0.0/16",
				"pool_y": "172.16.0.0/16",
			},
			wantErr: true,
			errMust: []string{"pool_x", "pool_y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkCIDROverlaps(tt.pools)
			if tt.wantErr {
				assert.Error(t, err)
				for _, sub := range tt.errMust {
					assert.Contains(t, err.Error(), sub)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
