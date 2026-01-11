import { useCallback } from 'react';
import { nodesAPI } from '../api/nodes';
import { usePolling } from './usePolling';

export function useNodeLogs(
  nodeId: number,
  options: { enabled?: boolean; pollingInterval?: number; limit?: number } = {}
) {
  const fetchFn = useCallback(() => nodesAPI.logs(nodeId, options.limit ?? 200), [nodeId, options.limit]);

  return usePolling<{ window: string; logs: string[] }>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false && Number.isFinite(nodeId) && nodeId > 0,
    onError: (error) => {
      console.error(`Failed to fetch node ${nodeId} logs:`, error);
    },
  });
}
