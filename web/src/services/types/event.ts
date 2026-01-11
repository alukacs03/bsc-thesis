export type EventKind =
  | 'node_offline'
  | 'node_online'
  | 'tunnel_down'
  | 'tunnel_up'
  | 'ospf_neighbor_down'
  | 'ospf_neighbor_up'
  | 'ip_pool_exhausted'
  | 'node_decommissioned';

export interface Event {
  id: number;
  created_at: string;
  kind: EventKind;
  node_id?: number;
  message: string;
  data?: unknown;
}

