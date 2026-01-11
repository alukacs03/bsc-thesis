interface OSPFTableRowProps {
    rowKey: string
    routerId: string
    area: string
    state: string
    interface: string
    helloTimer: string
    deadTimer: string
    cost: number | string
    priority: number | string
}

function getStatusColor(state: string) {
  switch (state) {
    case 'Full':
      return 'bg-green-600 text-white';
    case 'Down':
      return 'bg-red-600 text-white';
    case 'Init':
    case '2-Way':
    case 'ExStart':
    case 'Exchange':
    case 'Loading':
      return 'bg-yellow-600 text-white';
    default:
      return 'bg-gray-600 text-white';
  }
}

const OSPFTableRow = (props: OSPFTableRowProps) => {
  return (
    <tr key={props.rowKey} className="border-b border-slate-100 hover:bg-blue-50">
        <td className="py-3 px-4">
            <div className="flex items-center space-x-3">
                <div className={`w-3 h-3 rounded-full ${
                props.state === 'Full' ? 'bg-green-500' :
                props.state === 'Down' ? 'bg-red-500' :
                'bg-yellow-500'
                }`}></div>
                <span className="text-slate-800">{props.routerId}</span>
            </div>
        </td>
        <td className="py-3 px-4 text-slate-600">{props.area}</td>
        <td className="py-3 px-4">
            <span className={`px-2 py-1 rounded text-sm ${getStatusColor(props.state)}`}>
                {props.state.toUpperCase()}
            </span>
        </td>
        <td className="py-3 px-4 text-slate-600">{props.interface}</td>
        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{props.helloTimer}</td>
        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{props.deadTimer}</td>
        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{props.cost}</td>
        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{props.priority}</td>
    </tr>
  )
}

export default OSPFTableRow
