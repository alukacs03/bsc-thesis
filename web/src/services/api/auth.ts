import { apiClient } from './client';
import type { User } from '../types/user';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export const authAPI = {
  login: (data: LoginRequest) => apiClient.post<{ message: string }>('/login', data),

  logout: () => apiClient.post<{ message: string }>('/logout'),

  register: (data: RegisterRequest) => apiClient.post<{ message: string }>('/register', data),

  getCurrentUser: () => apiClient.get<User>('/user'),
};
