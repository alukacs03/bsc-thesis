interface CardActivityProps {
    message : string
    timestamp : string
    alertLevel? : 'info' | 'warning' | 'error'
}

const CardActivity = ({ message, timestamp, alertLevel }: CardActivityProps) => {
  return (
    <div className="flex items-start space-x-3">
        <div className={`w-2 h-2 rounded-full mt-2 ${alertLevel === 'error' ? 'bg-red-500' : alertLevel === 'warning' ? 'bg-yellow-500' : 'bg-green-500'}`}></div>
        <div className="flex-1">
            <p className="text-sm text-slate-800">{message}</p>
            <p className="text-xs text-slate-500">{timestamp}</p>
        </div>
    </div>
  )
}

export default CardActivity
