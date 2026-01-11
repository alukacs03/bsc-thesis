import type { SystemService } from './service';

export interface Node {
  id: number;
  created_at: string;
  updated_at: string;
  hostname: string;
  role: 'hub' | 'worker';
  public_ip: string;
  management_ip?: string;
  provider: string;
  os: string;
  labels?: Record<string, string>;
  status: 'active' | 'maintenance' | 'offline' | 'decommissioned';
  last_seen_at?: string;
  agent_version: string;
  cpu_usage?: number | null;
  memory_usage?: number | null;
  disk_usage?: number | null;
  disk_total_bytes?: number | null;
  disk_used_bytes?: number | null;
  uptime_seconds?: number | null;
  system_users?: string[] | null;
  system_services?: SystemService[] | null;
  reported_desired_role?: 'hub' | 'worker' | '' | null;
  k8s_state?: 'not_configured' | 'cluster_initialized' | 'joined_control_plane' | 'joined_worker' | 'error' | string;
  k8s_joined_at?: string | null;
  k8s_last_attempt_at?: string | null;
  k8s_last_error?: string | null;
  enrolled_by_id?: number;
  enrolled_by?: {
    id: number;
    name: string;
    email: string;
  };
  enrollment_request_id: number;
}

export interface NodeListResponse {
  nodes: Node[];
  total: number;
}
