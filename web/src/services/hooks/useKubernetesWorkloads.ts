import { useCallback } from 'react';
import { kubernetesAPI } from '../api/kubernetes';
import type { KubernetesWorkloadsResponse } from '../types/kubernetes';
import { usePolling } from './usePolling';

export function useKubernetesWorkloads(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => kubernetesAPI.getWorkloads(), []);

  return usePolling<KubernetesWorkloadsResponse>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch Kubernetes workloads:', error);
    },
  });
}

