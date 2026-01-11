import { apiClient } from './client';
import type {
  IPPool,
  IPAllocation,
  CreateIPPoolRequest,
  AllocateIPRequest,
  NextAvailableIPResponse,
} from '../types/ipam';

export const ipamAPI = {
  // IP Pools
  pools: {
    list: () => apiClient.get<IPPool[]>('/admin/ipam/pools'),
    create: (data: CreateIPPoolRequest) => apiClient.post<IPPool>('/admin/ipam/pools', data),
    delete: (id: number) => apiClient.delete<void>(`/admin/ipam/pools/${id}`),
    getNextAvailable: (id: number) =>
      apiClient.get<NextAvailableIPResponse>(`/admin/ipam/pools/${id}/next`),
    allocateNext: (id: number, data: Omit<AllocateIPRequest, 'value' | 'pool_id'>) =>
      apiClient.post<IPAllocation>(`/admin/ipam/pools/${id}/allocate-next`, data),
  },

  // IP Allocations
  allocations: {
    list: () => apiClient.get<IPAllocation[]>('/admin/ipam/allocations'),
    get: (id: number) => apiClient.get<IPAllocation>(`/admin/ipam/allocations/${id}`),
    create: (data: AllocateIPRequest) =>
      apiClient.post<IPAllocation>('/admin/ipam/allocations', data),
    delete: (id: number) => apiClient.delete<void>(`/admin/ipam/allocations/${id}`),
  },
};
