export interface IPPool {
  id: number;
  created_at: string;
  updated_at: string;
  kind: 'wireguard';
  purpose: 'loopback' | 'hub_to_hub' | 'hub1_worker' | 'hub2_worker';
  cidr: string;
  hub_number?: number;
}

export interface IPAllocation {
  id: number;
  created_at: string;
  pool_id: number;
  pool?: IPPool;
  node_id?: number;
  node?: {
    id: number;
    hostname: string;
  };
  interface_id?: number;
  value: string; // The IP address
  purpose: string;
}

export interface CreateIPPoolRequest {
  kind: 'wireguard';
  purpose: 'loopback' | 'hub_to_hub' | 'hub1_worker' | 'hub2_worker';
  cidr: string;
  hub_number?: number;
}

export interface AllocateIPRequest {
  pool_id: number;
  node_id?: number;
  interface_id?: number;
  value: string;
  purpose: string;
}

export interface NextAvailableIPResponse {
  next_ip: string;
  pool_id: number;
}
