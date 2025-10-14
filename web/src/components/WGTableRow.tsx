import Badge from "./Badge"

interface WGTableRowProps {
    rowKey?: string | number
    status :  'connected' | 'down' | 'flapping'
    name: string
    pubKey : string
    endpoint : string
    latestHandshake : string
    transferRx : string
    transferTx : string
}

function truncateKey(key: string) {
    if (key.length <= 10) return key;
    return key.slice(0, 6) + '...' + key.slice(-4);
}

function getStatusBadge(status: string) {
  switch (status) {
    case 'connected':
      return <Badge className="bg-green-600 text-white">CONNECTED</Badge>;
    case 'flapping':
      return <Badge className="bg-yellow-600 text-white">FLAPPING</Badge>;
    case 'down':
      return <Badge className="bg-red-600 text-white">DOWN</Badge>;
    default:
      return <Badge className="bg-gray-600 text-white">UNKNOWN</Badge>;
  }
}

const WGTableRow = ({ rowKey, status, name, pubKey, endpoint, latestHandshake, transferRx, transferTx }: WGTableRowProps) => {
  return (
    <>
        <tr key={rowKey} className="border-b border-slate-100 hover:bg-blue-50">
            <td className="py-3 px-4">
                <div className="flex items-center space-x-3">
                      <div className={`w-3 h-3 rounded-full ${
                        status === 'connected' ? 'bg-green-500' :
                        status === 'flapping' ? 'bg-yellow-500' :
                        'bg-red-500'
                      }`}></div>
                      <span className="text-slate-800">{name}</span>        
                </div>
            </td>
            <td className="py-3 px-4">
                <code className="text-sm bg-slate-100 px-2 py-1 rounded text-slate-600">
                    {truncateKey(pubKey)}
                </code>
            </td>
            <td className="py-3 px-4 text-slate-600">{endpoint}</td>
            <td className="py-3 px-4">{getStatusBadge(status)}</td>
            <td className="py-3 px-4 text-slate-600">{latestHandshake}</td>
            <td className="py-3 px-4 text-slate-600">
                <div className="text-sm">
                    <div className="flex items-center space-x-2">
                        <span>↓</span>
                        <span>{transferRx}</span>
                    </div>
                    <div className="flex items-center space-x-2">
                        <span>↑</span>
                        <span>{transferTx}</span>
                    </div>
                </div>
            </td>
            <td className="py-3 px-4">
                <div className="flex space-x-2">
                    <button 
                        className="text-blue-600 hover:text-blue-800 text-sm"
                        onClick={() => alert(`Restarting peer ${name}...`)}    
                    >
                        Restart
                    </button>
                    <button 
                        className="text-red-600 hover:text-red-800 text-sm"
                        onClick={() => alert(`Removing peer ${name}...`)}
                    >
                        Remove
                    </button>
                </div>
            </td>
        </tr> 
    </>
  )
}

export default WGTableRow
