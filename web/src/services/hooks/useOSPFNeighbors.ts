import { useCallback } from 'react';
import { networkingAPI } from '../api/networking';
import type { OSPFNeighbor } from '../types/ospf';
import { usePolling } from './usePolling';

export function useOSPFNeighbors(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => networkingAPI.listOSPFNeighbors(), []);

  return usePolling<OSPFNeighbor[]>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch OSPF neighbors:', error);
    },
  });
}

