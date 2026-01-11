import OverviewTab from "./nodes/OverviewTab";
import LogsTab from "./nodes/LogsTab";
import NetworkingTab from "./nodes/NetworkingTab";

interface NodesTabContentProps {
    selectedTab: string;
    onDelete?: () => void;
    node : {
        id: number;
        name: string;
        status: string;
        role: string;
        ip: string;
        location: string;
        uptime: string;
        version: string;
        cpu: number;
        memory: number;
        disk: number;
    }
}

const NodesTabContent = ({ selectedTab, node }: NodesTabContentProps) => {
  return (
    <>
    <div className="p-6">
      {selectedTab === "Overview" && <OverviewTab node={node} />}
      {selectedTab === "Networking" && <NetworkingTab nodeId={node.id} />}
      {selectedTab === "Logs" && <LogsTab nodeId={node.id} enabled={selectedTab === "Logs"} />}
    </div>
    </>
  )
}

export default NodesTabContent
