package models

import "time"

type PeerStatus string

const (
	PeerStatusActive   PeerStatus = "active"
	PeerStatusDisabled PeerStatus = "disabled"
)

// NodePeer represents a WireGuard peer relationship between two nodes
// Example: If Node A (hub) has Node B (worker) as a peer, there will be a NodePeer record
// with NodeID = A and PeerNodeID = B
type NodePeer struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Which node has this peer configuration
	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Which node is the peer
	PeerNodeID uint `json:"peer_node_id" gorm:"not null;index"`
	PeerNode   Node `json:"peer_node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// WireGuard peer configuration (desired state)
	Endpoint            string `json:"endpoint"` // e.g., "1.2.3.4:51820" or empty for dynamic endpoints
	AllowedIPs          string `json:"allowed_ips" gorm:"not null"`
	PersistentKeepAlive int    `json:"persistent_keep_alive" gorm:"default:25"`

	// Observed state (updated from agent heartbeats via 'wg show')
	LastHandshakeAt *time.Time `json:"last_handshake_at,omitempty"` // Last successful handshake
	RxBytes         uint64     `json:"rx_bytes" gorm:"default:0"`    // Received bytes
	TxBytes         uint64     `json:"tx_bytes" gorm:"default:0"`    // Transmitted bytes

	// Status
	Status PeerStatus `json:"status" gorm:"default:'active';not null"`
}
