import { useCallback } from 'react';
import { usePolling } from './usePolling';
import { sshKeysAPI } from '../api/sshKeys';
import type { NodeSSHKey } from '../types/ssh';

export function useNodeSSHKeys(
  nodeId: number,
  options: { enabled?: boolean; pollingInterval?: number } = {}
) {
  const fetchFn = useCallback(() => sshKeysAPI.listForNode(nodeId), [nodeId]);
  return usePolling<NodeSSHKey[]>({
    fetchFn,
    interval: options.pollingInterval ?? 30000,
    enabled: options.enabled !== false && Number.isFinite(nodeId) && nodeId > 0,
    onError: (error) => console.error(`Failed to fetch node ${nodeId} SSH keys:`, error),
  });
}

