export interface NodeEnrollmentRequest {
  id: number;
  requested_at: string;
  status: 'pending' | 'approved' | 'accepted' | 'rejected';
  hostname: string;
  public_ip: string;
  provider: string;
  os: string;
  desired_role: 'hub' | 'worker';
  approved_at?: string;
  approved_by_id?: number;
  approved_by?: {
    id: number;
    name: string;
    email: string;
  };
  rejection_reason?: string;
  rejected_at?: string;
  rejected_by_id?: number;
  rejected_by?: {
    id: number;
    name: string;
    email: string;
  };
  converted_node_id?: number;
}

export interface APIKeyResponse {
  api_key: string;
  node_id: number;
  message: string;
}

export interface EnrollmentApprovalRequest {
  node_id?: number;
}

export interface EnrollmentRejectionRequest {
  reason: string;
}
