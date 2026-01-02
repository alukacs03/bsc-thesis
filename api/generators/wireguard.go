package generators

import (
	"fmt"
	"strings"
)

type WireGuardPeer struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          []string
	PersistentKeepalive int
}

type WireGuardInterfaceConfig struct {
	ListenPort int
	PrivateKey string
	Peers      []WireGuardPeer
}

func GenerateWireGuardConfig(listenPort int, privateKey string, peers []WireGuardPeer) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("ListenPort = %d\n", listenPort))
	if privateKey != "" {
		sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", privateKey))
	} else {
		sb.WriteString("PrivateKey = PRIVATE_KEY_PLACEHOLDER\n")
	}

	for _, peer := range peers {
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))

		if len(peer.AllowedIPs) > 0 {
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		}

		if peer.Endpoint != "" {
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint))
		}

		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
	}

	return sb.String()
}

func GenerateWireGuardConfigForWorker(listenPort int, privateKey string, hubPublicKey string, hubEndpoint string, hubListenPort int) string {
	peer := WireGuardPeer{
		PublicKey: hubPublicKey,
		Endpoint:  fmt.Sprintf("%s:%d", hubEndpoint, hubListenPort),
		AllowedIPs: []string{
			"10.255.0.0/16",
			"224.0.0.5/32",
		},
		PersistentKeepalive: 0,
	}

	return GenerateWireGuardConfig(listenPort, privateKey, []WireGuardPeer{peer})
}

func GenerateWireGuardConfigForHub(listenPort int, privateKey string, peerPublicKey string, peerEndpoint string, peerListenPort int, peerLoopbackIP string, linkSubnet string, persistentKeepalive int) string {
	allowedIPs := []string{
		peerLoopbackIP + "/32",
		linkSubnet,
		"224.0.0.5/32",
	}

	peer := WireGuardPeer{
		PublicKey:           peerPublicKey,
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: persistentKeepalive,
	}

	if peerEndpoint != "" && peerListenPort > 0 {
		peer.Endpoint = fmt.Sprintf("%s:%d", peerEndpoint, peerListenPort)
	}

	return GenerateWireGuardConfig(listenPort, privateKey, []WireGuardPeer{peer})
}

func GenerateWireGuardConfigForHubToHub(listenPort int, privateKey string, peerPublicKey string, peerEndpoint string, peerListenPort int, peerLoopbackIP string, linkSubnet string, peerWorkerRanges []string) string {
	allowedIPs := []string{
		peerLoopbackIP + "/32",
		linkSubnet,
		"224.0.0.5/32",
	}

	allowedIPs = append(allowedIPs, peerWorkerRanges...)

	peer := WireGuardPeer{
		PublicKey:           peerPublicKey,
		Endpoint:            fmt.Sprintf("%s:%d", peerEndpoint, peerListenPort),
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: 0,
	}

	return GenerateWireGuardConfig(listenPort, privateKey, []WireGuardPeer{peer})
}
