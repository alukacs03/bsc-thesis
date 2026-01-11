import { useCallback } from 'react';
import { kubernetesAPI } from '../api/kubernetes';
import type { KubernetesNetworkingResponse } from '../types/kubernetes';
import { usePolling } from './usePolling';

export function useKubernetesNetworking(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => kubernetesAPI.getNetworking(), []);

  return usePolling<KubernetesNetworkingResponse>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch Kubernetes networking:', error);
    },
  });
}
