export default interface UserRequest {
    id: number;
    email: string;
    full_name: string;
    requested_at: string;
    rejection_reason?: string;
    approved_by?: User;
    approved_at?: string;
    status: string;
    rejected_by?: User;
    rejected_at?: string;
}

export interface User {
    id: number;
    name: string;
    email: string;
    passwordHash ?: string;
}