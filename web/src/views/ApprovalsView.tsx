import { Server, Users } from "lucide-react"
import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import ApprovalsTabContent from "../components/ApprovalsTabContent"
import React from "react"
import DetailsNavBar from "../components/DetailsNavBar"
import { useState } from "react"
import type NodeRequest from "../types/Node"


  const nodeApprovals: NodeRequest[] = [
    {
      id: 1,
      hostname: 'vps-prod-web-04.gluon.io',
      ipAddress: '203.0.113.45',
      location: 'US-East (Virginia)',
      provider: 'AWS EC2',
      instanceType: 't3.large',
      requestedAt: '2024-01-15 14:30:00',
      priority: 'urgent',
      requestedBy: 'ops-team@company.com',
      purpose: 'Production web server replacement for load balancing',
      specs: {
        cpu: '2 vCPU',
        memory: '8 GB RAM',
        storage: '100 GB SSD',
        bandwidth: '5 TB/month'
      },
      securityGroup: 'web-servers-prod',
      sshKey: 'gluon-prod-key-2024',
      estimatedCost: '$45.60/month',
      status: 'pending'
    },
    {
      id: 2,
      hostname: 'vps-dev-k8s-worker-07.gluon.io', 
      ipAddress: '198.51.100.89',
      location: 'EU-West (Ireland)',
      provider: 'DigitalOcean',
      instanceType: 's-4vcpu-8gb',
      requestedAt: '2024-01-15 13:45:00',
      priority: 'high',
      requestedBy: 'dev-team@company.com',
      purpose: 'Additional Kubernetes worker node for development cluster',
      specs: {
        cpu: '4 vCPU',
        memory: '8 GB RAM', 
        storage: '160 GB SSD',
        bandwidth: '5 TB/month'
      },
      securityGroup: 'k8s-workers-dev',
      sshKey: 'gluon-dev-key-2024',
      estimatedCost: '$48.00/month',
      status: 'pending'
    },
    {
      id: 3,
      hostname: 'vps-staging-db-02.gluon.io',
      ipAddress: '192.0.2.156',
      location: 'US-West (Oregon)',
      provider: 'Linode',
      instanceType: 'Dedicated 8GB',
      requestedAt: '2024-01-15 12:15:00',
      priority: 'medium',
      requestedBy: 'db-admin@company.com',
      purpose: 'Staging database server for testing new schema migrations',
      specs: {
        cpu: '4 vCPU',
        memory: '8 GB RAM',
        storage: '320 GB SSD',
        bandwidth: '8 TB/month'
      },
      securityGroup: 'database-staging',
      sshKey: 'gluon-staging-key-2024',
      estimatedCost: '$60.00/month',
      status: 'pending'
    },
    {
        id: 4,
        hostname: 'vps-prod-cache-01.gluon.io',
        ipAddress: '203.0.113.45',
        location: 'US-East (Virginia)',
        provider: 'AWS EC2',
        instanceType: 't3.medium',
        requestedAt: '2024-01-14 10:20:00',
        priority: 'high',
        requestedBy: 'ops-team@company.com',
        purpose: 'Production caching server for web applications',
        specs: {
            cpu: '2 vCPU',
            memory: '4 GB RAM',
            storage: '50 GB SSD',
            bandwidth: '3 TB/month'
        },
        securityGroup: 'cache-servers-prod',
        sshKey: 'gluon-prod-key-2024',
        estimatedCost: '$30.00/month',
        status: 'approved',
        approvedAt: '2024-01-14 11:00:00',
        approvedBy: 'admin@example.com'
    },
    {
        id: 5,
        hostname: 'vps-dev-web-01.gluon.io',
        ipAddress: '203.0.113.45',
        location: 'US-East (Virginia)',
        provider: 'AWS EC2',
        instanceType: 't3.medium',
        requestedAt: '2024-01-13 09:30:00',
        priority: 'medium',
        requestedBy: 'dev-team@company.com',
        purpose: 'Development web server for testing new features',
        specs: {
            cpu: '2 vCPU',
            memory: '4 GB RAM',
            storage: '50 GB SSD',
            bandwidth: '3 TB/month'
        },
        securityGroup: 'web-servers-dev',
        sshKey: 'gluon-dev-key-2024',
        estimatedCost: '$30.00/month',
        status: 'approved',
        approvedAt: '2024-01-13 10:00:00',
        approvedBy: 'admin@example.com'
    },
    {
        id: 6,
        hostname: 'vps-test-db-01.gluon.io',
        ipAddress: '192.0.2.156',
        location: 'US-West (Oregon)',
        provider: 'Linode',
        instanceType: 'Dedicated 4GB',
        requestedAt: '2024-01-12 08:00:00',
        priority: 'low',
        requestedBy: 'db-admin@company.com',
        purpose: 'Test database server for running integration tests',
        specs: {
            cpu: '2 vCPU',
            memory: '4 GB RAM',
            storage: '160 GB SSD',
            bandwidth: '4 TB/month'
        },
        securityGroup: 'database-test',
        sshKey: 'gluon-test-key-2024',
        estimatedCost: '$40.00/month',
        status: 'rejected',
        rejectedAt: '2024-01-12 09:00:00',
        rejectedBy: 'admin@example.com',
        reason: 'Insufficient justification for additional database server in test environment.'
    }
  ];

  // User registration approvals data
  const userApprovals = [
    {
      id: 1,
      username: 'john.smith',
      email: 'john.smith@example.com',
      fullName: 'John Smith',
      requestedAt: '2024-01-15 14:30:00',
      status: 'pending'
    },
    {
      id: 2,
      username: 'alice.johnson',
      email: 'alice.johnson@example.com',
      fullName: 'Alice Johnson',
      requestedAt: '2024-01-15 13:45:00',
      status: 'pending'
    },
    {
      id: 3,
      username: 'bob.wilson',
      email: 'bob.wilson@example.com',
      fullName: 'Bob Wilson',
      requestedAt: '2024-01-15 12:15:00',
      status: 'pending'
    },
        {
      id: 4,
      username: 'sarah.admin',
      email: 'sarah.admin@example.com',
      fullName: 'Sarah Wilson',
      requestedAt: '2024-01-15 10:20:00',
      approvedAt: '2024-01-15 10:45:00',
      approvedBy: 'admin@example.com',
      status: 'approved',
      approvedBy: 'admin@example.com',
      approvedAt: '2024-01-15 10:45:00',
    },
    {
        id: 5,
        username: 'mike.chen',
        email: 'mike.chen@example.com',
        fullName: 'Mike Chen',
        requestedAt: '2024-01-15 09:30:00',
        approvedAt: '2024-01-15 09:35:00',
        approvedBy: 'admin@example.com',
        status: 'approved',
        approvedBy: 'admin@example.com',
        approvedAt: '2024-01-15 10:45:00',
    },
    {
        id: 6,
        username: 'spam.user',
        email: 'spam@suspicious.com',
        fullName: 'Spam User',
        requestedAt: '2024-01-15 08:00:00',
        rejectedAt: '2024-01-15 08:30:00',
        rejectedBy: 'admin@example.com',
        reason: 'Suspicious registration attempt with invalid email domain.',
        status: 'rejected',
    },
    {
        id: 7,
        username: 'test.user',
        email: 'test.user@example.com',
        fullName: 'Test User',
        requestedAt: '2024-01-15 08:00:00',
        status: 'pending'
    },
    {
        id: 8,
        username: 'rejected.user',
        email: 'rejected.user@example.com',
        fullName: 'Rejected User',
        requestedAt: '2024-01-15 07:30:00',
        rejectedAt: '2024-01-15 08:00:00',
        rejectedBy: 'admin@example.com',
        reason: 'User did not provide required identification documents.',
        status: 'rejected'
    }
  ];

const ApprovalsView = () => {
  const [selectedCategory, setSelectedCategory] = React.useState<'vps' | 'users'>('vps');
  const [selectedTab, setSelectedTab] = useState<'pending' | 'approved' | 'rejected'>('pending');


  return (
    <>
        <div className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm">
                <div className="flex border-b border-slate-200">
                <button
                    onClick={() => setSelectedCategory('vps')}
                    className={`flex-1 flex items-center justify-center space-x-3 px-6 py-4 transition-colors ${
                    selectedCategory === 'vps'
                        ? 'bg-blue-50 border-b-2 border-blue-500 text-blue-600'
                        : 'text-slate-600 hover:text-slate-700 hover:bg-slate-50'
                    }`}
                >
                    <Server className="w-5 h-5" />
                    <span>VPS Node Approvals</span>
                    <span className="bg-red-500 text-white text-xs px-2 py-1 rounded-full">{nodeApprovals.length}</span>
                </button>
                <button
                    onClick={() => setSelectedCategory('users')}
                    className={`flex-1 flex items-center justify-center space-x-3 px-6 py-4 transition-colors ${
                    selectedCategory === 'users'
                        ? 'bg-blue-50 border-b-2 border-blue-500 text-blue-600'
                        : 'text-slate-600 hover:text-slate-700 hover:bg-slate-50'
                    }`}
                >
                    <Users className="w-5 h-5" />
                    <span>User Registration Approvals</span>
                    <span className="bg-red-500 text-white text-xs px-2 py-1 rounded-full">{userApprovals.length}</span>
                </button>
                </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-1 gap-6">
            {(selectedCategory === 'vps' && (
                <CardWithIcon
                    icon={<Server className="w-6 h-6 text-blue-600" />}
                    title="Total VPS Node Requests"
                    value={nodeApprovals.length}
                    noHover={true}
                />
            )) || (selectedCategory === 'users' && (
                <CardWithIcon
                    icon={<Users className="w-6 h-6 text-blue-600" />}
                    title="Total User Registration Requests"
                    value={userApprovals.length}
                    noHover={true}
                />
            ))}
            </div>
            <CardContainer
                title={selectedCategory === 'vps' ? 'VPS Node Approvals' : 'User Registration Approvals'}
                noPadding={true}
            >
                <DetailsNavBar
                    tabs={['pending', 'approved', 'rejected']}
                    setSelectedTab={(tab) => setSelectedTab(tab as 'pending' | 'approved' | 'rejected')}
                    selectedTab={selectedTab}
                />
                <div className="p-6">
                    <ApprovalsTabContent
                        selectedCategory={selectedCategory}
                        selectedTab={selectedTab}
                        nodeApprovals={nodeApprovals}
                        userApprovals={userApprovals}
                    />
                </div>

            </CardContainer>

        </div>
    </>
  )
}

export default ApprovalsView
