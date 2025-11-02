package models

import "time"

type IPPoolKind string

const (
	IPPoolKindWireGuard IPPoolKind = "wireguard"
	// Future: IPPoolKindService, IPPoolKindLoadBalancer, etc.
)

type IPPool struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ClusterID uint    `json:"cluster_id" gorm:"not null;index"`
	Cluster   Cluster `json:"cluster,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Kind     IPPoolKind `json:"kind" gorm:"not null"`
	CIDR     string     `json:"cidr" gorm:"not null"`
	NextHint string     `json:"next_hint" gorm:"not null"` // Optimization: next IP to try allocating
}

type IPAllocation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	PoolID uint   `json:"pool_id" gorm:"not null;index"`
	Pool   IPPool `json:"pool,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	NodeID *uint `json:"node_id,omitempty" gorm:"index"` // Nullable for reserved IPs
	Node   *Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Value   string `json:"value" gorm:"not null;unique"` // e.g., "10.0.0.5/32"
	Purpose string `json:"purpose" gorm:"not null"`      // e.g., "wg_ip", "reserve"
}
