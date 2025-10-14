import React from 'react'

interface CardWithIconProps {
    title: string
    value: React.ReactNode
    hint?: string
    onClick?: () => void
    textColorClass?: string
    valueColorClass?: string
    outlineColorClass?: string
    icon?: React.ReactNode
}

const CardWithIcon = ({
    title,
    value,
    hint,
    onClick,
    textColorClass = 'text-slate-600',
    valueColorClass = 'text-slate-800',
    outlineColorClass = '',
    icon
}: CardWithIconProps) => {
  return (
    <>
            <button 
                onClick={onClick}
                className={`bg-white rounded-lg shadow-sm p-4 md:p-6 ${outlineColorClass} hover:shadow-md hover:scale-105 transition-all duration-200 text-left`}
            >
            <div className="flex items-center justify-between">
                <div>
                    <p className={`text-sm ${textColorClass}`}>{title}</p>
                    <p className={`text-2xl md:text-3xl ${valueColorClass}`}>{value}</p>
                </div>
                <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                    {icon}
                </div>
            </div>
            <p className="text-xs text-slate-500 mt-2">{hint}</p>
            </button>
    </>

  )
}


export default CardWithIcon
