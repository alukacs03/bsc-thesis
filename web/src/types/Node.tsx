interface NodeRequest {
    id: number;
    hostname: string;
    ipAddress: string;
    location: string;
    provider: string;
    instanceType: string;
    requestedBy: string;
    requestedAt: string;
    purpose: string;
    specs: {
        cpu: string;
        memory: string;
        storage: string;
        bandwidth: string;
    };
    securityGroup: string;
    sshKey: string;
    estimatedCost: string;
    priority: string;
    reason?: string;
    status: string;
    approvedAt?: string;
    approvedBy?: string;
    rejectedAt?: string;
    rejectedBy?: string;
}

export default NodeRequest;