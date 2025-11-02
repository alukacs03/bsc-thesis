package models

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name string `json:"name" gorm:"not null"`              // e.g., "worker-1-primary-key"
	Hash string `json:"-" gorm:"not null"`                 // bcrypt hash of the actual key

	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type JoinTokenScope string

const (
	JoinTokenScopeNodeJoin JoinTokenScope = "node_join"
	// Future: JoinTokenScopeUserRegistration
)

type JoinToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Scope JoinTokenScope `json:"scope" gorm:"not null"`
	Hash  string         `json:"-" gorm:"not null"` // bcrypt hash of the actual token

	ClusterID uint    `json:"cluster_id" gorm:"not null;index"`
	Cluster   Cluster `json:"cluster,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	MaxUses   int `json:"max_uses" gorm:"default:1"`   // 0 = unlimited, 1 = one-time use
	UsedCount int `json:"used_count" gorm:"default:0"` // Increment on each use

	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	CreatedByID uint `json:"created_by_id" gorm:"not null"`
	CreatedBy   User `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
