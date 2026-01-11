import OverviewTab from "./nodemgmt/OverviewTab";
import ServicesTab from "./nodemgmt/ServicesTab";
import SSHTab from "./nodemgmt/SSHTab";
import LogsTab from "./nodemgmt/LogsTab";
import NetworkingTab from "./nodemgmt/NetworkingTab";
import type { SystemService } from "@/services/types/service";

interface NodeManagementTabContentProps {
    selectedTab: string;
    node : {
            id: number;
            name: string;
            hostname: string;
            ip: string;
            status: string;
            role: string;
            lastSeen: string;
            cpu: number | null | undefined;
            memory: number | null | undefined;
            disk: number | null | undefined;
            uptime: string;
            version: string;
            pods: number;
            location: string;
            cpuType: string;
            memorySize: string;
            diskSize: string;
            os: string;
            logs?: string[];
            diskUsedBytes?: number;
            diskTotalBytes?: number;
            systemUsers?: string[];
            services?: SystemService[];
            filesystemMounts?: {
                mountpoint: string;
                device: string;
                type: string;
                size: number;
                used: number;
            }[];
    }
}


const NodeManagementTabContent = ({ selectedTab, node }: NodeManagementTabContentProps) => {
  return (
    <>
        <div className="p-6">
            {selectedTab === "Overview" && <OverviewTab nodeData={node} />}
            {selectedTab === "SSH Keys" && <SSHTab nodeId={node.id} systemUsers={node.systemUsers} />}
            {selectedTab === "Services" && <ServicesTab nodeId={node.id} services={node.services} />}
            {selectedTab === "Networking" && <NetworkingTab nodeId={node.id} />}
            {selectedTab === "Logs" && <LogsTab nodeId={node.id} enabled={selectedTab === "Logs"} />}
        </div>
    </>
  )
}

export default NodeManagementTabContent
