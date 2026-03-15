package generators

import (
	"gluon-api/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Setenv("GLUON_SECRET_KEY", "test-secret-key")
	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestGenerateWireGuardConfig(t *testing.T) {
	tests := []struct {
		name       string
		listenPort int
		privateKey string
		peers      []WireGuardPeer
		checks     func(t *testing.T, result string)
	}{
		{
			name:       "basic config with one peer",
			listenPort: 51820,
			privateKey: "testPrivateKey123",
			peers: []WireGuardPeer{
				{
					PublicKey:  "peerPubKey1",
					Endpoint:   "1.2.3.4:51820",
					AllowedIPs: []string{"10.0.0.0/24"},
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "[Interface]")
				assert.Contains(t, result, "ListenPort = 51820")
				assert.Contains(t, result, "PrivateKey = testPrivateKey123")
				assert.Contains(t, result, "[Peer]")
				assert.Contains(t, result, "PublicKey = peerPubKey1")
				assert.Contains(t, result, "Endpoint = 1.2.3.4:51820")
				assert.Contains(t, result, "AllowedIPs = 10.0.0.0/24")
			},
		},
		{
			name:       "config with two peers",
			listenPort: 51821,
			privateKey: "myKey",
			peers: []WireGuardPeer{
				{
					PublicKey:  "peerA",
					Endpoint:   "10.0.0.1:51820",
					AllowedIPs: []string{"10.1.0.0/24"},
				},
				{
					PublicKey:  "peerB",
					Endpoint:   "10.0.0.2:51821",
					AllowedIPs: []string{"10.2.0.0/24", "10.3.0.0/24"},
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "ListenPort = 51821")
				assert.Contains(t, result, "PublicKey = peerA")
				assert.Contains(t, result, "PublicKey = peerB")
				assert.Contains(t, result, "AllowedIPs = 10.2.0.0/24, 10.3.0.0/24")
			},
		},
		{
			name:       "config with zero peers",
			listenPort: 51820,
			privateKey: "soloKey",
			peers:      []WireGuardPeer{},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "[Interface]")
				assert.Contains(t, result, "ListenPort = 51820")
				assert.Contains(t, result, "PrivateKey = soloKey")
				assert.NotContains(t, result, "[Peer]")
			},
		},
		{
			name:       "peer with empty endpoint",
			listenPort: 51820,
			privateKey: "key1",
			peers: []WireGuardPeer{
				{
					PublicKey:  "pubNoEndpoint",
					Endpoint:   "",
					AllowedIPs: []string{"10.0.0.0/24"},
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "PublicKey = pubNoEndpoint")
				assert.NotContains(t, result, "Endpoint =")
			},
		},
		{
			name:       "peer with PersistentKeepalive",
			listenPort: 51820,
			privateKey: "key2",
			peers: []WireGuardPeer{
				{
					PublicKey:           "pubWithKeepalive",
					Endpoint:            "5.5.5.5:51820",
					AllowedIPs:          []string{"10.0.0.0/24"},
					PersistentKeepalive: 25,
				},
			},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "PersistentKeepalive = 25")
			},
		},
		{
			name:       "empty private key uses placeholder",
			listenPort: 51820,
			privateKey: "",
			peers:      []WireGuardPeer{},
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "PrivateKey = PRIVATE_KEY_PLACEHOLDER")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateWireGuardConfig(tt.listenPort, tt.privateKey, tt.peers)
			tt.checks(t, result)
		})
	}
}

func TestGenerateWireGuardConfigForWorker(t *testing.T) {
	result := GenerateWireGuardConfigForWorker(51820, "workerPrivKey", "hubPubKey123", "203.0.113.1", 51820)

	assert.Contains(t, result, "[Interface]")
	assert.Contains(t, result, "ListenPort = 51820")
	assert.Contains(t, result, "PrivateKey = workerPrivKey")

	// Must contain a [Peer] section with the hub's public key
	assert.Contains(t, result, "[Peer]")
	assert.Contains(t, result, "PublicKey = hubPubKey123")

	// Endpoint should be hub's address:port
	assert.Contains(t, result, "Endpoint = 203.0.113.1:51820")

	// AllowedIPs must include the loopback CIDR from config
	assert.Contains(t, result, config.Current().LoopbackCIDR)

	// AllowedIPs must include OSPF multicast address
	assert.Contains(t, result, "224.0.0.5/32")
}

func TestGenerateWireGuardConfigForHub(t *testing.T) {
	tests := []struct {
		name                string
		listenPort          int
		privateKey          string
		peerPublicKey       string
		peerEndpoint        string
		peerListenPort      int
		peerLoopbackIP      string
		linkSubnet          string
		persistentKeepalive int
		checks              func(t *testing.T, result string)
	}{
		{
			name:                "basic hub config",
			listenPort:          51820,
			privateKey:          "hubPrivKey",
			peerPublicKey:       "workerPubKey",
			peerEndpoint:        "198.51.100.5",
			peerListenPort:      51820,
			peerLoopbackIP:      "10.255.0.2",
			linkSubnet:          "10.255.8.0/31",
			persistentKeepalive: 25,
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "[Interface]")
				assert.Contains(t, result, "ListenPort = 51820")
				assert.Contains(t, result, "PrivateKey = hubPrivKey")
				assert.Contains(t, result, "[Peer]")
				assert.Contains(t, result, "PublicKey = workerPubKey")
				assert.Contains(t, result, "Endpoint = 198.51.100.5:51820")

				// AllowedIPs must contain peer loopback/32
				assert.Contains(t, result, "10.255.0.2/32")
				// AllowedIPs must contain link subnet
				assert.Contains(t, result, "10.255.8.0/31")
				// AllowedIPs must contain OSPF multicast
				assert.Contains(t, result, "224.0.0.5/32")

				assert.Contains(t, result, "PersistentKeepalive = 25")
			},
		},
		{
			name:                "empty endpoint omits Endpoint line",
			listenPort:          51820,
			privateKey:          "hubPrivKey2",
			peerPublicKey:       "workerPubKey2",
			peerEndpoint:        "",
			peerListenPort:      0,
			peerLoopbackIP:      "10.255.0.3",
			linkSubnet:          "10.255.8.2/31",
			persistentKeepalive: 0,
			checks: func(t *testing.T, result string) {
				assert.Contains(t, result, "PublicKey = workerPubKey2")
				// With empty endpoint and port=0, no Endpoint line should appear
				assert.NotContains(t, result, "Endpoint =")
				// AllowedIPs should still be present
				assert.Contains(t, result, "10.255.0.3/32")
				assert.Contains(t, result, "10.255.8.2/31")
				assert.Contains(t, result, "224.0.0.5/32")
				// No keepalive with value 0
				assert.NotContains(t, result, "PersistentKeepalive")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateWireGuardConfigForHub(
				tt.listenPort, tt.privateKey, tt.peerPublicKey,
				tt.peerEndpoint, tt.peerListenPort, tt.peerLoopbackIP,
				tt.linkSubnet, tt.persistentKeepalive,
			)
			tt.checks(t, result)
		})
	}
}

func TestGenerateWireGuardConfigForHubToHub(t *testing.T) {
	result := GenerateWireGuardConfigForHubToHub(
		51820,
		"hub1PrivKey",
		"hub2PubKey",
		"203.0.113.10",
		51821,
		"10.255.0.5",
		"10.255.4.0/31",
		[]string{"10.255.12.0/22", "10.255.16.0/22"},
	)

	assert.Contains(t, result, "[Interface]")
	assert.Contains(t, result, "ListenPort = 51820")
	assert.Contains(t, result, "PrivateKey = hub1PrivKey")

	assert.Contains(t, result, "[Peer]")
	assert.Contains(t, result, "PublicKey = hub2PubKey")
	assert.Contains(t, result, "Endpoint = 203.0.113.10:51821")

	// AllowedIPs must contain peer loopback/32
	assert.Contains(t, result, "10.255.0.5/32")
	// AllowedIPs must contain link subnet
	assert.Contains(t, result, "10.255.4.0/31")
	// AllowedIPs must contain OSPF multicast
	assert.Contains(t, result, "224.0.0.5/32")

	// Worker ranges must be appended to AllowedIPs
	assert.Contains(t, result, "10.255.12.0/22")
	assert.Contains(t, result, "10.255.16.0/22")
}
