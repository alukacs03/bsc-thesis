export interface NodeSSHKey {
  id: number;
  node_id: number;
  username: string;
  public_key: string;
  comment?: string;
  created_at: string;
  updated_at: string;
}

export interface GenerateNodeSSHKeyResponse {
  id: number;
  node_id: number;
  username: string;
  public_key: string;
  private_key_pem: string;
}

