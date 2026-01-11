import { usePolling } from './usePolling';
import { ipamAPI } from '../api/ipam';
import type { IPPool, IPAllocation } from '../types/ipam';

export function useIPPools(pollingInterval = 30000) {
  return usePolling<IPPool[]>({
    fetchFn: ipamAPI.pools.list,
    interval: pollingInterval,
    onError: (error) => {
      console.error('Failed to fetch IP pools:', error);
    },
  });
}

export function useIPAllocations(pollingInterval = 30000) {
  return usePolling<IPAllocation[]>({
    fetchFn: ipamAPI.allocations.list,
    interval: pollingInterval,
    onError: (error) => {
      console.error('Failed to fetch IP allocations:', error);
    },
  });
}
