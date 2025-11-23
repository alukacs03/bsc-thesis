package models

import (
	"time"
)

type IPPoolKind string

const (
	IPPoolKindWireGuard IPPoolKind = "wireguard"
	// Future: IPPoolKindService, IPPoolKindLoadBalancer, etc.
)

type IPPool struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Kind IPPoolKind `json:"kind" gorm:"not null"`
	CIDR string     `json:"cidr" gorm:"not null"`
}

type IPAllocation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	PoolID uint   `json:"pool_id" gorm:"not null;index"`
	Pool   IPPool `json:"pool,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Optional reference to node (for general allocations)
	NodeID *uint `json:"node_id,omitempty" gorm:"index"`
	Node   *Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Optional reference to specific WireGuard interface (for WireGuard IPs)
	InterfaceID *uint               `json:"interface_id,omitempty" gorm:"index"`
	Interface   *WireGuardInterface `json:"interface,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	IP      string `json:"value" gorm:"not null;unique"` // e.g., "10.0.0.5/32" or "172.31.253.154/30"
	Purpose string `json:"purpose" gorm:"not null"`      // e.g., "wg_interface", "dummy", "reserve"
}
