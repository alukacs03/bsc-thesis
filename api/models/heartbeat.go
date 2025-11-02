package models

import (
	"time"

	"gorm.io/datatypes"
)

// Heartbeat contains moderate node health information, sent frequently (30-60 seconds)
type Heartbeat struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"` // Received timestamp

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	AgentVersion string `json:"agent_version" gorm:"not null"` // e.g., "v1.0.0"

	// Simple health checks
	WireGuardOK bool `json:"wireguard_ok"` // Is WireGuard interface up?
	OSPFOK      bool `json:"ospf_ok"`      // Is OSPF running and has neighbors?
	KubeletOK   bool `json:"kubelet_ok"`   // Is kubelet running? (future)

	Details datatypes.JSON `json:"details,omitempty"` // Lightweight extra data
}

// HeartbeatDetail contains full diagnostic information, sent less frequently (5-10 minutes)
type HeartbeatDetail struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Full diagnostic data in JSON format
	WGPeersSummary datatypes.JSON `json:"wg_peers_summary,omitempty"` // Full peer stats from 'wg show'
	OSPFNeighbors  datatypes.JSON `json:"ospf_neighbors,omitempty"`   // Full OSPF neighbor list
	SystemMetrics  datatypes.JSON `json:"system_metrics,omitempty"`   // CPU, memory, disk, network
}
