package models

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"unique;not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name      string `json:"name" gorm:"not null"`    // e.g., "worker-1-primary-key"
	Hash      string `json:"-" gorm:"not null"`       // bcrypt hash of the actual key
	HashIndex string `json:"-" gorm:"not null;index"` // first 8 bytes of SHA256 hash of the actual key, hex-encoded

	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}
