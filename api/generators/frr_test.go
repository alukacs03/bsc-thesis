package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFRRConfig(t *testing.T) {
	tests := []struct {
		name   string
		config FRRConfig
		checks func(t *testing.T, result string)
	}{
		{
			name: "hub config includes ip forwarding and router ospf section",
			config: FRRConfig{
				Hostname:   "hub1",
				RouterID:   "10.255.0.1",
				IsHub:      true,
				LoopbackIP: "10.255.0.1",
				Interfaces: []OSPFInterface{
					{Name: "dummy", IsDummy: true},
					{Name: "wg0", Cost: 10, IsPointToPoint: true, PrefixSuppression: true},
				},
				OSPFArea: 10,
			},
			checks: func(t *testing.T, result string) {
				// Hub should NOT have "no ip forwarding"
				assert.NotContains(t, result, "no ip forwarding")
				// Hub should NOT have route-map
				assert.NotContains(t, result, "route-map RM_SET_SRC")
				// Must have correct router ospf section
				assert.Contains(t, result, "router ospf")
				assert.Contains(t, result, "ospf router-id 10.255.0.1")
				assert.Contains(t, result, "passive-interface default")
				// Hub should NOT have log-adjacency-changes or max-metric
				assert.NotContains(t, result, "log-adjacency-changes")
				assert.NotContains(t, result, "max-metric router-lsa administrative")
				// Should contain hostname
				assert.Contains(t, result, "hostname hub1")
				// Should contain frr version header
				assert.Contains(t, result, "frr version 10.5.0")
				// Interface config
				assert.Contains(t, result, "interface wg0")
				assert.Contains(t, result, "ip ospf cost 10")
				assert.Contains(t, result, "ip ospf network point-to-point")
				assert.Contains(t, result, "ip ospf prefix-suppression")
				// Dummy interface
				assert.Contains(t, result, "interface dummy")
			},
		},
		{
			name: "worker config includes no ip forwarding and route-map",
			config: FRRConfig{
				Hostname:   "worker1",
				RouterID:   "10.255.0.10",
				IsHub:      false,
				LoopbackIP: "10.255.0.10",
				Interfaces: []OSPFInterface{
					{Name: "dummy", IsDummy: true},
					{
						Name:              "wg-hub1",
						Cost:              10,
						IsPointToPoint:    true,
						HelloInterval:     1,
						DeadInterval:      3,
						PrefixSuppression: true,
					},
				},
				OSPFArea: 10,
			},
			checks: func(t *testing.T, result string) {
				// Worker should have "no ip forwarding"
				assert.Contains(t, result, "no ip forwarding")
				// Worker should have route-map
				assert.Contains(t, result, "route-map RM_SET_SRC permit 10")
				assert.Contains(t, result, "set src 10.255.0.10")
				assert.Contains(t, result, "ip protocol ospf route-map RM_SET_SRC")
				// Router ospf section
				assert.Contains(t, result, "router ospf")
				assert.Contains(t, result, "ospf router-id 10.255.0.10")
				assert.Contains(t, result, "passive-interface default")
				// Worker should have log-adjacency-changes and max-metric
				assert.Contains(t, result, "log-adjacency-changes")
				assert.Contains(t, result, "max-metric router-lsa administrative")
				// Hostname
				assert.Contains(t, result, "hostname worker1")
				// Interface with hello/dead intervals
				assert.Contains(t, result, "ip ospf hello-interval 1")
				assert.Contains(t, result, "ip ospf dead-interval 3")
				assert.Contains(t, result, "ip ospf cost 10")
				assert.Contains(t, result, "ip ospf area 10")
			},
		},
		{
			name: "worker config without loopback IP omits route-map",
			config: FRRConfig{
				Hostname:   "worker-nolb",
				RouterID:   "10.255.0.20",
				IsHub:      false,
				LoopbackIP: "",
				Interfaces: []OSPFInterface{},
				OSPFArea:   10,
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "no ip forwarding")
				assert.NotContains(t, result, "route-map RM_SET_SRC")
				assert.NotContains(t, result, "ip protocol ospf route-map")
			},
		},
		{
			name: "dummy interface gets area but no cost",
			config: FRRConfig{
				Hostname:   "node1",
				RouterID:   "10.255.0.1",
				IsHub:      true,
				LoopbackIP: "10.255.0.1",
				Interfaces: []OSPFInterface{
					{Name: "dummy", IsDummy: true},
				},
				OSPFArea: 20,
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface dummy")
				assert.Contains(t, result, "ip ospf area 20")
				assert.Contains(t, result, "no ip ospf passive")
				// Dummy should NOT have cost
				assert.NotContains(t, result, "ip ospf cost")
			},
		},
		{
			name: "interface without hello interval omits timers",
			config: FRRConfig{
				Hostname:   "node2",
				RouterID:   "10.255.0.2",
				IsHub:      true,
				LoopbackIP: "10.255.0.2",
				Interfaces: []OSPFInterface{
					{Name: "wg0", Cost: 50, IsPointToPoint: false, HelloInterval: 0, DeadInterval: 0},
				},
				OSPFArea: 10,
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "ip ospf cost 50")
				assert.NotContains(t, result, "ip ospf hello-interval")
				assert.NotContains(t, result, "ip ospf dead-interval")
				assert.NotContains(t, result, "ip ospf network point-to-point")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFRRConfig(tt.config)
			tt.checks(t, result)
		})
	}
}

func TestGenerateFRRConfigForWorker(t *testing.T) {
	tests := []struct {
		name          string
		hostname      string
		loopbackIP    string
		hubInterfaces []string
		checks        func(t *testing.T, result string)
	}{
		{
			name:          "worker with two hub interfaces",
			hostname:      "worker1",
			loopbackIP:    "10.255.0.50",
			hubInterfaces: []string{"wg-hub1", "wg-hub2"},
			checks: func(t *testing.T, result string) {
				// Should contain dummy interface
				assert.Contains(t, result, "interface dummy")

				// Should contain both hub interfaces
				assert.Contains(t, result, "interface wg-hub1")
				assert.Contains(t, result, "interface wg-hub2")

				// Worker-to-hub cost (default: 10)
				assert.Contains(t, result, "ip ospf cost 10")

				// Should be point-to-point
				assert.Contains(t, result, "ip ospf network point-to-point")

				// Should have hello/dead intervals from config defaults (1 and 3)
				assert.Contains(t, result, "ip ospf hello-interval 1")
				assert.Contains(t, result, "ip ospf dead-interval 3")

				// Should have prefix suppression
				assert.Contains(t, result, "ip ospf prefix-suppression")

				// Worker-specific config
				assert.Contains(t, result, "no ip forwarding")
				assert.Contains(t, result, "route-map RM_SET_SRC permit 10")
				assert.Contains(t, result, "set src 10.255.0.50")

				// Router OSPF section
				assert.Contains(t, result, "router ospf")
				assert.Contains(t, result, "ospf router-id 10.255.0.50")
				assert.Contains(t, result, "log-adjacency-changes")
				assert.Contains(t, result, "max-metric router-lsa administrative")

				// OSPF area (default: 10)
				assert.Contains(t, result, "ip ospf area 10")

				// Hostname
				assert.Contains(t, result, "hostname worker1")
			},
		},
		{
			name:          "worker with single hub interface",
			hostname:      "worker-solo",
			loopbackIP:    "10.255.0.99",
			hubInterfaces: []string{"wg-hub1"},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface dummy")
				assert.Contains(t, result, "interface wg-hub1")
				assert.Contains(t, result, "hostname worker-solo")
				assert.Contains(t, result, "ospf router-id 10.255.0.99")
			},
		},
		{
			name:          "worker with no hub interfaces still has dummy",
			hostname:      "worker-lonely",
			loopbackIP:    "10.255.0.77",
			hubInterfaces: []string{},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface dummy")
				assert.Contains(t, result, "hostname worker-lonely")
				assert.Contains(t, result, "ospf router-id 10.255.0.77")
				// No wg interfaces
				assert.NotContains(t, result, "ip ospf cost")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFRRConfigForWorker(tt.hostname, tt.loopbackIP, tt.hubInterfaces)
			tt.checks(t, result)
		})
	}
}

func TestGenerateFRRConfigForHub(t *testing.T) {
	tests := []struct {
		name               string
		hostname           string
		loopbackIP         string
		hubToHubInterfaces []string
		workerInterfaces   []string
		checks             func(t *testing.T, result string)
	}{
		{
			name:               "hub with hub-to-hub and worker interfaces",
			hostname:           "hub1",
			loopbackIP:         "10.255.0.1",
			hubToHubInterfaces: []string{"wg-hub2", "wg-hub3"},
			workerInterfaces:   []string{"wg-w1", "wg-w2"},
			checks: func(t *testing.T, result string) {
				// Should contain dummy interface
				assert.Contains(t, result, "interface dummy")

				// Hub-to-hub interfaces with hub-to-hub cost (default: 10)
				assert.Contains(t, result, "interface wg-hub2")
				assert.Contains(t, result, "interface wg-hub3")

				// Worker interfaces with hub-to-worker cost (default: 100)
				assert.Contains(t, result, "interface wg-w1")
				assert.Contains(t, result, "interface wg-w2")

				// Hub-to-hub cost = 10, hub-to-worker cost = 100
				assert.Contains(t, result, "ip ospf cost 10")
				assert.Contains(t, result, "ip ospf cost 100")

				// Hub should NOT have "no ip forwarding"
				assert.NotContains(t, result, "no ip forwarding")

				// Hub should NOT have route-map
				assert.NotContains(t, result, "route-map RM_SET_SRC")

				// Router OSPF section
				assert.Contains(t, result, "router ospf")
				assert.Contains(t, result, "ospf router-id 10.255.0.1")
				assert.Contains(t, result, "passive-interface default")
				assert.NotContains(t, result, "log-adjacency-changes")
				assert.NotContains(t, result, "max-metric router-lsa administrative")

				// Hostname
				assert.Contains(t, result, "hostname hub1")

				// All interfaces should be point-to-point
				assert.Contains(t, result, "ip ospf network point-to-point")

				// Worker interfaces should have hello/dead intervals
				assert.Contains(t, result, "ip ospf hello-interval 1")
				assert.Contains(t, result, "ip ospf dead-interval 3")

				// All should have prefix suppression
				assert.Contains(t, result, "ip ospf prefix-suppression")

				// OSPF area
				assert.Contains(t, result, "ip ospf area 10")
			},
		},
		{
			name:               "hub with empty hub-to-hub list",
			hostname:           "hub-solo",
			loopbackIP:         "10.255.0.2",
			hubToHubInterfaces: []string{},
			workerInterfaces:   []string{"wg-w1"},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface dummy")
				assert.Contains(t, result, "interface wg-w1")
				assert.Contains(t, result, "hostname hub-solo")
				assert.Contains(t, result, "ospf router-id 10.255.0.2")
				// Hub-to-worker cost only (default: 100)
				assert.Contains(t, result, "ip ospf cost 100")
			},
		},
		{
			name:               "hub with whitespace-only hub-to-hub entries skips them",
			hostname:           "hub-ws",
			loopbackIP:         "10.255.0.3",
			hubToHubInterfaces: []string{"", "  ", "wg-hub2"},
			workerInterfaces:   []string{},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface dummy")
				assert.Contains(t, result, "interface wg-hub2")
				// Should NOT contain interfaces for empty/whitespace entries
				assert.NotContains(t, result, "interface  ")
				assert.NotContains(t, result, "interface \n")
				// Hub-to-hub cost (default: 10)
				assert.Contains(t, result, "ip ospf cost 10")
			},
		},
		{
			name:               "hub with only workers and no hub-to-hub",
			hostname:           "hub-workers-only",
			loopbackIP:         "10.255.0.4",
			hubToHubInterfaces: nil,
			workerInterfaces:   []string{"wg-w1", "wg-w2", "wg-w3"},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "interface wg-w1")
				assert.Contains(t, result, "interface wg-w2")
				assert.Contains(t, result, "interface wg-w3")
				assert.Contains(t, result, "hostname hub-workers-only")
				// Only hub-to-worker cost should be present (default: 100)
				assert.Contains(t, result, "ip ospf cost 100")
				// Hub-to-hub cost (10) should NOT appear since there are no hub-to-hub interfaces
				// (Note: we can't strictly check NotContains for cost 10 because area is also 10,
				// but we can verify no hub-to-hub interface names appear)
			},
		},
		{
			name:               "hub-to-hub interfaces have no hello/dead intervals",
			hostname:           "hub-timers",
			loopbackIP:         "10.255.0.5",
			hubToHubInterfaces: []string{"wg-hub2"},
			workerInterfaces:   []string{},
			checks: func(t *testing.T, result string) {
				// Hub-to-hub interfaces should NOT have hello/dead intervals
				// (they are only set on worker interfaces in GenerateFRRConfigForHub)
				assert.NotContains(t, result, "ip ospf hello-interval")
				assert.NotContains(t, result, "ip ospf dead-interval")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFRRConfigForHub(tt.hostname, tt.loopbackIP, tt.hubToHubInterfaces, tt.workerInterfaces)
			tt.checks(t, result)
		})
	}
}
