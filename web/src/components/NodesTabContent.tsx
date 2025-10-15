import OverviewTab from "./nodes/OverviewTab";
import LogsTab from "./nodes/LogsTab";

interface NodesTabContentProps {
    selectedTab: string;
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
        logs?: string[];
    }
}

const NodesTabContent = ({ selectedTab, node }: NodesTabContentProps) => {
  return (
    <>
    <div className="p-6">
      {selectedTab === "Overview" && <OverviewTab node={node} />}
      {selectedTab === "Logs" && <LogsTab logs={node.logs} />}
    </div>
    </>
  )
}

export default NodesTabContent
