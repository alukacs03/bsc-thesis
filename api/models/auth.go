package models

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"unique;not null;index"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Name      string `json:"name" gorm:"not null"`
	Hash      string `json:"-" gorm:"not null"`
	HashIndex string `json:"-" gorm:"not null;index"`

	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}
