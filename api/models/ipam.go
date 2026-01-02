package models

import (
	"time"
)

type IPPoolKind string

const (
	IPPoolKindWireGuard IPPoolKind = "wireguard"
)

type IPPoolPurpose string

const (
	IPPoolPurposeLoopback   IPPoolPurpose = "loopback"
	IPPoolPurposeHubToHub   IPPoolPurpose = "hub_to_hub"
	IPPoolPurposeHub1Worker IPPoolPurpose = "hub1_worker"
	IPPoolPurposeHub2Worker IPPoolPurpose = "hub2_worker"
)

type IPPool struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Kind    IPPoolKind    `json:"kind" gorm:"not null"`
	Purpose IPPoolPurpose `json:"purpose" gorm:"not null"`
	CIDR    string        `json:"cidr" gorm:"not null;unique"`

	HubNumber *int `json:"hub_number,omitempty"`
}

type IPAllocation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	PoolID uint   `json:"pool_id" gorm:"not null;index"`
	Pool   IPPool `json:"pool,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	NodeID *uint `json:"node_id,omitempty" gorm:"index"`
	Node   *Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	InterfaceID *uint               `json:"interface_id,omitempty" gorm:"index"`
	Interface   *WireGuardInterface `json:"interface,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	IP      string `json:"value" gorm:"not null;unique"`
	Purpose string `json:"purpose" gorm:"not null"`
}

type LinkAllocation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	PoolID uint   `json:"pool_id" gorm:"not null;index"`
	Pool   IPPool `json:"pool,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	NodeAID uint `json:"node_a_id" gorm:"not null;index"`
	NodeA   Node `json:"node_a,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:NodeAID"`
	NodeBID uint `json:"node_b_id" gorm:"not null;index"`
	NodeB   Node `json:"node_b,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:NodeBID"`

	Subnet string `json:"subnet" gorm:"not null;unique"`

	NodeAIP string `json:"node_a_ip" gorm:"not null"`
	NodeBIP string `json:"node_b_ip" gorm:"not null"`
}

type NodeConfig struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Version int `json:"version" gorm:"not null;default:1"`

	WireGuardConfigs       string `json:"wireguard_configs" gorm:"type:text"`
	NetworkInterfaceConfig string `json:"network_interface_config" gorm:"type:text"`
	FRRConfig              string `json:"frr_config" gorm:"type:text"`

	Hash string `json:"hash" gorm:"not null"`

	GeneratedAt time.Time  `json:"generated_at"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
}
