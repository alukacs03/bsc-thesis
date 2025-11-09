interface NodeTableRowProps {
    handleNodeClick : (nodeId : string) => void
    nodeName : string
    nodeIP : string
    nodeStatus : 'online' | 'offline' | 'degraded'
    nodeRole : string
    lastSeen : string
}

const NodeTableRow = ({ handleNodeClick, nodeName, nodeIP, nodeStatus, nodeRole, lastSeen }: NodeTableRowProps) => {
  return (
        <tr 
            onClick={() => handleNodeClick('1')}
            className="border-b border-slate-100 hover:bg-blue-50 cursor-pointer transition-colors"
            >
            <td className="py-2 md:py-3 px-3 md:px-6 text-sm text-slate-700">{nodeName}</td>
            <td className="py-2 md:py-3 px-3 md:px-6 text-sm text-slate-600 hidden sm:table-cell">{nodeIP}</td>
            <td className="py-2 md:py-3 px-3 md:px-6">
                {nodeStatus === 'online' && (
                    <span className="bg-green-600 text-white px-2 py-1 rounded text-xs md:text-sm cursor-pointer hover:bg-green-700 whitespace-nowrap">online</span>
                ) || nodeStatus === 'offline' && (
                    <span className="bg-red-600 text-white px-2 py-1 rounded text-xs md:text-sm cursor-pointer hover:bg-red-700 whitespace-nowrap">offline</span>
                ) || nodeStatus === 'degraded' && (
                    <span className="bg-yellow-500 text-white px-2 py-1 rounded text-xs md:text-sm cursor-pointer hover:bg-yellow-600 whitespace-nowrap">degraded</span>
                )}
            </td>
            <td className="py-2 md:py-3 px-3 md:px-6 text-sm text-slate-600 hidden md:table-cell">{nodeRole}</td>
            <td className="py-2 md:py-3 px-3 md:px-6 text-sm text-slate-600 hidden lg:table-cell">{lastSeen}</td>
        </tr>
  )
}

export default NodeTableRow
