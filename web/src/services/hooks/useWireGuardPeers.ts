import { useCallback } from 'react';
import { networkingAPI } from '../api/networking';
import type { WireGuardPeer } from '../types/wireguard';
import { usePolling } from './usePolling';

export function useWireGuardPeers(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => networkingAPI.listWireGuardPeers(), []);

  return usePolling<WireGuardPeer[]>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch WireGuard peers:', error);
    },
  });
}

