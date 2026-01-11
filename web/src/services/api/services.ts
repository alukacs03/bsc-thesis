import { apiClient } from './client';

export const servicesAPI = {
  restart: (nodeId: number, name: string) =>
    apiClient.post<{ command_id: number }>(`/admin/nodes/${nodeId}/services/restart`, { name }),
};

