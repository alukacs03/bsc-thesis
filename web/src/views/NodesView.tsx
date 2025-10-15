import CardWithIcon from "../components/CardWithIcon"
import { AlertTriangle, Server, XCircle, CheckCircle2, ServerCog } from "lucide-react"
import CardContainer from "../components/CardContainer"
import NodeListItem from "../components/NodeListItem"
import { useState } from "react"
import DetailsNavBar from "../components/DetailsNavBar"
import NodesTabContent from "../components/NodesTabContent"

const NodesView = () => {
  const [selectedNode, setSelectedNode] = useState<number | null>(null);
  const [selectedTab, setSelectedTab] = useState<string>("Overview");

    const nodes = [
        {
            id: 1,
            name: 'gluon-master-01',
            ip: '10.0.1.10',
            status: 'online',
            role: 'Control Plane',
            lastHeartbeat: '2 minutes ago',
            cpu: 45,
            memory: 62,
            disk: 38,
            uptime: '15 days, 6 hours',
            version: 'v1.28.2',
            pods: 12,
            location: 'us-east-1a',
            logs : [
                "2024-10-01 12:00:00 [INFO] Node initialized successfully.",
                "2024-10-01 12:05:00 [WARN] High memory usage detected.",
                "2024-10-01 12:10:00 [INFO] New pod scheduled: pod-xyz.",
                "2024-10-01 12:15:00 [ERROR] Failed to pull image for pod-abc.",
                "2024-10-01 12:20:00 [INFO] Node heartbeat sent.",
                "2024-10-01 12:25:00 [INFO] Pod pod-xyz is now running.",
                "2024-10-01 12:30:00 [WARN] Disk space running low.",
                "2024-10-01 12:35:00 [INFO] Node metrics updated.",
                "2024-10-01 12:40:00 [ERROR] Network latency detected.",
                "2024-10-01 12:45:00 [INFO] Node operating normally."
            ]
        },
        {
            id: 2,
            name: 'gluon-worker-01',
            ip: '10.0.1.11',
            status: 'online',
            role: 'Worker',
            lastHeartbeat: '1 minute ago',
            cpu: 78,
            memory: 82,
            disk: 55,
            uptime: '15 days, 6 hours',
            version: 'v1.28.2',
            pods: 24,
            location: 'us-east-1b'
        },
        {
            id: 3,
            name: 'gluon-worker-02',
            ip: '10.0.1.12',
            status: 'online',
            role: 'Worker',
            lastHeartbeat: '30 seconds ago',
            cpu: 32,
            memory: 48,
            disk: 41,
            uptime: '12 days, 14 hours',
            version: 'v1.28.2',
            pods: 18,
            location: 'us-east-1c'
        },
        {
            id: 4,
            name: 'gluon-worker-03',
            ip: '10.0.1.13',
            status: 'offline',
            role: 'Worker',
            lastHeartbeat: '2 hours ago',
            cpu: 0,
            memory: 0,
            disk: 68,
            uptime: '0 days, 0 hours',
            version: 'v1.28.2',
            pods: 0,
            location: 'us-east-1a'
        },
        {
            id: 5,
            name: 'gluon-worker-04',
            ip: '10.0.1.14',
            status: 'maintenance',
            role: 'Worker',
            lastHeartbeat: '5 minutes ago',
            cpu: 15,
            memory: 25,
            disk: 33,
            uptime: '8 days, 2 hours',
            version: 'v1.28.1',
            pods: 3,
            location: 'us-east-1b'
        }
    ];


  return (
    <>
        <div className="space-y-6">
            <div className="grid grid-cols1 md:grid-cols-4 gap-6">
                <CardWithIcon
                    title="Total Nodes"
                    value="5"
                    textColorClass="text-slate-600"
                    valueColorClass="text-blue-600"
                    iconBGColorClass="bg-blue-100"
                    icon={<Server className="w-6 h-6 text-slate-800"/>}
                />
                <CardWithIcon
                    title="Nodes Online"
                    value="4"
                    textColorClass="text-slate-600"
                    valueColorClass="text-green-600"
                    iconBGColorClass="bg-green-100"
                    icon={<CheckCircle2 className="w-6 h-6 text-green-600"/>}
                />
                <CardWithIcon
                    title="Nodes Offline"
                    value="1"
                    textColorClass="text-slate-600"
                    valueColorClass="text-red-600"
                    iconBGColorClass="bg-red-100"
                    icon={<XCircle className="w-6 h-6 text-red-600"/>}
                />
                <CardWithIcon
                    title="Maintenance"
                    value="0"
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
                            onClick={() => alert('Add Node clicked (would navigate to add node form)')}
                        >
                            Add Node
                        </button>
                    }
                    noPadding={true}
                 >
                    {nodes.map(node => (
                        <NodeListItem
                            id={node.id}
                            name={node.name}
                            status={node.status}
                            ip={node.ip}
                            role={node.role}
                            cpu={node.cpu}
                            memory={node.memory}
                            pods={node.pods}
                            lastHeartbeat={node.lastHeartbeat}
                            selectedNode={selectedNode}
                            setSelectedNode={(id: number) => {setSelectedNode(id)}}
                        />
                    ))}
                </CardContainer>
                <CardContainer
                    title={(selectedNode !== null) ? `Node Details - ID: ${selectedNode}` : "Select Node to See Details"}
                    icon={<ServerCog className="w-5 h-5"/>}
                    noPadding={true}
                 >
                    <DetailsNavBar
                        tabs={["Overview", "Logs"]}
                        setSelectedTab={setSelectedTab}
                        selectedTab={selectedTab}
                    />
                    {selectedNode !== null && (
                        <NodesTabContent node={nodes[selectedNode - 1]} selectedTab={selectedTab} />
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
