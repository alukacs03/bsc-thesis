import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import { Globe, Server, FolderPen, HardDrive } from "lucide-react"
import ClusterListItem from "../components/ClusterListItem"
import { useState } from "react"

const KubernetesView = () => {
  const [selectedCluster, setSelectedCluster] = useState<number | null>(1);
  return (
    <>
        <div className="space-y-6">
            <div className="grid grid-cols1 md:grid-cols-4 gap-6">
                <CardWithIcon
                    title="Total Clusters"
                    value="3"
                    textColorClass="text-slate-600"
                    valueColorClass="text-blue-600"
                    iconBGColorClass="bg-blue-100"
                    icon={<Server className="w-6 h-6 text-slate-800"/>}
                />
                <CardWithIcon
                    title="Total Pods"
                    value="2"
                    textColorClass="text-slate-600"
                    valueColorClass="text-green-600"
                    iconBGColorClass="bg-green-100"
                    icon={<HardDrive className="w-6 h-6 text-green-600"/>}
                />
                <CardWithIcon
                    title="Total Services"
                    value="1"
                    textColorClass="text-slate-600"
                    valueColorClass="text-red-600"
                    iconBGColorClass="bg-red-100"
                    icon={<Globe className="w-6 h-6 text-red-600"/>}
                />
                <CardWithIcon
                    title="Total Namespaces"
                    value="4"
                    textColorClass="text-slate-600"
                    valueColorClass="text-yellow-600"
                    iconBGColorClass="bg-yellow-100"
                    icon={<FolderPen className="w-6 h-6 text-yellow-600"/>}
                />
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <CardContainer
                    title="Clusters"
                    icon={<Server className="w-5 h-5"/>}
                    button={
                        <button
                            className="bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition"
                            onClick={() => alert('Add Cluster clicked (would navigate to add cluster form)')}
                        >
                            Add Cluster
                        </button>
                    }
                    noPadding={true}
                >
                <div className="divide-y divide-slate-200">
                    <ClusterListItem
                        id={1}
                        name="prod-cluster-01"
                        status="Active"
                        version="v1.24.3"
                        location="us-east-1"
                        nodes={5}
                        pods={20}
                        services={10}
                        selectedCluster={selectedCluster}
                        setSelectedCluster={(id: number) => {setSelectedCluster(id)}}
                    />
                    <ClusterListItem
                        id={2}
                        name="staging-cluster-01"
                        status="Inactive"
                        version="v1.23.8"
                        location="us-west-2"
                        nodes={3}
                        pods={15}
                        services={7}
                        selectedCluster={selectedCluster}
                        setSelectedCluster={(id: number) => {setSelectedCluster(id)}}
                    />
                    <ClusterListItem
                        id={3}
                        name="dev-cluster-01"
                        status="Error"
                        version="v1.22.10"
                        location="eu-central-1"
                        nodes={2}
                        pods={8}
                        services={4}
                        selectedCluster={selectedCluster}
                        setSelectedCluster={(id: number) => {setSelectedCluster(id)}}
                    />
                </div>

                </CardContainer>
                <CardContainer
                    title="Select A Cluster"
                    icon={<Globe className="w-5 h-5"/>}
                    colSpan="lg:col-span-2"
                 >
                    <p className="text-sm text-slate-600">Select a cluster to view its details</p>
                </CardContainer>
            </div>
        </div>
    </>
  )
}

export default KubernetesView
