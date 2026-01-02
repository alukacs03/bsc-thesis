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

type WireGuardInterface struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name string `json:"name" gorm:"not null"`

	PublicKey  string `json:"public_key" gorm:"not null"`
	Address    string `json:"address" gorm:"not null;unique"`
	ListenPort int    `json:"listen_port" gorm:"not null"`

	Status InterfaceStatus `json:"status" gorm:"default:'down';not null"`

}

type NodePeer struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	InterfaceID uint               `json:"interface_id" gorm:"not null;index"`
	Interface   WireGuardInterface `json:"interface,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	PeerNodeID uint `json:"peer_node_id" gorm:"not null;index"`
	PeerNode   Node `json:"peer_node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	PeerPublicKey       string `json:"peer_public_key"`
	Endpoint            string `json:"endpoint"`
	AllowedIPs          string `json:"allowed_ips" gorm:"not null"`
	PersistentKeepAlive int    `json:"persistent_keep_alive" gorm:"default:25"`

	LastHandshakeAt *time.Time `json:"last_handshake_at,omitempty"`
	RxBytes         uint64     `json:"rx_bytes" gorm:"default:0"`
	TxBytes         uint64     `json:"tx_bytes" gorm:"default:0"`

	Status PeerStatus `json:"status" gorm:"default:'active';not null"`
}
