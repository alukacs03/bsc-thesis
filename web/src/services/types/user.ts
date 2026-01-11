export interface User {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  email: string;
}

export interface UserRegistrationRequest {
  id: number;
  requested_at: string;
  status: 'pending' | 'approved' | 'rejected';
  email: string;
  full_name: string;
  approved_at?: string;
  approved_by_id?: number;
  approved_by?: User;
  rejection_reason?: string;
  rejected_at?: string;
  rejected_by_id?: number;
  rejected_by?: User;
  converted_user_id?: number;
}

export interface ModifyRegistrationRequest {
  request_id: string;
  status: 'approved' | 'rejected';
  rejection_reason?: string;
}
