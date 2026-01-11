import { apiClient } from './client';
import type {
  NodeEnrollmentRequest,
  APIKeyResponse,
  EnrollmentApprovalRequest,
  EnrollmentRejectionRequest,
} from '../types/enrollment';

export const enrollmentsAPI = {
  // List all enrollment requests
  list: () => apiClient.get<NodeEnrollmentRequest[]>('/admin/enrollments'),

  // Approve enrollment
  approve: (id: number, data?: EnrollmentApprovalRequest) =>
    apiClient.post<{ message: string }>(`/admin/enrollments/${id}/approve`, data),

  // Reject enrollment
  reject: (id: number, data: EnrollmentRejectionRequest) =>
    apiClient.post<{ message: string }>(`/admin/enrollments/${id}/reject`, data),

  // Generate API key for approved node
  generateAPIKey: (nodeId: number) =>
    apiClient.post<APIKeyResponse>('/admin/generateAPIKey', { node_id: nodeId }),

  // Revoke API key
  revokeAPIKey: (nodeId: number) =>
    apiClient.post<{ message: string }>('/admin/revokeApiKey', { node_id: nodeId }),
};
