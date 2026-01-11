import { apiClient } from './client';
import type { Node } from '../types/node';

export interface NodeLogsResponse {
  window: string;
  logs: string[];
}

export const nodesAPI = {
  // List all nodes
  list: () => apiClient.get<Node[]>('/admin/nodes'),

  // Get single node
  get: (id: number) => apiClient.get<Node>(`/admin/nodes/${id}`),

  // Get node logs/events
  logs: (id: number, limit = 200) =>
    apiClient.get<NodeLogsResponse>(`/admin/nodes/${id}/logs?limit=${limit}`),

  // Delete node
  delete: (id: number) => apiClient.delete<void>(`/admin/nodes/${id}`),
};
