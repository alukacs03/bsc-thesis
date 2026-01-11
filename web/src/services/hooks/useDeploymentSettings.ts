import { useCallback } from 'react';
import { settingsAPI } from '../api/settings';
import type { DeploymentSettings } from '../types/settings';
import { usePolling } from './usePolling';

export function useDeploymentSettings(options: { enabled?: boolean; pollingInterval?: number } = {}) {
  const fetchFn = useCallback(() => settingsAPI.getDeploymentSettings(), []);

  return usePolling<DeploymentSettings>({
    fetchFn,
    interval: options.pollingInterval ?? 60000,
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch deployment settings:', error);
    },
  });
}
