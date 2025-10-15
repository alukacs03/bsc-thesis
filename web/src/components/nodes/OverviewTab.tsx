import { useNavigate } from "react-router-dom";
import { getStatusColor, getMetricColor } from "../../utils/Helpers";

interface OverviewTabProps {
    node: {
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

const OverviewTab = ({ node }: OverviewTabProps) => {
    const navigate = useNavigate();
  return (
        <div className="p-6">
            <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
                <div>
                <p className="text-sm text-slate-600">Status</p>
                <span className={`inline-block px-2 py-1 rounded text-sm mt-1 ${getStatusColor(node?.status)}`}>
                    {node?.status}
                </span>
                </div>
                <div>
                <p className="text-sm text-slate-600">Role</p>
                <p className="text-sm text-slate-800 mt-1">{node?.role}</p>
                </div>
                <div>
                <p className="text-sm text-slate-600">IP Address</p>
                <p className="text-sm text-slate-800 mt-1">{node?.ip}</p>
                </div>
                <div>
                <p className="text-sm text-slate-600">Location</p>
                <p className="text-sm text-slate-800 mt-1">{node?.location}</p>
                </div>
                <div>
                <p className="text-sm text-slate-600">Uptime</p>
                <p className="text-sm text-slate-800 mt-1">{node?.uptime}</p>
                </div>
                <div>
                <p className="text-sm text-slate-600">Version</p>
                <p className="text-sm text-slate-800 mt-1">{node?.version}</p>
                </div>
            </div>

            <div className="space-y-4">
                <div>
                <div className="flex justify-between text-sm mb-2">
                    <span className="text-slate-600">CPU Usage</span>
                    <span className={getMetricColor(node?.cpu, 'cpu')}>{node?.cpu}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                    className={`h-2 rounded-full ${node?.cpu > 80 ? 'bg-red-500' : node?.cpu > 60 ? 'bg-yellow-500' : 'bg-green-500'}`}
                    style={{ width: `${node?.cpu}%` }}
                    ></div>
                </div>
                </div>

                <div>
                <div className="flex justify-between text-sm mb-2">
                    <span className="text-slate-600">Memory Usage</span>
                    <span className={getMetricColor(node?.memory, 'memory')}>{node?.memory}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                    className={`h-2 rounded-full ${node?.memory > 80 ? 'bg-red-500' : node?.memory > 60 ? 'bg-yellow-500' : 'bg-green-500'}`}
                    style={{ width: `${node?.memory}%` }}
                    ></div>
                </div>
                </div>

                <div>
                <div className="flex justify-between text-sm mb-2">
                    <span className="text-slate-600">Disk Usage</span>
                    <span className={getMetricColor(node?.disk, 'disk')}>{node?.disk}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                    className={`h-2 rounded-full ${node?.disk > 85 ? 'bg-red-500' : node?.disk > 70 ? 'bg-yellow-500' : 'bg-green-500'}`}
                    style={{ width: `${node?.disk}%` }}
                    ></div>
                </div>
                </div>
            </div>

            <div className="flex space-x-2">
                <button 
                onClick={() => navigate(`/nodes/${node?.id}`, )}
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm transition-colors flex items-center space-x-2"
                >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span>Manage Node</span>
                </button>
                <a href={`ssh://${node?.ip}`} className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg text-sm transition-colors"
                >
                SSH Connect
                </a>
            </div>
            </div>
        </div>
  )
}

export default OverviewTab
