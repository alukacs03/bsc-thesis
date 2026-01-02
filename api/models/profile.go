package models

import (
	"time"

	"gorm.io/datatypes"
)

type WireGuardProfile struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name    string `json:"name" gorm:"unique;not null"`
	Version string `json:"version"`

	ListenPort          int `json:"listen_port" gorm:"default:51820"`
	PersistentKeepalive int `json:"persistent_keepalive" gorm:"default:25"`
	MTU                 int `json:"mtu" gorm:"default:1420"`

	DefaultAllowedIPs datatypes.JSON `json:"default_allowed_ips"`
	Extra             datatypes.JSON `json:"extra,omitempty"`
}

type OSPFProfile struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name    string `json:"name" gorm:"unique;not null"`
	Version string `json:"version"`

	Area          string  `json:"area" gorm:"default:'0.0.0.0'"`
	HelloInterval float64 `json:"hello_interval" gorm:"default:1.0"`
	DeadInterval  float64 `json:"dead_interval" gorm:"default:3.0"`
	Cost          int     `json:"cost" gorm:"default:10"`

	PassiveInterfaces datatypes.JSON `json:"passive_interfaces,omitempty"`
	Extra             datatypes.JSON `json:"extra,omitempty"`
}
