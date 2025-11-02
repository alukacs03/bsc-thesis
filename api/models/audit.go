package models

import (
	"time"

	"gorm.io/datatypes"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ActorID   *uint     `json:"actor_id"`
	Actor     *User     `json:"actor,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Action    string `json:"action" gorm:"not null"`
	Entity    string `json:"entity" gorm:"not null"`
	EntityID  uint   `json:"entity_id" gorm:"not null"`
	IP        string `json:"ip" gorm:"not null"`
	UserAgent string `json:"user_agent" gorm:"not null"`

	Details datatypes.JSON `json:"details,omitempty"`
}
