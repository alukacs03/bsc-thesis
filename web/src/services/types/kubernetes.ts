export interface KubernetesClusterSummary {
  id: number;
  created_at: string;
  updated_at: string;

  bootstrap_node_id?: number | null;

  control_plane_endpoint: string;
  pod_cidr: string;
  service_cidr: string;
  kubernetes_version: string;

  initialized_at?: string | null;
  join_command_expires_at?: string | null;
}

export interface KubernetesClusterResponse {
  cluster: KubernetesClusterSummary | null;
}

export interface KubernetesWorkloadNamespaceSummary {
  namespace: string;
  deployments_total: number;
  deployments_ready: number;
  statefulsets_total: number;
  statefulsets_ready: number;
  daemonsets_total: number;
  daemonsets_ready: number;
  jobs_total: number;
  jobs_active: number;
  jobs_succeeded: number;
  jobs_failed: number;
  pods_total: number;
  pods_running: number;
  pods_pending: number;
  pods_succeeded: number;
  pods_failed: number;
  pods_unhealthy: number;
  restarts_total: number;
}

export interface KubernetesWorkloadNodeSummary {
  node: string;
  pods: number;
  unhealthy_pods: number;
}

export interface KubernetesWorkloadPodIssue {
  namespace: string;
  name: string;
  node?: string;
  phase: string;
  reason?: string;
  message?: string;
  images?: string[];
  restarts: number;
  age_seconds: number;
}

export interface KubernetesWorkloadsResponse {
  generated_at: string;
  namespaces: KubernetesWorkloadNamespaceSummary[];
  nodes: KubernetesWorkloadNodeSummary[];
  unhealthy_pods: KubernetesWorkloadPodIssue[];
  resources: KubernetesWorkloadResource[];
}

export interface ApplyManifestResponse {
  success: boolean;
  output: string;
  error?: string;
}

export interface GetResourceYAMLResponse {
  yaml: string;
  error?: string;
}

export interface KubernetesWorkloadResource {
  namespace: string;
  name: string;
  kind: string;
  ready: string;
  images?: string[];
  age_seconds: number;
}

export interface KubernetesServiceInfo {
  namespace: string;
  name: string;
  type: string;
  cluster_ip: string;
  external_ip?: string;
  ports: string[];
  age_seconds: number;
}

export interface KubernetesIngressRule {
  host: string;
  path: string;
  path_type?: string;
  service_name: string;
  service_port: string;
}

export interface KubernetesIngressInfo {
  namespace: string;
  name: string;
  ingress_class?: string;
  tls: boolean;
  tls_hosts?: string[];
  rules: KubernetesIngressRule[];
  address?: string;
  age_seconds: number;
}

export interface KubernetesNetworkingResponse {
  generated_at: string;
  services: KubernetesServiceInfo[];
  ingresses: KubernetesIngressInfo[];
}

export interface DeleteResourceResponse {
  success: boolean;
  output: string;
  error?: string;
}
