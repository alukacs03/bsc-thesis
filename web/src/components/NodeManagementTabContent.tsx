import OverviewTab from "./nodemgmt/OverviewTab";
import ServicesTab from "./nodemgmt/ServicesTab";
import SSHTab from "./nodemgmt/SSHTab";
import FilesystemTab from "./nodemgmt/FilesystemTab";
import LogsTab from "./nodemgmt/LogsTab";

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
            cpu: number;
            memory: number;
            disk: number;
            uptime: string;
            version: string;
            pods: number;
            location: string;
            cpuType: string;
            memorySize: string;
            diskSize: string;
            os: string;
            logs?: string[];
            sshKeys? : {
                user: string;
                name: string;
                type: string;
                bits: number;
                fingerprint: string;
                publicKey: string;
                created: string;
                lastUsed: string;
                status: string;
                id: number;
            }[],
            services?: {
                name: string;
                status: string;
                enabled: boolean;
                description: string;
            }[],
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
            {selectedTab === "SSH Keys" && <SSHTab sshKeys={node.sshKeys} />}
            {selectedTab === "Services" && <ServicesTab services={node.services} />}
            {selectedTab === "Filesystem" && <FilesystemTab filesystemMounts={node.filesystemMounts} />}
            {selectedTab === "Logs & Terminal" && <LogsTab />}
        </div>
    </>
  )
}

export default NodeManagementTabContent
