package models

import "time"

type PeerStatus string

const (
	PeerStatusActive   PeerStatus = "active"
	PeerStatusDisabled PeerStatus = "disabled"
)

type InterfaceStatus string

const (
	InterfaceStatusUp   InterfaceStatus = "up"
	InterfaceStatusDown InterfaceStatus = "down"
)

// WireGuardInterface represents a single WireGuard interface on a node
// A node can have multiple interfaces (e.g., vpn01wg, vpn02wg) connecting to different hubs
type WireGuardInterface struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Interface identity
	Name string `json:"name" gorm:"not null"` // e.g., "vpn01wg", "vpn02wg"

	// WireGuard configuration
	// Note: PublicKey is the same across all interfaces on the same node (derived from same private key)
	PublicKey  string `json:"public_key" gorm:"not null"`
	Address    string `json:"address" gorm:"not null;unique"` // e.g., "172.31.253.154/30"
	ListenPort int    `json:"listen_port" gorm:"not null"`

	// Status
	Status InterfaceStatus `json:"status" gorm:"default:'down';not null"`

	// GORM will create unique constraint on (NodeID, Name) to prevent duplicate interface names per node
}

// NodePeer represents a WireGuard peer configuration on a specific interface
// Each interface typically has ONE peer (the hub it connects to)
type NodePeer struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Which interface has this peer configuration
	InterfaceID uint               `json:"interface_id" gorm:"not null;index"`
	Interface   WireGuardInterface `json:"interface,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Which node is the peer
	PeerNodeID uint `json:"peer_node_id" gorm:"not null;index"`
	PeerNode   Node `json:"peer_node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// WireGuard peer configuration (desired state)
	Endpoint            string `json:"endpoint"` // e.g., "91.227.138.14:62784" or empty for dynamic endpoints
	AllowedIPs          string `json:"allowed_ips" gorm:"not null"`
	PersistentKeepAlive int    `json:"persistent_keep_alive" gorm:"default:25"`

	// Observed state (updated from agent via Prometheus metrics)
	LastHandshakeAt *time.Time `json:"last_handshake_at,omitempty"` // Last successful handshake
	RxBytes         uint64     `json:"rx_bytes" gorm:"default:0"`    // Received bytes
	TxBytes         uint64     `json:"tx_bytes" gorm:"default:0"`    // Transmitted bytes

	// Status
	Status PeerStatus `json:"status" gorm:"default:'active';not null"`
}
