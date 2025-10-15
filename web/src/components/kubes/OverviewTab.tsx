import { getStatusColor } from "../../utils/Helpers";

interface OverviewTabProps {
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
    }
}


const OverviewTab = ({ cluster }: OverviewTabProps) => {
  return (
    <div className="space-y-6">
        <div className="grid grid-cols-2 gap-6">
            <div>
            <p className="text-sm text-slate-600">Cluster Status</p>
            <span className={`inline-block px-2 py-1 rounded text-sm mt-1 ${getStatusColor(cluster.status)}`}>
                {cluster.status}
            </span>
            </div>
            <div>
            <p className="text-sm text-slate-600">Kubernetes Version</p>
            <p className="text-sm text-slate-800 mt-1">{cluster.version}</p>
            </div>
            <div>
            <p className="text-sm text-slate-600">Location</p>
            <p className="text-sm text-slate-800 mt-1">{cluster.location}</p>
            </div>
            <div>
            <p className="text-sm text-slate-600">Created</p>
            <p className="text-sm text-slate-800 mt-1">{cluster.created}</p>
            </div>
        </div>

        <div className="grid grid-cols-3 gap-6">
            <div className="bg-blue-50 rounded-lg p-4">
            <div className="flex items-center justify-between">
                <div>
                <p className="text-sm text-blue-600">Nodes</p>
                <p className="text-2xl text-blue-700">{cluster.nodes}</p>
                </div>
                <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
                </svg>
                </div>
            </div>
            </div>
            <div className="bg-green-50 rounded-lg p-4">
            <div className="flex items-center justify-between">
                <div>
                <p className="text-sm text-green-600">Pods</p>
                <p className="text-2xl text-green-700">{cluster.pods}</p>
                </div>
                <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                </svg>
                </div>
            </div>
            </div>
            <div className="bg-purple-50 rounded-lg p-4">
            <div className="flex items-center justify-between">
                <div>
                <p className="text-sm text-purple-600">Services</p>
                <p className="text-2xl text-purple-700">{cluster.services}</p>
                </div>
                <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
                </div>
            </div>
            </div>
        </div>

        <div className="bg-slate-50 rounded-lg p-4">
            <h4 className="text-lg text-slate-800 mb-4">Namespaces</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {cluster.namespaces?.slice(0, 6).map((ns) => (
                <div key={ns.name} className="flex items-center justify-between p-3 bg-white rounded-lg">
                <div>
                    <p className="text-sm text-slate-800">{ns.name}</p>
                    <p className="text-xs text-slate-600">{ns.pods} pods • {ns.age}</p>
                </div>
                <span className={`px-2 py-1 rounded text-xs ${getStatusColor(ns.status)}`}>
                    {ns.status}
                </span>
                </div>
            ))}
            </div>
        </div>
    </div>
  )
}

export default OverviewTab
