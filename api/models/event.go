package models

import (
	"time"

	"gorm.io/datatypes"
)

type EventKind string

const (
	EventKindNodeOffline      EventKind = "node_offline"
	EventKindNodeOnline       EventKind = "node_online"
	EventKindTunnelDown       EventKind = "tunnel_down"
	EventKindTunnelUp         EventKind = "tunnel_up"
	EventKindOSPFNeighborDown EventKind = "ospf_neighbor_down"
	EventKindOSPFNeighborUp   EventKind = "ospf_neighbor_up"
	EventKindIPPoolExhausted  EventKind = "ip_pool_exhausted"
	EventKindNodeDecommission EventKind = "node_decommissioned"
)

type Event struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Kind EventKind `json:"kind" gorm:"not null;index"`

	NodeID *uint `json:"node_id,omitempty" gorm:"index"`
	Node   *Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Message string         `json:"message" gorm:"not null"`
	Data    datatypes.JSON `json:"data,omitempty"`
}
