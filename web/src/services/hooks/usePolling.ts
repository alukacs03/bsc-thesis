import { useEffect, useRef, useState, useCallback } from 'react';

export interface UsePollingOptions<T> {
  fetchFn: () => Promise<T>;
  interval?: number; // milliseconds, default 10000 (10s)
  enabled?: boolean; // default true
  onError?: (error: Error) => void;
}

export interface UsePollingResult<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function usePolling<T>({
  fetchFn,
  interval = 10000,
  enabled = true,
  onError,
}: UsePollingOptions<T>): UsePollingResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isVisible, setIsVisible] = useState(true);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isMountedRef = useRef(true);
  const onErrorRef = useRef<UsePollingOptions<T>['onError']>(onError);

  useEffect(() => {
    onErrorRef.current = onError;
  }, [onError]);

  const fetchData = useCallback(async () => {
    try {
      const result = await fetchFn();
      if (isMountedRef.current) {
        setData(result);
        setError(null);
        setLoading(false);
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error');
      if (isMountedRef.current) {
        setError(error);
        setLoading(false);
      }
      onErrorRef.current?.(error);
    }
  }, [fetchFn]);

  const scheduleNextFetch = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    if (enabled && isVisible) {
      timeoutRef.current = setTimeout(() => {
        fetchData().then(scheduleNextFetch);
      }, interval);
    }
  }, [enabled, isVisible, interval, fetchData]);

  // Handle visibility changes
  useEffect(() => {
    const handleVisibilityChange = () => {
      setIsVisible(document.visibilityState === 'visible');
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  // Initial fetch and polling setup
  useEffect(() => {
    if (enabled) {
      fetchData().then(scheduleNextFetch);
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
  }, [enabled, fetchData, scheduleNextFetch]);

  // Track mount/unmount
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  const refetch = useCallback(async () => {
    setLoading(true);
    await fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch };
}
