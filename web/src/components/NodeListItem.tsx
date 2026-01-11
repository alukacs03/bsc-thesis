import { getStatusColor } from "../utils/Helpers"
import { formatPercent } from "@/utils/format"
interface NodeListItemProps {
    id : number
    name : string
    status : string
    ip : string
    role : string
    cpu : number | null | undefined
    memory : number | null | undefined
    pods : number
    lastSeen : string
    selectedNode : string | number | null
    setSelectedNode : (id: number) => void
}

function getMetricColor(value: number | null | undefined, type: 'cpu' | 'memory') {
    if (value === null || value === undefined || !Number.isFinite(value)) return 'text-slate-500';
    if (type === 'cpu') {
        if (value < 50) return 'text-green-700';
        if (value < 80) return 'text-yellow-700';
        return 'text-red-700';
    } else {
        if (value < 60) return 'text-green-700';
        if (value < 85) return 'text-yellow-700';
        return 'text-red-700';
    }
}

const NodeListItem = ({ id, name, status, ip, role, cpu, memory, pods, lastSeen, selectedNode, setSelectedNode }: NodeListItemProps) => {
  return (
    <>
        <div
        key={id}
        className={`p-4 hover:bg-blue-50 cursor-pointer transition-colors ${
            selectedNode === id ? 'bg-blue-50 border-r-4 border-blue-500' : ''
        }`}
        onClick={() => setSelectedNode(id)}
        >
        <div className="flex items-center justify-between">
            <div className="flex-1">
            <div className="flex items-center space-x-3">
                <h4 className="text-slate-800">{name}</h4>
                <span className={`px-2 py-1 rounded text-xs ${getStatusColor(status)}`}>
                {status}
                </span>
            </div>
            <p className="text-sm text-slate-600 mt-1">{ip} • {role}</p>
            <div className="flex items-center space-x-4 mt-2">
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">CPU:</span>
                <span className={`text-xs ${getMetricColor(cpu, 'cpu')}`}>{formatPercent(cpu)}</span>
                </div>
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">MEM:</span>
                <span className={`text-xs ${getMetricColor(memory, 'memory')}`}>{formatPercent(memory)}</span>
                </div>
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">PODS:</span>
                <span className="text-xs text-slate-700">{pods}</span>
                </div>
            </div>
            </div>
            <div className="text-right">
            <p className="text-xs text-slate-500">Last seen</p>
            <p className="text-xs text-slate-700">{lastSeen}</p>
            </div>
        </div>
        </div>
    </>
  )
}

export default NodeListItem
