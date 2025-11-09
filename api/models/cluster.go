package models

import "time"

type Cluster struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string `json:"name" gorm:"unique;not null"`
	Description string `json:"description"`

	// Kubernetes configuration (for future use, not used in MVP)
	KubernetesVersion string `json:"kubernetes_version"`
}
