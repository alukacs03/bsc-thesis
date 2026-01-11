import { usePolling } from './usePolling';
import { enrollmentsAPI } from '../api/enrollments';
import type { NodeEnrollmentRequest } from '../types/enrollment';
import { useCallback } from 'react';

export interface UseEnrollmentsOptions {
  pollingInterval?: number;
  enabled?: boolean;
}

export function useEnrollments(options: UseEnrollmentsOptions = {}) {
  const fetchFn = useCallback(() => enrollmentsAPI.list(), []);

  return usePolling<NodeEnrollmentRequest[]>({
    fetchFn,
    interval: options.pollingInterval || 15000, // 15s for less frequent polling
    enabled: options.enabled !== false,
    onError: (error) => {
      console.error('Failed to fetch enrollments:', error);
    },
  });
}
