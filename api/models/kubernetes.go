package models

import "time"



type KubernetesCluster struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BootstrapNodeID *uint `json:"bootstrap_node_id,omitempty"`

	ControlPlaneEndpoint string `json:"control_plane_endpoint" gorm:"not null;default:''"`
	PodCIDR              string `json:"pod_cidr" gorm:"not null;default:'10.244.0.0/16'"`
	ServiceCIDR          string `json:"service_cidr" gorm:"not null;default:'10.96.0.0/12'"`
	KubernetesVersion    string `json:"kubernetes_version" gorm:"not null;default:'v1.29'"`

	InitializedAt *time.Time `json:"initialized_at,omitempty"`

	WorkerJoinCommand       string     `json:"worker_join_command" gorm:"not null;default:''"`
	ControlPlaneJoinCommand string     `json:"control_plane_join_command" gorm:"not null;default:''"`
	JoinCommandExpiresAt    *time.Time `json:"join_command_expires_at,omitempty"`
}

