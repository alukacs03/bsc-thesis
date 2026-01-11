import { usePolling } from './usePolling';
import { nodesAPI } from '../api/nodes';
import type { Node } from '../types/node';
import { useCallback } from 'react';

export interface UseNodesOptions {
  pollingInterval?: number;
  enabled?: boolean;
}

export function useNodes(options: UseNodesOptions = {}) {
  const fetchFn = useCallback(() => nodesAPI.list(), []);

  return usePolling<Node[]>({
    fetchFn,
    interval: options.pollingInterval ?? 10000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch nodes:', error);
    },
  });
}

export function useNode(id: number, options: UseNodesOptions = {}) {
  const fetchFn = useCallback(() => nodesAPI.get(id), [id]);

  return usePolling<Node>({
    fetchFn,
    interval: options.pollingInterval ?? 10000,
    enabled: options.enabled !== false && Number.isFinite(id) && id > 0,
    onError: (error) => {
      console.error(`Failed to fetch node ${id}:`, error);
    },
  });
}
