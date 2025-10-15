import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import { Globe, Server, FolderPen, HardDrive } from "lucide-react"
import ClusterListItem from "../components/ClusterListItem"
import KubernetesNavBar from "../components/KubernetesNavBar"
import { useState } from "react"
import KubernetesTabContent from "../components/KubernetesTabContent"


const KubernetesView = () => {
    const [selectedCluster, setSelectedCluster] = useState<number | null>(1);
    const [selectedTab, setSelectedTab] = useState<"Overview" | "Workloads" | "Networking" | "Storage" | "Logs">("Overview");

    type ClusterStatus = "Active" | "Inactive" | "Error";

    const clusters: {
    id: number;
    name: string;
    status: ClusterStatus;
    version: string;
    nodes: number;
    pods: number;
    services: number;
    location: string;
    created: string;
    lastUpdate: string;
    namespaces?: { name: string; status: string; pods: number; age: string; }[];
    workloads?: { name: string; type: string; namespace: string; replicas: string; status: string; age: string; image: string; cluster: string; }[];
    servicesList?: { name: string; type: string; namespace: string; clusterIP: string; externalIP?: string; ports?: string; age: string; cluster: string; }[];
    storageList?: { name: string; type: string; namespace: string; capacity: string; accessModes: string; age: string; cluster: string; }[];
    }[] = [
        {
            id: 1,
            name: 'production-cluster',
            status: 'Active',
            version: 'v1.28.2',
            nodes: 4,
            pods: 57,
            services: 12,
            location: 'us-east-1',
            created: '2023-12-01',
            lastUpdate: '2024-01-15 10:30:00',
            namespaces: [
                { name: 'default', status: 'Active', pods: 15, age: '45d' },
                { name: 'production', status: 'Active', pods: 18, age: '30d' },
                { name: 'staging', status: 'Active', pods: 12, age: '15d' },
                { name: 'database', status: 'Active', pods: 3, age: '30d' },
                { name: 'cache', status: 'Active', pods: 4, age: '7d' },
                { name: 'monitoring', status: 'Active', pods: 5, age: '20d' }
            ],
            workloads: [
                { name: 'nginx-deployment', type: 'Deployment', namespace: 'default', replicas: '3/3', status: 'Running', age: '15d', image: 'nginx:1.20', cluster: 'production-cluster' },
                { name: 'auth-service', type: 'Deployment', namespace: 'production', replicas: '2/3', status: 'Pending', age: '2d', image: 'auth-service:v2.1.0', cluster: 'production-cluster' },
                { name: 'postgres-db', type: 'StatefulSet', namespace: 'database', replicas: '1/1', status: 'Running', age: '30d', image: 'postgres:13', cluster: 'production-cluster' }
            ],
            servicesList: [
                { name: 'nginx-service', type: 'LoadBalancer', namespace: 'default', clusterIP: '10.96.1.15', externalIP: '203.0.113.25', ports: '80:30080/TCP', age: '15d', cluster: 'production-cluster' },
                { name: 'auth-service-svc', type: 'ClusterIP', namespace: 'production', clusterIP: '10.96.1.16', age: '2d', cluster: 'production-cluster' },
                { name: 'postgres-service', type: 'ClusterIP', namespace: 'database', clusterIP: '10.96.1.17', age: '30d', cluster: 'production-cluster' }
            ]
        },
        {
            id: 2,
            name: 'staging-cluster',
            status: 'Inactive',
            version: 'v1.28.1',
            nodes: 2,
            pods: 23,
            services: 8,
            location: 'us-west-2',
            created: '2024-01-10',
            lastUpdate: '2024-01-15 09:45:00',
            namespaces: [
                { name: 'default', status: 'Inactive', pods: 5, age: '10d' },
                { name: 'production', status: 'Inactive', pods: 8, age: '5d' },
                { name: 'staging', status: 'Inactive', pods: 4, age: '2d' }
            ],
            workloads: [
                { name: 'redis-cache', type: 'Deployment', namespace: 'cache', replicas: '2/2', status: 'Running', age: '7d', image: 'redis:6.2', cluster: 'staging-cluster' }
            ],
            servicesList: [
                { name: 'redis-service', type: 'ClusterIP', namespace: 'cache', clusterIP: '10.96.1.18', age: '7d', cluster: 'staging-cluster' }
            ]
        },
        {
            id: 3,
            name: 'dev-cluster',
            status: 'Error',
            version: 'v1.27.5',
            nodes: 3,
            pods: 34,
            services: 10,
            location: 'eu-central-1',
            created: '2023-11-20',
            lastUpdate: '2024-01-14 16:20:00',
            namespaces: [
                { name: 'default', status: 'Error', pods: 5, age: '10d' },
                { name: 'production', status: 'Error', pods: 8, age: '5d' },
                { name: 'staging', status: 'Error', pods: 4, age: '2d' }
            ],
            workloads: [
                { name: 'monitoring-agent', type: 'DaemonSet', namespace: 'monitoring', replicas: '3/3', status: 'Running', age: '20d', image: 'monitoring-agent:v1.5.0', cluster: 'dev-cluster' }
            ],
            servicesList: [
                { name: 'monitoring-service', type: 'ClusterIP', namespace: 'monitoring', clusterIP: '10.96.1.19', age: '20d', cluster: 'dev-cluster' }
            ],
            storageList: [
                { name: 'dev-pv-1', type: 'PersistentVolume', namespace: 'default', capacity: '100Gi', accessModes: 'ReadWriteOnce', age: '10d', cluster: 'dev-cluster' },
                { name: 'dev-pv-2', type: 'PersistentVolume', namespace: 'production', capacity: '200Gi', accessModes: 'ReadWriteMany', age: '5d', cluster: 'dev-cluster' }
            ]
        },
        {
            id: 4,
            name: 'test-cluster',
            status: 'Active',
            version: 'v1.28.0',
            nodes: 1,
            pods: 5,
            services: 2,
            location: 'ap-southeast-1',
            created: '2024-02-01',
            lastUpdate: '2024-02-10 11:00:00',
            namespaces: [
                { name: 'default', status: 'Active', pods: 2, age: '5d' },
                { name: 'testing', status: 'Active', pods: 3, age: '3d' }
            ]
        }
    ];


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
                    {
                        clusters.map((cluster) => (
                            <ClusterListItem
                                id={cluster.id}
                                name={cluster.name}
                                status={cluster.status}
                                version={cluster.version}
                                location={cluster.location}
                                nodes={cluster.nodes}
                                pods={cluster.pods}
                                services={cluster.services}
                                selectedCluster={selectedCluster}
                                setSelectedCluster={(id: number) => {setSelectedCluster(id)}}
                            />
                        ))
                    }
                </div>

                </CardContainer>
                <CardContainer
                    title={selectedCluster ? `${clusters[selectedCluster - 1].name} Details` : "Select A Cluster"}
                    icon={<Globe className="w-5 h-5"/>}
                    colSpan="lg:col-span-2"
                    noPadding={true}
                >
                    {selectedCluster ? (
                        <div className="border-b border-slate-200">
                            <KubernetesNavBar setSelectedTab={setSelectedTab} selectedTab={selectedTab}/>
                            <KubernetesTabContent selectedTab={selectedTab} cluster={clusters[selectedCluster - 1]} />
                        </div>
                    ) : (
                        <p className="text-sm text-slate-600">Select a cluster to view its details</p>
                    )}
                </CardContainer>
            </div>
        </div>
    </>
  )
}

export default KubernetesView
