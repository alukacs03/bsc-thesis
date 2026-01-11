export interface DeploymentSettings {
  id: number;
  loopback_cidr: string;
  hub_to_hub_cidr: string;
  hub1_worker_cidr: string;
  hub2_worker_cidr: string;
  hub3_worker_cidr: string;
  kubernetes_pod_cidr: string;
  kubernetes_service_cidr: string;
  ospf_area: number;
  ospf_hello_interval: number;
  ospf_dead_interval: number;
  ospf_hub_to_hub_cost: number;
  ospf_hub_to_worker_cost: number;
  ospf_worker_to_hub_cost: number;
}

export interface DeploymentSettingsUpdate {
  loopback_cidr: string;
  hub_to_hub_cidr: string;
  hub1_worker_cidr: string;
  hub2_worker_cidr: string;
  hub3_worker_cidr: string;
  kubernetes_pod_cidr: string;
  kubernetes_service_cidr: string;
  ospf_area: number;
  ospf_hello_interval: number;
  ospf_dead_interval: number;
  ospf_hub_to_hub_cost: number;
  ospf_hub_to_worker_cost: number;
  ospf_worker_to_hub_cost: number;
  rebuild?: boolean;
}
