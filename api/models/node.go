package models

import (
	"time"

	"gorm.io/datatypes"
)

type NodeRole string

const (
	NodeRoleHub    NodeRole = "hub"
	NodeRoleWorker NodeRole = "worker"
)

type NodeStatus string

const (
	NodeStatusActive         NodeStatus = "active"
	NodeStatusMaintenance    NodeStatus = "maintenance"
	NodeStatusOffline        NodeStatus = "offline"
	NodeStatusDecommissioned NodeStatus = "decommissioned"
)

type Node struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Hostname string   `json:"hostname" gorm:"not null"`
	Role     NodeRole `json:"role" gorm:"not null"`

	PublicIP     string `json:"public_ip" gorm:"not null"`
	ManagementIP string `json:"management_ip,omitempty"`
	Provider     string `json:"provider" gorm:"not null"`
	OS           string `json:"os" gorm:"not null"`

	Labels     datatypes.JSON `json:"labels,omitempty"`
	Status     NodeStatus     `json:"status" gorm:"default:'active';not null"`
	LastSeenAt *time.Time     `json:"last_seen_at,omitempty"`

	AgentVersion string   `json:"agent_version" gorm:"not null;default:''"`
	CPUUsage     *float64 `json:"cpu_usage"`
	MemoryUsage  *float64 `json:"memory_usage"`
	DiskUsage    *float64 `json:"disk_usage"`
	DiskTotalBytes *uint64 `json:"disk_total_bytes"`
	DiskUsedBytes  *uint64 `json:"disk_used_bytes"`
	UptimeSeconds  *uint64 `json:"uptime_seconds"`
	HeartbeatLogs datatypes.JSON `json:"heartbeat_logs,omitempty"`
	OSPFNeighbors datatypes.JSON `json:"ospf_neighbors,omitempty"`
	SystemUsers datatypes.JSON `json:"system_users,omitempty"`
	SystemServices datatypes.JSON `json:"system_services,omitempty"`

	
	
	ReportedDesiredRole string `json:"reported_desired_role" gorm:"not null;default:''"`

	
	K8sState         string     `json:"k8s_state" gorm:"not null;default:'not_configured'"`
	K8sJoinedAt      *time.Time `json:"k8s_joined_at,omitempty"`
	K8sLastAttemptAt *time.Time `json:"k8s_last_attempt_at,omitempty"`
	K8sLastError     string     `json:"k8s_last_error,omitempty" gorm:"not null;default:''"`

	EnrolledByID        *uint `json:"enrolled_by_id,omitempty"`
	EnrolledBy          *User `json:"enrolled_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EnrollmentRequestID uint  `json:"enrollment_request_id" gorm:"not null"`
}

type NodeEnrollmentRequest struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestedAt time.Time `json:"requested_at" gorm:"autoCreateTime"`
	Status      string    `json:"status" gorm:"default:'pending'"`

	SecretHash      string `json:"-" gorm:"not null;default:''"`
	SecretHashIndex string `json:"-" gorm:"index;not null;default:''"`

	Hostname    string   `json:"hostname" gorm:"not null"`
	PublicIP    string   `json:"public_ip" gorm:"not null"`
	Provider    string   `json:"provider" gorm:"not null"`
	OS          string   `json:"os" gorm:"not null"`
	DesiredRole NodeRole `json:"desired_role" gorm:"not null"`

	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedByID *uint      `json:"approved_by_id,omitempty"`
	ApprovedBy   *User      `json:"approved_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	RejectionReason string     `json:"rejection_reason,omitempty" gorm:"default:null"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectedByID    *uint      `json:"rejected_by_id,omitempty"`
	RejectedBy      *User      `json:"rejected_by,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	ConvertedNodeID *uint `json:"converted_node_id,omitempty"`
}
