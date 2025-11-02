package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password []byte `json:"-"`
}

type UserRegistrationRequest struct {
	// metadata
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestedAt time.Time `json:"requested_at" gorm:"autoCreateTime"`
	Status      string    `json:"status" gorm:"default:'pending'"`

	// request data
	Email    string `json:"email" gorm:"unique;not null"`
	Password []byte `json:"-" gorm:"not null"`
	FullName string `json:"full_name" gorm:"not null"`

	// decision - approved
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedByID *uint      `json:"approved_by_id,omitempty"`
	ApprovedBy   *User      `json:"approved_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// decision - rejected
	RejectionReason string     `json:"rejection_reason,omitempty" gorm:"default:null"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectedByID    *uint      `json:"rejected_by_id,omitempty"`
	RejectedBy      *User      `json:"rejected_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// if approved, the ID of the created User
	ConvertedUserID *uint `json:"converted_user_id,omitempty"`
}
