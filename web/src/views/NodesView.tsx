import CardWithIcon from "../components/CardWithIcon"
import { AlertTriangle, Server, XCircle, CheckCircle2, ServerCog } from "lucide-react"
import CardContainer from "../components/CardContainer"
import NodeListItem from "../components/NodeListItem"
import { useState } from "react"
import DetailsNavBar from "../components/DetailsNavBar"
import NodesTabContent from "../components/NodesTabContent"
import { useNodes } from "@/services/hooks/useNodes"
import { LoadingSpinner } from "@/components/LoadingSpinner"
import { ErrorMessage } from "@/components/ErrorMessage"
import { nodesAPI } from "@/services/api/nodes"
import { toast } from "sonner"
import { handleAPIError } from "@/utils/errorHandler"
import { useNavigate } from "react-router-dom"
import { formatUptimeSeconds } from "@/utils/format"

const NodesView = () => {
  const [selectedNode, setSelectedNode] = useState<number | null>(null);
  const [selectedTab, setSelectedTab] = useState<string>("Overview");
  const { data: nodes, loading, error, refetch } = useNodes();
  const navigate = useNavigate();

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this node? This action cannot be undone.')) {
      return;
    }

    try {
      await nodesAPI.delete(id);
      toast.success('Node deleted successfully');
      if (selectedNode === id) {
        setSelectedNode(null);
      }
      await refetch();
    } catch (error) {
      const message = handleAPIError(error, 'delete node');
      toast.error(message);
    }
  };

  if (loading && !nodes) {
    return <LoadingSpinner />;
  }

  if (error) {
    return <ErrorMessage message={error.message} onRetry={refetch} />;
  }

  const nodesList = nodes || [];
  const totalNodes = nodesList.length;
  const onlineNodes = nodesList.filter(n => n.status === 'active').length;
  const offlineNodes = nodesList.filter(n => n.status === 'offline').length;
  const maintenanceNodes = nodesList.filter(n => n.status === 'maintenance').length;

  // Map backend Node to display format expected by components
  const mockNodes = nodesList.map(node => ({
    id: node.id,
    name: node.hostname,
    ip: node.public_ip,
    status: node.status === 'active' ? 'online' : node.status,
    role: node.role === 'hub' ? 'Hub' : 'Worker',
    lastSeen: node.last_seen_at ? formatLastSeen(node.last_seen_at) : 'Never',
    cpu: node.cpu_usage ?? Number.NaN,
    memory: node.memory_usage ?? Number.NaN,
    disk: node.disk_usage ?? Number.NaN,
    diskUsedBytes: node.disk_used_bytes ?? undefined,
    diskTotalBytes: node.disk_total_bytes ?? undefined,
    uptime: formatUptimeSeconds(node.uptime_seconds),
    version: node.agent_version || 'Unknown',
    pods: 0, // Not available yet
    location: node.provider,
    logs: [] // Not available yet
  }));

  function formatLastSeen(timestamp: string): string {
    const diff = Date.now() - new Date(timestamp).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  }

  return (
    <>
        <div className="space-y-6">
            <div className="grid grid-cols1 md:grid-cols-4 gap-6">
                <CardWithIcon
                    title="Total Nodes"
                    value={totalNodes.toString()}
                    textColorClass="text-slate-600"
                    valueColorClass="text-blue-600"
                    iconBGColorClass="bg-blue-100"
                    icon={<Server className="w-6 h-6 text-slate-800"/>}
                />
                <CardWithIcon
                    title="Nodes Online"
                    value={onlineNodes.toString()}
                    textColorClass="text-slate-600"
                    valueColorClass="text-green-600"
                    iconBGColorClass="bg-green-100"
                    icon={<CheckCircle2 className="w-6 h-6 text-green-600"/>}
                />
                <CardWithIcon
                    title="Nodes Offline"
                    value={offlineNodes.toString()}
                    textColorClass="text-slate-600"
                    valueColorClass="text-red-600"
                    iconBGColorClass="bg-red-100"
                    icon={<XCircle className="w-6 h-6 text-red-600"/>}
                />
                <CardWithIcon
                    title="Maintenance"
                    value={maintenanceNodes.toString()}
                    textColorClass="text-slate-600"
                    valueColorClass="text-yellow-600"
                    iconBGColorClass="bg-yellow-100"
                    icon={<AlertTriangle className="w-6 h-6 text-yellow-600"/>}
                />
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <CardContainer
                    title="Node List"
                    icon={<Server className="w-5 h-5"/>}
                    button={
                        <button
                            className="bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition"
                            onClick={() => navigate('/approvals')}
                        >
                            Approve Nodes
                        </button>
                    }
                    noPadding={true}
                 >
                    {mockNodes.length > 0 ? (
                      mockNodes.map(node => (
                          <NodeListItem
                              key={node.id}
                              id={node.id}
                              name={node.name}
                              status={node.status}
                              ip={node.ip}
                              role={node.role}
                              cpu={node.cpu}
                              memory={node.memory}
                              pods={node.pods}
                              lastSeen={node.lastSeen}
                              selectedNode={selectedNode}
                              setSelectedNode={(id: number) => {setSelectedNode(id)}}
                          />
                      ))
                    ) : (
                      <div className="p-6 text-slate-600">
                        No nodes available. Approve enrollment requests to add nodes.
                      </div>
                    )}
                </CardContainer>
                <CardContainer
                    title={(selectedNode !== null) ? `Node Details - ${mockNodes.find(n => n.id === selectedNode)?.name}` : "Select Node to See Details"}
                    icon={<ServerCog className="w-5 h-5"/>}
                    noPadding={true}
                 >
                    <DetailsNavBar
                        tabs={["Overview", "Networking", "Logs"]}
                        setSelectedTab={setSelectedTab}
                        selectedTab={selectedTab}
                    />
                    {selectedNode !== null && mockNodes.find(n => n.id === selectedNode) && (
                        <NodesTabContent
                          node={mockNodes.find(n => n.id === selectedNode)!}
                          selectedTab={selectedTab}
                          onDelete={() => handleDelete(selectedNode)}
                        />
                    )}
                    {selectedNode === null && (
                        <div className="p-6 text-slate-600">
                            Please select a node from the list to view its details.
                        </div>
                    )}
                </CardContainer>
            </div>
        </div>
    </>
  )
}

export default NodesView
