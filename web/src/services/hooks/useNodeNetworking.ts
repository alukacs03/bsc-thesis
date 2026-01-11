import { useCallback } from 'react';
import { networkingAPI } from '../api/networking';
import type { WireGuardPeer } from '../types/wireguard';
import type { OSPFNeighbor } from '../types/ospf';
import { usePolling } from './usePolling';

export function useNodeWireGuardPeers(
  nodeId: number,
  options: { enabled?: boolean; pollingInterval?: number } = {}
) {
  const fetchFn = useCallback(() => networkingAPI.listWireGuardPeersForNode(nodeId), [nodeId]);
  return usePolling<WireGuardPeer[]>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false && Number.isFinite(nodeId) && nodeId > 0,
    onError: (error) => console.error(`Failed to fetch node ${nodeId} WG peers:`, error),
  });
}

export function useNodeOSPFNeighbors(
  nodeId: number,
  options: { enabled?: boolean; pollingInterval?: number } = {}
) {
  const fetchFn = useCallback(() => networkingAPI.listOSPFNeighborsForNode(nodeId), [nodeId]);
  return usePolling<OSPFNeighbor[]>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false && Number.isFinite(nodeId) && nodeId > 0,
    onError: (error) => console.error(`Failed to fetch node ${nodeId} OSPF neighbors:`, error),
  });
}

