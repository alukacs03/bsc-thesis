package models

import (
	"time"

	"gorm.io/datatypes"
)

type WireGuardProfile struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ClusterID uint    `json:"cluster_id" gorm:"not null;index"`
	Cluster   Cluster `json:"cluster,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name    string `json:"name" gorm:"unique;not null"` // e.g., "default"
	Version string `json:"version"`                     // WireGuard version compatibility

	ListenPort          int `json:"listen_port" gorm:"default:51820"`
	PersistentKeepalive int `json:"persistent_keepalive" gorm:"default:25"` // seconds
	MTU                 int `json:"mtu" gorm:"default:1420"`

	DefaultAllowedIPs datatypes.JSON `json:"default_allowed_ips"` // e.g., ["0.0.0.0/0", "::/0"]
	Extra             datatypes.JSON `json:"extra,omitempty"`     // Future custom settings
}

type OSPFProfile struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ClusterID uint    `json:"cluster_id" gorm:"not null;index"`
	Cluster   Cluster `json:"cluster,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name    string `json:"name" gorm:"unique;not null"` // e.g., "default"
	Version string `json:"version"`                     // FRR version

	Area          string  `json:"area" gorm:"default:'0.0.0.0'"` // OSPF area
	HelloInterval float64 `json:"hello_interval" gorm:"default:1.0"`
	DeadInterval  float64 `json:"dead_interval" gorm:"default:3.0"`
	Cost          int     `json:"cost" gorm:"default:10"`

	PassiveInterfaces datatypes.JSON `json:"passive_interfaces,omitempty"` // e.g., ["eth0"]
	Extra             datatypes.JSON `json:"extra,omitempty"`              // Future custom settings
}
