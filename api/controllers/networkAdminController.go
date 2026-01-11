package controllers

import (
	"encoding/json"
	"fmt"
	"gluon-api/database"
	"gluon-api/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type wireGuardPeerView struct {
	ID             uint       `json:"id"`
	LocalNodeID      uint       `json:"local_node_id"`
	LocalNodeHostname string     `json:"local_node_hostname"`
	LocalInterfaceName string    `json:"local_interface_name"`
	LocalPublicKey    string     `json:"local_public_key"`
	LocalEndpoint     string     `json:"local_endpoint"`

	PeerNodeID      uint       `json:"peer_node_id"`
	PeerHostname    string     `json:"peer_hostname"`
	PeerInterfaceName string    `json:"peer_interface_name"`
	PeerPublicKey   string     `json:"peer_public_key"`
	PeerEndpoint    string     `json:"peer_endpoint"`
	AllowedIPs      string     `json:"allowed_ips"`
	LastHandshakeAt *time.Time `json:"last_handshake_at,omitempty"`
	RxBytes         uint64     `json:"rx_bytes"`
	TxBytes         uint64     `json:"tx_bytes"`
	Status          string     `json:"status"`
	UIStatus        string     `json:"ui_status"`
}

func ListWireGuardPeers(c *fiber.Ctx) error {
	return listWireGuardPeers(c, nil)
}

func ListWireGuardPeersForNode(c *fiber.Ctx) error {
	id := c.Params("id")
	return listWireGuardPeers(c, &id)
}

func listWireGuardPeers(c *fiber.Ctx, nodeID *string) error {
	var peers []models.NodePeer
	q := database.DB.
		Preload("Interface").
		Preload("Interface.Node").
		Preload("PeerNode")
	if nodeID != nil {
		q = q.Joins("JOIN wire_guard_interfaces wgi ON wgi.id = node_peers.interface_id").
			Where("wgi.node_id = ?", *nodeID)
	}

	if err := q.Find(&peers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve WireGuard peers",
		})
	}

	now := time.Now()
	out := make([]wireGuardPeerView, 0, len(peers))
	for _, p := range peers {
		uiStatus := "unknown"
		if p.Status == models.PeerStatusDisabled {
			uiStatus = "down"
		} else if p.LastHandshakeAt != nil {
			age := now.Sub(*p.LastHandshakeAt)
			if age <= 3*time.Minute+5*time.Second {
				uiStatus = "connected"
			} else if age < 4*time.Minute {
				uiStatus = "potentially_failing"
			} else {
				uiStatus = "down"
			}
		}

		localHostname := ""
		localEndpoint := ""
		localPublicKey := ""
		if p.Interface.Node.ID != 0 {
			localHostname = p.Interface.Node.Hostname
			if p.Interface.Node.PublicIP != "" && p.Interface.ListenPort != 0 {
				localEndpoint = p.Interface.Node.PublicIP + ":" + fmt.Sprint(p.Interface.ListenPort)
			}
		}
		if p.Interface.PublicKey != "" {
			localPublicKey = p.Interface.PublicKey
		}

		peerHostname := ""
		if p.PeerNode.ID != 0 {
			peerHostname = p.PeerNode.Hostname
		}

		peerIfaceName := ""
		if p.PeerNodeID != 0 && p.PeerPublicKey != "" {
			var peerIface models.WireGuardInterface
			if err := database.DB.
				Where("node_id = ? AND public_key = ?", p.PeerNodeID, p.PeerPublicKey).
				First(&peerIface).Error; err == nil {
				peerIfaceName = peerIface.Name
			}
		}

		out = append(out, wireGuardPeerView{
			ID:                p.ID,
			LocalNodeID:        p.Interface.NodeID,
			LocalNodeHostname:  localHostname,
			LocalInterfaceName: p.Interface.Name,
			LocalPublicKey:     localPublicKey,
			LocalEndpoint:      localEndpoint,
			PeerNodeID:         p.PeerNodeID,
			PeerHostname:       peerHostname,
			PeerInterfaceName:  peerIfaceName,
			PeerPublicKey:      p.PeerPublicKey,
			PeerEndpoint:       p.Endpoint,
			AllowedIPs:         p.AllowedIPs,
			LastHandshakeAt:    p.LastHandshakeAt,
			RxBytes:            p.RxBytes,
			TxBytes:            p.TxBytes,
			Status:             string(p.Status),
			UIStatus:           uiStatus,
		})
	}

	return c.JSON(out)
}

type ospfNeighborView struct {
	NodeID               uint    `json:"node_id"`
	NodeHostname         string  `json:"node_hostname"`
	RouterID             string  `json:"router_id"`
	Area                 string  `json:"area"`
	State                string  `json:"state"`
	Interface            string  `json:"interface"`
	HelloIntervalSeconds *uint64 `json:"hello_interval_seconds"`
	DeadIntervalSeconds  *uint64 `json:"dead_interval_seconds"`
	Cost                 *uint64 `json:"cost"`
	Priority             *uint64 `json:"priority"`
}

func ListOSPFNeighbors(c *fiber.Ctx) error {
	var nodes []models.Node
	if err := database.DB.Select("id", "hostname", "ospf_neighbors").Find(&nodes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve nodes",
		})
	}

	out := make([]ospfNeighborView, 0)
	for _, node := range nodes {
		if len(node.OSPFNeighbors) == 0 {
			continue
		}

		var neighbors []struct {
			RouterID             string  `json:"router_id"`
			Area                 string  `json:"area"`
			State                string  `json:"state"`
			Interface            string  `json:"interface"`
			HelloIntervalSeconds *uint64 `json:"hello_interval_seconds"`
			DeadIntervalSeconds  *uint64 `json:"dead_interval_seconds"`
			Cost                 *uint64 `json:"cost"`
			Priority             *uint64 `json:"priority"`
		}
		if err := json.Unmarshal(node.OSPFNeighbors, &neighbors); err != nil {
			continue
		}
		for _, n := range neighbors {
			out = append(out, ospfNeighborView{
				NodeID:               node.ID,
				NodeHostname:         node.Hostname,
				RouterID:             n.RouterID,
				Area:                 n.Area,
				State:                n.State,
				Interface:            n.Interface,
				HelloIntervalSeconds: n.HelloIntervalSeconds,
				DeadIntervalSeconds:  n.DeadIntervalSeconds,
				Cost:                 n.Cost,
				Priority:             n.Priority,
			})
		}
	}

	return c.JSON(out)
}

func ListOSPFNeighborsForNode(c *fiber.Ctx) error {
	id := c.Params("id")
	var node models.Node
	if err := database.DB.Select("id", "hostname", "ospf_neighbors").First(&node, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	out := make([]ospfNeighborView, 0)
	if len(node.OSPFNeighbors) == 0 {
		return c.JSON(out)
	}

	var neighbors []struct {
		RouterID             string  `json:"router_id"`
		Area                 string  `json:"area"`
		State                string  `json:"state"`
		Interface            string  `json:"interface"`
		HelloIntervalSeconds *uint64 `json:"hello_interval_seconds"`
		DeadIntervalSeconds  *uint64 `json:"dead_interval_seconds"`
		Cost                 *uint64 `json:"cost"`
		Priority             *uint64 `json:"priority"`
	}
	if err := json.Unmarshal(node.OSPFNeighbors, &neighbors); err != nil {
		return c.JSON(out)
	}

	for _, n := range neighbors {
		out = append(out, ospfNeighborView{
			NodeID:               node.ID,
			NodeHostname:         node.Hostname,
			RouterID:             n.RouterID,
			Area:                 n.Area,
			State:                n.State,
			Interface:            n.Interface,
			HelloIntervalSeconds: n.HelloIntervalSeconds,
			DeadIntervalSeconds:  n.DeadIntervalSeconds,
			Cost:                 n.Cost,
			Priority:             n.Priority,
		})
	}
	return c.JSON(out)
}
