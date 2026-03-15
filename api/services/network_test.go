package services

import (
	"gluon-api/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNextAvailableIP(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		allocations []models.IPAllocation
		wantIP      *string
		wantErr     bool
	}{
		{
			name:        "empty allocations returns first usable IP",
			cidr:        "10.0.0.0/24",
			allocations: []models.IPAllocation{},
			wantIP:      strPtr("10.0.0.1"),
		},
		{
			name: "skips already allocated IPs",
			cidr: "10.0.0.0/24",
			allocations: []models.IPAllocation{
				{IP: "10.0.0.1"},
				{IP: "10.0.0.2"},
			},
			wantIP: strPtr("10.0.0.3"),
		},
		{
			name: "handles /32 suffix in allocation IP fields",
			cidr: "10.0.0.0/24",
			allocations: []models.IPAllocation{
				{IP: "10.0.0.1/32"},
				{IP: "10.0.0.2/32"},
				{IP: "10.0.0.3/32"},
			},
			wantIP: strPtr("10.0.0.4"),
		},
		{
			name: "mixed bare and CIDR-notation allocations",
			cidr: "10.0.0.0/24",
			allocations: []models.IPAllocation{
				{IP: "10.0.0.1"},
				{IP: "10.0.0.2/32"},
				{IP: "10.0.0.3"},
			},
			wantIP: strPtr("10.0.0.4"),
		},
		{
			name: "skips gaps and returns next free",
			cidr: "10.0.0.0/24",
			allocations: []models.IPAllocation{
				{IP: "10.0.0.1"},
				{IP: "10.0.0.3"},
			},
			wantIP: strPtr("10.0.0.2"),
		},
		{
			name: "pool exhausted returns nil",
			cidr: "10.0.0.0/30",
			allocations: []models.IPAllocation{
				{IP: "10.0.0.1"},
				{IP: "10.0.0.2"},
				{IP: "10.0.0.3"},
			},
			wantIP: nil,
		},
		{
			name:    "invalid CIDR returns error",
			cidr:    "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findNextAvailableIP(tt.cidr, tt.allocations)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantIP == nil {
				assert.Nil(t, got, "expected nil when pool is exhausted")
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.wantIP, *got)
			}
		})
	}
}

func TestHubWorkerListenPort(t *testing.T) {
	// Formula: 52000 + (hubNumber-1)*1000 + workerID
	tests := []struct {
		name      string
		hubNumber int
		workerID  uint
		want      int
	}{
		{"hub1 worker1", 1, 1, 52001},
		{"hub1 worker5", 1, 5, 52005},
		{"hub1 worker100", 1, 100, 52100},
		{"hub2 worker1", 2, 1, 53001},
		{"hub2 worker10", 2, 10, 53010},
		{"hub3 worker1", 3, 1, 54001},
		{"hub3 worker50", 3, 50, 54050},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hubWorkerListenPort(tt.hubNumber, tt.workerID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHubToHubListenPort(t *testing.T) {
	// Formula: 51820 + localHubNumber*10 + remoteHubNumber
	tests := []struct {
		name       string
		localHub   int
		remoteHub  int
		want       int
	}{
		{"hub1 to hub2", 1, 2, 51832},
		{"hub2 to hub1", 2, 1, 51841},
		{"hub1 to hub3", 1, 3, 51833},
		{"hub3 to hub1", 3, 1, 51851},
		{"hub2 to hub3", 2, 3, 51843},
		{"hub3 to hub2", 3, 2, 51852},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hubToHubListenPort(tt.localHub, tt.remoteHub)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
