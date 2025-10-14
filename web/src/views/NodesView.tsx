import CardWithIcon from "../components/CardWithIcon"
import { AlertTriangle, Server, XCircle, CheckCircle2, ServerCog } from "lucide-react"
import CardContainer from "../components/CardContainer"
import NodeListItem from "../components/NodeListItem"
import { useState } from "react"

const NodesView = () => {
  const [selectedNode, setSelectedNode] = useState<number | null>(null);
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
                    <NodeListItem
                        id={1}
                        name="gluon-master-01"
                        status="Ready"
                        ip="192.168.1.1"
                        role="master"
                        cpu={45}
                        memory={70}
                        pods={12}
                        lastHeartbeat="2024-10-01 12:34:56"
                        selectedNode={selectedNode}
                        setSelectedNode={(id: number) => {setSelectedNode(id)}}
                    />
                    <NodeListItem
                        id={2}
                        name="gluon-worker-01"
                        status="Ready"
                        ip="192.168.1.2"
                        role="worker"
                        cpu={30}
                        memory={50}
                        pods={8}
                        lastHeartbeat="2024-10-01 12:34:56"
                        selectedNode={selectedNode}
                        setSelectedNode={(id: number) => {setSelectedNode(id)}}
                    />
                    <NodeListItem
                        id={3}
                        name="gluon-worker-02"
                        status="NotReady"
                        ip="192.168.1.3"
                        role="worker"
                        cpu={30}
                        memory={50}
                        pods={8}
                        lastHeartbeat="2024-10-01 12:34:56"
                        selectedNode={selectedNode}
                        setSelectedNode={(id: number) => {setSelectedNode(id)}}
                    />
                    <NodeListItem
                        id={4}
                        name="gluon-worker-03"
                        status="Unknown"
                        ip="192.168.1.4"
                        role="worker"
                        cpu={30}
                        memory={50}
                        pods={8}
                        lastHeartbeat="2024-10-01 12:34:56"
                        selectedNode={selectedNode}
                        setSelectedNode={(id: number) => {setSelectedNode(id)}}
                    />
                </CardContainer>
                <CardContainer
                    title="Select A Node"
                    icon={<ServerCog className="w-5 h-5"/>}
                 >

                </CardContainer>
            </div>
        </div>
    </>
  )
}

export default NodesView
