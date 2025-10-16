export default interface UserRequest {
    id: number;
    username: string;
    email: string;
    fullName: string;
    requestedAt: string;
    reason?: string;
    approvedBy?: string;
    approvedAt?: string;
    status: string;
    rejectedBy?: string;
    rejectedAt?: string;
}
