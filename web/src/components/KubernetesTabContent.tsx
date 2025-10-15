import OverviewTab from "./kubes/OverviewTab";
import WorkloadsTab from "./kubes/WorkloadsTab";
import NetworkingTab from "./kubes/NetworkingTab";
import StorageTab from "./kubes/StorageTab";
import LogsTab from "./kubes/LogsTab";

interface KubernetesTabContentProps {
    selectedTab: string;
    cluster: {
        name: string;
        status: string;
        version: string;
        nodes: number;
        pods: number;
        services: number;
        location: string;
        created: string;
        namespaces?: { name: string; status: string; pods: number; age: string; }[];
        workloads?: { name: string; type: string; namespace: string; replicas: string; status: string; age: string; image: string; cluster: string; }[];
        servicesList?: { name: string; type: string; namespace: string; clusterIP: string; externalIP?: string; ports?: string; age: string; cluster: string; }[];
        storageList?: { name: string; type: string; namespace: string; capacity: string; accessModes: string; age: string; cluster: string; }[];
    }
}

const KubernetesTabContent = ({ selectedTab, cluster }: KubernetesTabContentProps) => {
  return (
    <>
    <div className="p-6">
      {selectedTab === "Overview" && <OverviewTab cluster={cluster} />}
      {selectedTab === "Workloads" && <WorkloadsTab workloads={cluster.workloads} />}
      {selectedTab === "Networking" && <NetworkingTab servicesList={cluster.servicesList} />}
      {selectedTab === "Storage" && <StorageTab storageList={cluster.storageList} />}
      {selectedTab === "Logs" && <LogsTab workloads={cluster.workloads} />}
    </div>
    </>
  )
}

export default KubernetesTabContent
