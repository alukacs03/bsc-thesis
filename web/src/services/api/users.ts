import { apiClient } from './client';
import type { UserRegistrationRequest, ModifyRegistrationRequest } from '../types/user';

export const usersAPI = {
  // List user registration requests
  listRegistrationRequests: () =>
    apiClient.get<UserRegistrationRequest[]>('/admin/userRegRequests'),

  // Modify registration request (approve/reject)
  modifyRegistrationRequest: (data: ModifyRegistrationRequest) =>
    apiClient.post<{ message: string }>('/admin/modifyRegistrationRequest', data),

  // Delete user
  deleteUser: (userId: number) =>
    apiClient.post<{ message: string }>('/admin/deleteUser', { user_id: userId }),
};
