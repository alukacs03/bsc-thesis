import { apiClient } from './client';
import type { WireGuardPeer } from '../types/wireguard';
import type { OSPFNeighbor } from '../types/ospf';

export const networkingAPI = {
  listWireGuardPeers: () => apiClient.get<WireGuardPeer[]>('/admin/network/wireguard/peers'),
  listOSPFNeighbors: () => apiClient.get<OSPFNeighbor[]>('/admin/network/ospf/neighbors'),
  listWireGuardPeersForNode: (nodeId: number) =>
    apiClient.get<WireGuardPeer[]>(`/admin/nodes/${nodeId}/network/wireguard/peers`),
  listOSPFNeighborsForNode: (nodeId: number) =>
    apiClient.get<OSPFNeighbor[]>(`/admin/nodes/${nodeId}/network/ospf/neighbors`),
};
