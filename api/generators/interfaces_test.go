package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNetworkInterfacesConfig(t *testing.T) {
	tests := []struct {
		name         string
		loopbackIP   string
		wgInterfaces []NetworkInterface
		checks       func(t *testing.T, result string)
	}{
		{
			name:       "loopback dummy interface is generated",
			loopbackIP: "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "auto dummy")
				assert.Contains(t, result, "iface dummy inet static")
				assert.Contains(t, result, "address 10.255.0.1/32")
				assert.Contains(t, result, "pre-up /sbin/ip link add dummy type dummy || true")
				assert.Contains(t, result, "post-down /sbin/ip link del dummy || true")
			},
		},
		{
			name:       "wireguard interface auto and iface blocks",
			loopbackIP: "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "auto wg-worker1")
				assert.Contains(t, result, "iface wg-worker1 inet static")
				assert.Contains(t, result, "address 10.255.8.0/31")
				assert.Contains(t, result, "pre-up /sbin/ip link add wg-worker1 type wireguard || true")
				assert.Contains(t, result, "pre-up /usr/bin/wg setconf wg-worker1 /etc/wireguard/wg-worker1.conf")
				assert.Contains(t, result, "post-down /sbin/ip link delete wg-worker1 || true")
			},
		},
		{
			name:       "post-up and pre-down commands included when present",
			loopbackIP: "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{
				{
					Name:           "wg-worker1",
					Address:        "10.255.8.0/31",
					WireGuardConf:  "/etc/wireguard/wg-worker1.conf",
					PostUpCommands: []string{"ip route add 10.0.0.0/8 dev wg-worker1"},
					PreDownCommands: []string{"ip route del 10.0.0.0/8 dev wg-worker1"},
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "post-up ip route add 10.0.0.0/8 dev wg-worker1")
				assert.Contains(t, result, "pre-down ip route del 10.0.0.0/8 dev wg-worker1")
			},
		},
		{
			name:       "empty post-up and pre-down commands are omitted",
			loopbackIP: "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{
				{
					Name:            "wg-worker1",
					Address:         "10.255.8.0/31",
					WireGuardConf:   "/etc/wireguard/wg-worker1.conf",
					PostUpCommands:  []string{"", "  ", ""},
					PreDownCommands: []string{"", " "},
				},
			},
			checks: func(t *testing.T, result string) {
				assert.NotContains(t, result, "post-up \n")
				assert.NotContains(t, result, "pre-down \n")
				// Should not have any post-up or pre-down lines at all
				assert.NotContains(t, result, "\tpost-up ")
				assert.NotContains(t, result, "\tpre-down ")
			},
		},
		{
			name:         "no wireguard interfaces only produces dummy block",
			loopbackIP:   "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "auto dummy")
				assert.Contains(t, result, "iface dummy inet static")
				assert.NotContains(t, result, "wireguard")
			},
		},
		{
			name:       "multiple wireguard interfaces",
			loopbackIP: "10.255.0.1/32",
			wgInterfaces: []NetworkInterface{
				{
					Name:          "wg-hub1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-hub1.conf",
				},
				{
					Name:          "wg-hub2",
					Address:       "10.255.8.2/31",
					WireGuardConf: "/etc/wireguard/wg-hub2.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "auto wg-hub1")
				assert.Contains(t, result, "iface wg-hub1 inet static")
				assert.Contains(t, result, "auto wg-hub2")
				assert.Contains(t, result, "iface wg-hub2 inet static")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNetworkInterfacesConfig(tt.loopbackIP, tt.wgInterfaces)
			tt.checks(t, result)
		})
	}
}

func TestGenerateNetworkInterfacesConfigForWorker(t *testing.T) {
	tests := []struct {
		name        string
		loopbackIP  string
		hub1Address string
		hub2Address string
		checks      func(t *testing.T, result string)
	}{
		{
			name:        "creates wg-hub1 and wg-hub2 with correct addresses",
			loopbackIP:  "10.255.0.5/32",
			hub1Address: "10.255.8.1/31",
			hub2Address: "10.255.8.3/31",
			checks: func(t *testing.T, result string) {
				// Dummy loopback
				assert.Contains(t, result, "auto dummy")
				assert.Contains(t, result, "address 10.255.0.5/32")

				// wg-hub1
				assert.Contains(t, result, "auto wg-hub1")
				assert.Contains(t, result, "iface wg-hub1 inet static")
				assert.Contains(t, result, "address 10.255.8.1/31")
				assert.Contains(t, result, "pre-up /usr/bin/wg setconf wg-hub1 /etc/wireguard/wg-hub1.conf")

				// wg-hub2
				assert.Contains(t, result, "auto wg-hub2")
				assert.Contains(t, result, "iface wg-hub2 inet static")
				assert.Contains(t, result, "address 10.255.8.3/31")
				assert.Contains(t, result, "pre-up /usr/bin/wg setconf wg-hub2 /etc/wireguard/wg-hub2.conf")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNetworkInterfacesConfigForWorker(tt.loopbackIP, tt.hub1Address, tt.hub2Address)
			tt.checks(t, result)
		})
	}
}

func TestGenerateNetworkInterfacesConfigForHub(t *testing.T) {
	tests := []struct {
		name             string
		loopbackIP       string
		hubToHubAddress  string
		otherHubName     string
		workerInterfaces []NetworkInterface
		checks           func(t *testing.T, result string)
	}{
		{
			name:            "hub-to-hub interface included when both address and name are set",
			loopbackIP:      "10.255.0.1/32",
			hubToHubAddress: "10.255.4.0/31",
			otherHubName:    "wg-hub2",
			workerInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				// Hub-to-hub interface
				assert.Contains(t, result, "auto wg-hub2")
				assert.Contains(t, result, "iface wg-hub2 inet static")
				assert.Contains(t, result, "address 10.255.4.0/31")
				assert.Contains(t, result, "pre-up /usr/bin/wg setconf wg-hub2 /etc/wireguard/wg-hub2.conf")

				// Worker interface appended after hub-to-hub
				assert.Contains(t, result, "auto wg-worker1")
				assert.Contains(t, result, "iface wg-worker1 inet static")
				assert.Contains(t, result, "address 10.255.8.0/31")
			},
		},
		{
			name:             "hub-to-hub interface omitted when address is empty",
			loopbackIP:       "10.255.0.1/32",
			hubToHubAddress:  "",
			otherHubName:     "wg-hub2",
			workerInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				assert.NotContains(t, result, "auto wg-hub2")
				assert.NotContains(t, result, "iface wg-hub2 inet static")
				// Worker interface still present
				assert.Contains(t, result, "auto wg-worker1")
			},
		},
		{
			name:             "hub-to-hub interface omitted when name is empty",
			loopbackIP:       "10.255.0.1/32",
			hubToHubAddress:  "10.255.4.0/31",
			otherHubName:     "",
			workerInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				// No hub-to-hub when name is empty
				assert.NotContains(t, result, "address 10.255.4.0/31")
				// Worker interface still present
				assert.Contains(t, result, "auto wg-worker1")
			},
		},
		{
			name:             "multiple worker interfaces appended",
			loopbackIP:       "10.255.0.1/32",
			hubToHubAddress:  "",
			otherHubName:     "",
			workerInterfaces: []NetworkInterface{
				{
					Name:          "wg-worker1",
					Address:       "10.255.8.0/31",
					WireGuardConf: "/etc/wireguard/wg-worker1.conf",
				},
				{
					Name:          "wg-worker2",
					Address:       "10.255.8.2/31",
					WireGuardConf: "/etc/wireguard/wg-worker2.conf",
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "auto wg-worker1")
				assert.Contains(t, result, "auto wg-worker2")
				assert.Contains(t, result, "address 10.255.8.0/31")
				assert.Contains(t, result, "address 10.255.8.2/31")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNetworkInterfacesConfigForHub(tt.loopbackIP, tt.hubToHubAddress, tt.otherHubName, tt.workerInterfaces)
			tt.checks(t, result)
		})
	}
}

func TestNewWorkerInterface(t *testing.T) {
	tests := []struct {
		name             string
		workerHostname   string
		address          string
		expectedName     string
		expectedConfPath string
	}{
		{
			name:             "creates interface with wg- prefix",
			workerHostname:   "worker1",
			address:          "10.255.8.0/31",
			expectedName:     "wg-worker1",
			expectedConfPath: "/etc/wireguard/wg-worker1.conf",
		},
		{
			name:             "handles hyphenated hostname",
			workerHostname:   "eu-west-node3",
			address:          "10.255.8.4/31",
			expectedName:     "wg-eu-west-node3",
			expectedConfPath: "/etc/wireguard/wg-eu-west-node3.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := NewWorkerInterface(tt.workerHostname, tt.address)
			assert.Equal(t, tt.expectedName, iface.Name)
			assert.Equal(t, tt.address, iface.Address)
			assert.Equal(t, tt.expectedConfPath, iface.WireGuardConf)
		})
	}
}
