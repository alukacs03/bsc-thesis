import { useNavigate, useParams } from "react-router-dom"
import { RefreshCw, HeartPulse, Server, Clock, Wifi, Key, Cog, HardDrive, Terminal } from "lucide-react"
import CardWithIcon from "../components/CardWithIcon";
import { getStatusColor } from "../utils/Helpers";
import CardContainer from "../components/CardContainer";
import DetailsNavBar from "../components/DetailsNavBar";
import React from "react";
import NodeManagementTabContent from "../components/NodeManagementTabContent";


const NodeManagementView = () => {
    const navigate = useNavigate();
    const [selectedTab, setSelectedTab] = React.useState<string>("Overview");

    // only one would come from the API with a nodeID parametered request
    const nodes = [
        {
            id: 1,
            name: 'gluon-master-01',
            hostname: 'gluon-master-01.numnet.local',
            ip: '10.0.1.10',
            status: 'online',
            role: 'Control Plane',
            lastSeen: '2 minutes ago',
            cpu: 45,
            memory: 62,
            disk: 38,
            uptime: '15 days, 6 hours',
            version: 'v1.28.2',
            pods: 12,
            location: 'us-east-1a',
            cpuType: '2 vCPUS (E5-2676 v3)',
            memorySize: '8 GB (DDR4)',
            diskSize: '100 GB (SSD)',
            os: 'Debian GNU/Linux 12 (bookworm)',
            logs : [
                "2024-10-01 12:00:00 [INFO] Node initialized successfully.",
                "2024-10-01 12:05:00 [WARN] High memory usage detected.",
                "2024-10-01 12:10:00 [INFO] New pod scheduled: pod-xyz.",
                "2024-10-01 12:15:00 [ERROR] Failed to pull image for pod-abc.",
                "2024-10-01 12:20:00 [INFO] Node metrics updated.",
                "2024-10-01 12:25:00 [INFO] Pod pod-xyz is now running.",
                "2024-10-01 12:30:00 [WARN] Disk space running low.",
                "2024-10-01 12:35:00 [INFO] Node metrics updated.",
                "2024-10-01 12:40:00 [ERROR] Network latency detected.",
                "2024-10-01 12:45:00 [INFO] Node operating normally."
            ],
            sshKeys: [
                {
                    user: 'test_jozsef',
                    id: 1,
                    name: 'admin-key',
                    type: 'RSA',
                    bits: 2048,
                    fingerprint: 'SHA256:abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567',
                    publicKey: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...',
                    created: '2024-09-15 10:00:00',
                    lastUsed: '2024-10-01 11:30:00',
                    status: 'active'
                },
                {
                    user: 'deploy_bela',
                    id: 2,
                    name: 'deploy-key',
                    type: 'ED25519',
                    bits: 256,
                    fingerprint: 'SHA256:xyz789uvw456rst123opq012lmn345ijk678hgf901edc234ba567',
                    publicKey: 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB...',
                    created: '2024-09-20 14:30:00',
                    lastUsed: '2024-09-30 09:15:00',
                    status: 'inactive'
                }
            ],
            services: [
                {
                    name: 'kubelet',
                    status: 'running',
                    enabled: true,
                    description: 'Kubernetes Node Agent'
                },
                {
                    name: 'gluon-agent',
                    status: 'running',
                    enabled: false,
                    description: 'Gluon Agent'
                },
                {
                    name: 'docker',
                    status: 'stopped',
                    enabled: false,
                    description: 'Docker Daemon'
                },
                {
                    name: 'ssh',
                    status: 'stopped',
                    enabled: true,
                    description: 'OpenSSH Server'
                },
                {
                    name: 'wireguard',
                    status: 'running',
                    enabled: true,
                    description: 'WireGuard VPN Service'
                }
            ],
            filesystemMounts: [
                {
                    mountpoint: '/',
                    device: '/dev/sda1',
                    type: 'ext4',
                    size: 100,
                    used: 38,
                },
                {
                    mountpoint: '/var/lib/docker',
                    device: '/dev/sdb1',
                    type: 'xfs',
                    size: 200,
                    used: 120,

                },
                {
                    mountpoint: '/mnt/data',
                    device: '/dev/sdc1',
                    type: 'ext4',
                    size: 500,
                    used: 450,
                }
            ]
        },
        {
            id: 2,
            name: 'gluon-worker-01',
            hostname: 'gluon-worker-01.numnet.local',
            ip: '10.0.1.11',
            status: 'online',
            role: 'Worker',
            lastSeen: '1 minute ago',
            cpu: 78,
            memory: 82,
            disk: 55,
            uptime: '15 days, 6 hours',
            version: 'v1.28.2',
            pods: 24,
            location: 'us-east-1b',
            cpuType: '2 vCPUS (E5-2676 v3)',
            memorySize: '8 GB (DDR4)',
            diskSize: '100 GB (SSD)',
            os: 'Debian GNU/Linux 12 (bookworm)',
        },
        {
            id: 3,
            name: 'gluon-worker-02',
            hostname: 'gluon-worker-02.numnet.local',
            ip: '10.0.1.12',
            status: 'online',
            role: 'Worker',
            lastSeen: '30 seconds ago',
            cpu: 32,
            memory: 48,
            disk: 41,
            uptime: '12 days, 14 hours',
            version: 'v1.28.2',
            pods: 18,
            location: 'us-east-1c',
            cpuType: '1 vCPU (E5-2676 v3)',
            memorySize: '8 GB (DDR4)',
            diskSize: '100 GB (SSD)',
            os: 'Debian GNU/Linux 12 (bookworm)',
        },
        {
            id: 4,
            name: 'gluon-worker-03',
            hostname: 'gluon-worker-03.numnet.local',
            ip: '10.0.1.13',
            status: 'offline',
            role: 'Worker',
            lastSeen: '2 hours ago',
            cpu: 0,
            memory: 0,
            disk: 68,
            uptime: '0 days, 0 hours',
            version: 'v1.28.2',
            pods: 0,
            location: 'us-east-1a',
            cpuType: '2 vCPUS (E5-2676 v3)',
            memorySize: '8 GB (DDR4)',
            diskSize: '100 GB (SSD)',
            os: 'Debian GNU/Linux 12 (bookworm)',
        },
        {
            id: 5,
            name: 'gluon-worker-04',
            hostname: 'gluon-worker-04.numnet.local',
            ip: '10.0.1.14',
            status: 'maintenance',
            role: 'Worker',
            lastSeen: '5 minutes ago',
            cpu: 15,
            memory: 25,
            disk: 33,
            uptime: '8 days, 2 hours',
            version: 'v1.28.1',
            pods: 3,
            location: 'us-east-1b',
            cpuType: '2 vCPUS (E5-2676 v3)',
            memorySize: '8 GB (DDR4)',
            diskSize: '100 GB (SSD)',
            os: 'Debian GNU/Linux 12 (bookworm)',
        }
    ];


  const { nodeId } = useParams<{ nodeId: string }>();
    const nodeData = nodes.find(node => node.id === Number(nodeId)) || nodes[0];
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate(-1)}
            className="p-2 text-slate-600 hover:text-slate-800 hover:bg-slate-100 rounded-lg transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 className="text-3xl text-slate-800">Node Management</h1>
            <p className="text-slate-600">{nodeData.name} ({nodeData.hostname})</p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-2 text-sm">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <span className="text-slate-600">Agent {nodeData.status.charAt(0).toUpperCase() + nodeData.status.slice(1)}</span>
          </div>
          <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2" onClick={() => alert('Refresh action triggered')}>
            <RefreshCw className="w-4 h-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <CardWithIcon
            title="Status"
            value={nodeData.status.charAt(0).toUpperCase() + nodeData.status.slice(1)}
            textColorClass="text-slate-600"
            valueColorClass="text-black-600"
            valueTextSize="text-xl"
            iconBGColorClass={getStatusColor(nodeData.status).replace('text', 'bg') + ' bg-opacity-20'}
            icon={<HeartPulse className="w-6 h-6 text-slate-800" />}
            hint={"Last seen: " + nodeData.lastSeen}
        />
        <CardWithIcon
            title="Uptime"
            value={nodeData.uptime}
            textColorClass="text-slate-600"
            valueColorClass="text-black-600"
            valueTextSize="text-xl"
            iconBGColorClass="bg-blue-100"
            icon={<Clock className="w-6 h-6 text-slate-800"/>}
            hint={"Agent version: " + nodeData.version}
        />
        <CardWithIcon
            title="Resources"
            value={nodeData.cpuType}
            textColorClass="text-slate-600"
            valueColorClass="text-black-600"
            valueTextSize="text-xl"
            icon={<Server className="w-6 h-6 text-slate-800"/>}
            iconBGColorClass="bg-blue-100"
            hint={`${nodeData.memorySize} RAM / ${nodeData.diskSize} Disk`}
        />
        <CardWithIcon
            title="Network"
            value={nodeData.ip}
            textColorClass="text-slate-600"
            valueColorClass="text-black-600"
            valueTextSize="text-xl"
            iconBGColorClass="bg-blue-100"
            icon={<Wifi className="w-6 h-6 text-slate-800"/>}
            hint={nodeData.role + " Node"}
        />
        <CardContainer
            title="Node Controls"
            colSpan="col-span-1 md:col-span-4"
            icon={<Server className="w-5 h-5"/>}
            noPadding={true}
        >
            <DetailsNavBar 
                tabs={['Overview', 'SSH Keys', 'Services', 'Filesystem', 'Logs & Terminal']}
                icons={[<Server className="w-4 h-4"/>, <Key className="w-4 h-4"/>, <Cog className="w-4 h-4"/>, <HardDrive className="w-4 h-4"/>, <Terminal className="w-4 h-4"/>]}
                selectedTab={selectedTab}
                setSelectedTab={setSelectedTab}
             />
             <NodeManagementTabContent selectedTab={selectedTab} node={nodeData} />
        </CardContainer>
      </div>
    </div>
  )
}

export default NodeManagementView
