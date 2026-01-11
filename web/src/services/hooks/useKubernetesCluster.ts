import { useCallback } from 'react';
import { kubernetesAPI } from '../api/kubernetes';
import type { KubernetesClusterResponse } from '../types/kubernetes';
import { usePolling } from './usePolling';

export function useKubernetesCluster(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => kubernetesAPI.getCluster(), []);

  return usePolling<KubernetesClusterResponse>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch Kubernetes cluster:', error);
    },
  });
}

