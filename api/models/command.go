package models

import (
	"time"

	"gorm.io/datatypes"
)

type NodeCommandStatus string

const (
	NodeCommandStatusPending   NodeCommandStatus = "pending"
	NodeCommandStatusRunning   NodeCommandStatus = "running"
	NodeCommandStatusSucceeded NodeCommandStatus = "succeeded"
	NodeCommandStatusFailed    NodeCommandStatus = "failed"
)

type NodeCommand struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Kind    string             `json:"kind" gorm:"not null;index"`
	Payload datatypes.JSON     `json:"payload,omitempty"`
	Status  NodeCommandStatus  `json:"status" gorm:"not null;default:'pending';index"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Output string `json:"output,omitempty" gorm:"type:text"`
	Error  string `json:"error,omitempty" gorm:"type:text"`
}

