interface OverviewTabProps {
    nodeData: {
        id?: number;
        name?: string;
        hostname?: string;
        ip?: string;
        status?: string;
        role?: string;
        lastSeen?: string;
        cpu?: number;
        memory?: number;
        disk?: number;
        uptime?: string;
        version?: string;
        pods?: number;
        location?: string;
        cpuType?: string;
        memorySize?: string;
        diskSize?: string;
        os?: string;
        logs?: string[];
    }
}

const OverviewTab = ({ nodeData }: OverviewTabProps) => {
  return (
    <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
            <h3 className="text-lg text-slate-800 mb-4">System Information</h3>
            <div className="space-y-3">
            <div className="flex justify-between">
                <span className="text-slate-600">Hostname:</span>
                <span className="text-slate-800">{nodeData.hostname}</span>
            </div>
            <div className="flex justify-between">
                <span className="text-slate-600">IP Address:</span>
                <span className="text-slate-800">{nodeData.ip}</span>
            </div>
            <div className="flex justify-between">
                <span className="text-slate-600">Operating System:</span>
                <span className="text-slate-800">{nodeData.os}</span>
            </div>
            <div className="flex justify-between">
                <span className="text-slate-600">Architecture:</span>
                <span className="text-slate-800">{nodeData.cpuType}</span>
            </div>
            <div className="flex justify-between">
                <span className="text-slate-600">Agent Version:</span>
                <span className="text-slate-800">{nodeData.version}</span>
            </div>
            </div>
        </div>
        <div>
            <h3 className="text-lg text-slate-800 mb-4">Resource Usage</h3>
            <div className="space-y-4">
            <div>
                <div className="flex justify-between mb-1">
                <span className="text-sm text-slate-600">CPU Usage</span>
                <span className="text-sm text-slate-800">{nodeData.cpu}</span>
                </div>
                <div className="w-full bg-slate-200 rounded-full h-2">
                <div className="bg-blue-600 h-2 rounded-full" style={{ width: `${nodeData.cpu}%` }}></div>
                </div>
            </div>
            <div>
                <div className="flex justify-between mb-1">
                <span className="text-sm text-slate-600">Memory Usage</span>
                <span className="text-sm text-slate-800">{nodeData.memory}</span>
                </div>
                <div className="w-full bg-slate-200 rounded-full h-2">
                <div className="bg-green-600 h-2 rounded-full" style={{ width: `${nodeData.memory}%` }}></div>
                </div>
            </div>
            <div>
                <div className="flex justify-between mb-1">
                <span className="text-sm text-slate-600">Disk Usage</span>
                <span className="text-sm text-slate-800">{nodeData.disk}</span>
                </div>
                <div className="w-full bg-slate-200 rounded-full h-2">
                <div className="bg-yellow-600 h-2 rounded-full" style={{ width: `${nodeData.disk}%` }}></div>
                </div>
            </div>
            </div>
        </div>
        </div>
    </div>
  )
}

export default OverviewTab
