
interface ClusterListItemProps {
    id : number
    name : string
    status : 'Active' | 'Inactive' | 'Error'
    version : string
    location : string
    nodes : number
    pods : number
    services : number
    selectedCluster : string | number | null
    setSelectedCluster : (id: number) => void
}

function getStatusColor(status: string) {
  switch (status) {
    case 'Active':
      return 'bg-green-600 text-white';
    case 'Inactive':
      return 'bg-yellow-600 text-white';
    case 'Error':
      return 'bg-red-600 text-white';
    default:
      return 'bg-gray-600 text-white';
  }
}



const ClusterListItem = ({ id, name, status, version, location, nodes, pods, services, selectedCluster, setSelectedCluster }: ClusterListItemProps) => {
  return (
    <div
        key={id}
        className={`p-4 hover:bg-blue-50 cursor-pointer transition-colors ${
            selectedCluster === id ? 'bg-blue-50 border-r-4 border-blue-500' : ''
        }`}
        onClick={() => setSelectedCluster(id)}
        >
        <div className="flex items-center justify-between">
            <div className="flex-1">
            <div className="flex items-center space-x-3">
                <h4 className="text-slate-800">{name}</h4>
                <span className={`px-2 py-1 rounded text-xs ${getStatusColor(status)}`}>
                {status}
                </span>
            </div>
            <p className="text-sm text-slate-600 mt-1">{version} • {location}</p>
            <div className="flex items-center space-x-4 mt-2">
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">Nodes:</span>
                <span className="text-xs text-slate-700">{nodes}</span>
                </div>
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">Pods:</span>
                <span className="text-xs text-slate-700">{pods}</span>
                </div>
                <div className="flex items-center space-x-1">
                <span className="text-xs text-slate-500">Services:</span>
                <span className="text-xs text-slate-700">{services}</span>
                </div>
            </div>
            </div>
        </div>
    </div>
  )
}

export default ClusterListItem
