import CardContainer from "@/components/CardContainer";
import DetailsNavBar from "@/components/DetailsNavBar";
import { ErrorMessage } from "@/components/ErrorMessage";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import ClusterNetworkingTab from "@/components/kubernetes/ClusterNetworkingTab";
import ClusterNodesTab from "@/components/kubernetes/ClusterNodesTab";
import ClusterOverviewTab from "@/components/kubernetes/ClusterOverviewTab";
import ClusterStorageTab from "@/components/kubernetes/ClusterStorageTab";
import ClusterWorkloadsTab from "@/components/kubernetes/ClusterWorkloadsTab";
import { useKubernetesCluster } from "@/services/hooks/useKubernetesCluster";
import { useNodes } from "@/services/hooks/useNodes";
import { kubernetesAPI } from "@/services/api/kubernetes";
import { Boxes, HardDrive, LayoutDashboard, Network, Server } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";

const tabs = ["Overview", "Nodes", "Networking", "Workloads", "Storage"];
const tabIcons = [
  <LayoutDashboard key="overview" className="w-4 h-4 mr-2" />,
  <Server key="nodes" className="w-4 h-4 mr-2" />,
  <Network key="networking" className="w-4 h-4 mr-2" />,
  <Boxes key="workloads" className="w-4 h-4 mr-2" />,
  <HardDrive key="storage" className="w-4 h-4 mr-2" />,
];

export default function KubernetesView() {
  const [selectedTab, setSelectedTab] = React.useState<string>("Overview");
  const { data: nodes, loading, error, refetch } = useNodes({ pollingInterval: 30000 });
  const { data: clusterResp, loading: clusterLoading, error: clusterError, refetch: refetchCluster } = useKubernetesCluster({ pollingInterval: 30000 });
  const [refreshingJoinTokens, setRefreshingJoinTokens] = React.useState(false);

  if ((loading && !nodes) || (clusterLoading && !clusterResp)) {
    return <LoadingSpinner />;
  }

  if (error || clusterError) {
    return <ErrorMessage message={(error ?? clusterError)!.message} onRetry={() => { refetch(); refetchCluster(); }} />;
  }

  const safeNodes = nodes ?? [];
  const cluster = clusterResp?.cluster ?? null;

  const handleRefreshJoinTokens = async () => {
    try {
      setRefreshingJoinTokens(true);
      await kubernetesAPI.refreshJoinTokens();
      toast.success("Requested join token refresh");
      refetchCluster();
    } catch (err) {
      toast.error(handleAPIError(err, "refresh join tokens"));
    } finally {
      setRefreshingJoinTokens(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl text-slate-800">Kubernetes</h1>
        <p className="text-slate-600">Single-cluster view for the Gluon cluster.</p>
      </div>

      <CardContainer title="Gluon Kubernetes Cluster" icon={<Server className="w-5 h-5" />} noPadding>
        <div className="border-b border-slate-200">
          <DetailsNavBar tabs={tabs} icons={tabIcons} selectedTab={selectedTab} setSelectedTab={setSelectedTab} />
        </div>

        <div className="p-4 md:p-6">
          {selectedTab === "Overview" && (
            <ClusterOverviewTab
              cluster={cluster}
              nodes={safeNodes}
              onRefreshJoinTokens={cluster ? handleRefreshJoinTokens : undefined}
              refreshingJoinTokens={refreshingJoinTokens}
            />
          )}
          {selectedTab === "Nodes" && <ClusterNodesTab nodes={safeNodes} />}
          {selectedTab === "Networking" && <ClusterNetworkingTab />}
          {selectedTab === "Workloads" && <ClusterWorkloadsTab />}
          {selectedTab === "Storage" && <ClusterStorageTab nodes={safeNodes} />}
        </div>
      </CardContainer>
    </div>
  );
}
