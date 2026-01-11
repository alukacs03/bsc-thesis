package models

import "time"

type NodeSSHAuthorizedKey struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	NodeID uint `json:"node_id" gorm:"not null;index;uniqueIndex:idx_node_user_pub,priority:1"`
	Node   Node `json:"node,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Username  string `json:"username" gorm:"not null;index;uniqueIndex:idx_node_user_pub,priority:2"`
	PublicKey string `json:"public_key" gorm:"type:text;not null;uniqueIndex:idx_node_user_pub,priority:3"`
	Comment   string `json:"comment,omitempty" gorm:"default:''"`

	CreatedByID *uint `json:"created_by_id,omitempty" gorm:"index"`
	CreatedBy   *User `json:"created_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
