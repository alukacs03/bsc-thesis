import React from 'react'

interface CardActionProps {
    onClick?: () => void
    icon?: React.ReactNode
    title : string
    subtitle : string
}

const CardAction = ({ onClick, icon, title, subtitle }: CardActionProps) => {
  return (
    <button 
        onClick={onClick}
        className="w-full flex items-center space-x-3 p-3 rounded-lg border border-slate-200 hover:border-blue-300 hover:bg-blue-50 transition-colors"
    >
        <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
        {icon}
        </div>
        <div className="flex-1 text-left">
        <p className="text-sm text-slate-800">{title}</p>
        <p className="text-xs text-slate-500">{subtitle}</p>
        </div>
    </button>
  )
}

export default CardAction
