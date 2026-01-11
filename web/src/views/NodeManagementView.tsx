import { useNavigate, useParams } from "react-router-dom"
import { HeartPulse, Server, Clock, Wifi, Key, Cog, Terminal, Trash2 } from "lucide-react"
import CardWithIcon from "../components/CardWithIcon";
import { getStatusColor } from "../utils/Helpers";
import CardContainer from "../components/CardContainer";
import DetailsNavBar from "../components/DetailsNavBar";
import React from "react";
import NodeManagementTabContent from "../components/NodeManagementTabContent";
import { useNode } from "@/services/hooks/useNodes";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import { ErrorMessage } from "@/components/ErrorMessage";
import { nodesAPI } from "@/services/api/nodes";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";
import { formatBytes, formatPercent, formatUptimeSeconds } from "@/utils/format";


const NodeManagementView = () => {
    const navigate = useNavigate();
    const [selectedTab, setSelectedTab] = React.useState<string>("Overview");
    const { nodeId } = useParams<{ nodeId: string }>();

    const { data: node, loading, error, refetch } = useNode(Number(nodeId), { pollingInterval: 30000 });

    const handleDelete = async () => {
      if (!node) return;

      if (!confirm(`Are you sure you want to delete node "${node.hostname}"? This action cannot be undone.`)) {
        return;
      }

      try {
        await nodesAPI.delete(node.id);
        toast.success('Node deleted successfully');
        navigate('/nodes');
      } catch (error) {
        const message = handleAPIError(error, 'delete node');
        toast.error(message);
      }
    };

    if (loading && !node) {
      return <LoadingSpinner />;
    }

    if (error) {
      return <ErrorMessage message={error.message} onRetry={refetch} />;
    }

    if (!node) {
      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <p className="text-slate-600 mb-4">Node not found</p>
            <button
              onClick={() => navigate('/nodes')}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Back to Nodes
            </button>
          </div>
        </div>
      );
    }

    function formatLastSeen(timestamp?: string): string {
      if (!timestamp) return 'Never';
      const diff = Date.now() - new Date(timestamp).getTime();
      const minutes = Math.floor(diff / 60000);
      if (minutes < 1) return 'Just now';
      if (minutes < 60) return `${minutes}m ago`;
      const hours = Math.floor(minutes / 60);
      if (hours < 24) return `${hours}h ago`;
      return `${Math.floor(hours / 24)}d ago`;
    }

    // Map backend node to display format
    const nodeData = {
      id: node.id,
      name: node.hostname,
      hostname: node.hostname,
      ip: node.public_ip,
      status: node.status === 'active' ? 'online' : node.status,
      role: node.role === 'hub' ? 'Hub' : 'Worker',
      lastSeen: formatLastSeen(node.last_seen_at),
      cpu: node.cpu_usage ?? undefined,
      memory: node.memory_usage ?? undefined,
      disk: node.disk_usage ?? undefined,
      diskUsedBytes: node.disk_used_bytes ?? undefined,
      diskTotalBytes: node.disk_total_bytes ?? undefined,
      uptime: formatUptimeSeconds(node.uptime_seconds),
      version: node.agent_version || 'Unknown',
      pods: 0, // Not available yet
      location: node.provider,
      cpuType: 'Unknown', // Not available yet
      memorySize: 'Unknown', // Not available yet
      diskSize: 'Unknown', // Not available yet
      os: node.os,
      systemUsers: (node.system_users ?? undefined) || undefined,
      logs: [], // Not available yet
      services: node.system_services ?? undefined,
      filesystemMounts: [] // Not available yet
    };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate('/nodes')}
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
            <div className={`w-2 h-2 rounded-full ${nodeData.status === 'online' ? 'bg-green-500' : nodeData.status === 'offline' ? 'bg-red-500' : 'bg-yellow-500'}`}></div>
            <span className="text-slate-600">Agent {nodeData.status.charAt(0).toUpperCase() + nodeData.status.slice(1)}</span>
          </div>
          <button
            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
            onClick={handleDelete}
          >
            <Trash2 className="w-4 h-4" />
            <span>Delete</span>
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
            value={`CPU ${formatPercent(node.cpu_usage)}`}
            textColorClass="text-slate-600"
            valueColorClass="text-black-600"
            valueTextSize="text-xl"
            icon={<Server className="w-6 h-6 text-slate-800"/>}
            iconBGColorClass="bg-blue-100"
            hint={`RAM ${formatPercent(node.memory_usage)} / Disk ${formatPercent(node.disk_usage)} (${formatBytes(node.disk_used_bytes)} / ${formatBytes(node.disk_total_bytes)})`}
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
                tabs={['Overview', 'Networking', 'SSH Keys', 'Services', 'Logs']}
                icons={[<Server className="w-4 h-4"/>, <Wifi className="w-4 h-4"/>, <Key className="w-4 h-4"/>, <Cog className="w-4 h-4"/>, <Terminal className="w-4 h-4"/>]}
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
