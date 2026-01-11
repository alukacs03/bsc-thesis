export type WireGuardPeerUIStatus = 'connected' | 'potentially_failing' | 'down' | 'unknown';

export interface WireGuardPeer {
  id: number;
  local_node_id: number;
  local_node_hostname: string;
  local_interface_name: string;
  local_public_key: string;
  local_endpoint: string;
  peer_node_id: number;
  peer_hostname: string;
  peer_interface_name: string;
  peer_public_key: string;
  peer_endpoint: string;
  allowed_ips: string;
  last_handshake_at?: string;
  rx_bytes: number;
  tx_bytes: number;
  status: string;
  ui_status: WireGuardPeerUIStatus;
}
