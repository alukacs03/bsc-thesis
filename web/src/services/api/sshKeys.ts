import { apiClient } from './client';
import type { NodeSSHKey, GenerateNodeSSHKeyResponse } from '../types/ssh';

export const sshKeysAPI = {
  listForNode: (nodeId: number) => apiClient.get<NodeSSHKey[]>(`/admin/nodes/${nodeId}/ssh-keys`),
  createForNode: (nodeId: number, input: { username: string; public_key: string; comment?: string }) =>
    apiClient.post<NodeSSHKey>(`/admin/nodes/${nodeId}/ssh-keys`, input),
  generateForNode: (nodeId: number, input: { username: string; comment?: string; bits?: number }) =>
    apiClient.post<GenerateNodeSSHKeyResponse>(`/admin/nodes/${nodeId}/ssh-keys/generate`, input),
  deleteForNode: (nodeId: number, keyId: number) =>
    apiClient.delete<void>(`/admin/nodes/${nodeId}/ssh-keys/${keyId}`),
};

