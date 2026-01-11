package models

import "time"

type DeploymentSettings struct {
	ID                     uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	LoopbackCIDR            string    `json:"loopback_cidr"`
	HubToHubCIDR            string    `json:"hub_to_hub_cidr"`
	Hub1WorkerCIDR          string    `json:"hub1_worker_cidr"`
	Hub2WorkerCIDR          string    `json:"hub2_worker_cidr"`
	Hub3WorkerCIDR          string    `json:"hub3_worker_cidr"`
	KubernetesPodCIDR       string    `json:"kubernetes_pod_cidr"`
	KubernetesServiceCIDR   string    `json:"kubernetes_service_cidr"`
	OSPFArea                int       `json:"ospf_area"`
	OSPFHelloInterval       int       `json:"ospf_hello_interval"`
	OSPFDeadInterval        int       `json:"ospf_dead_interval"`
	OSPFHubToHubCost        int       `json:"ospf_hub_to_hub_cost"`
	OSPFHubToWorkerCost     int       `json:"ospf_hub_to_worker_cost"`
	OSPFWorkerToHubCost     int       `json:"ospf_worker_to_hub_cost"`
}
