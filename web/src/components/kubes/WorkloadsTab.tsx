import { getStatusColor } from "../../utils/Helpers";

interface WorkloadsTabProps {
    workloads?: { name: string; type: string; namespace: string; replicas: string; status: string; age: string; image: string; cluster: string; }[];
}

const WorkloadsTab = ({ workloads }: WorkloadsTabProps) => {
  return (
    <div className="space-y-4">
        <div className="flex justify-between items-center">
            <h4 className="text-lg text-slate-800">Deployments & StatefulSets</h4>
            <button className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm transition-colors">
                Deploy App
            </button>
        </div>
        <div className="overflow-x-auto">
            <table className="w-full">
            <thead>
                <tr className="border-b border-slate-200">
                <th className="text-left py-3 px-4 text-slate-700">Name</th>
                <th className="text-left py-3 px-4 text-slate-700">Type</th>
                <th className="text-left py-3 px-4 text-slate-700">Namespace</th>
                <th className="text-left py-3 px-4 text-slate-700">Replicas</th>
                <th className="text-left py-3 px-4 text-slate-700">Status</th>
                <th className="text-left py-3 px-4 text-slate-700">Age</th>
                </tr>
            </thead>
            <tbody>
                {!workloads || workloads.length === 0 ? (
                    <tr>
                        <td colSpan={6} className="py-3 px-4 text-slate-500 text-center">No workloads</td>
                    </tr>
                ) : (
                    workloads?.map((workload: { name: string; type: string; namespace: string; replicas: string; status: string; age: string; image: string; cluster: string; }, index: number) => (
                    <tr key={index} className="border-b border-slate-100 hover:bg-blue-50">
                        <td className="py-3 px-4 text-slate-700">{workload.name}</td>
                        <td className="py-3 px-4 text-slate-600">{workload.type}</td>
                        <td className="py-3 px-4 text-slate-600">{workload.namespace}</td>
                        <td className="py-3 px-4 text-slate-600">{workload.replicas}</td>
                        <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded text-xs ${getStatusColor(workload.status as string)}`}>
                            {workload.status}
                        </span>
                        </td>
                        <td className="py-3 px-4 text-slate-600">{workload.age}</td>
                    </tr>
                    ))
                )}
            </tbody>
            </table>
        </div>
    </div>
  )
}

export default WorkloadsTab
