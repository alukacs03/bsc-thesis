import React from 'react'

interface CardContainerProps {
    children? : React.ReactNode
    title : string
    noPadding? : boolean
}

const CardContainer = ({ children, title, noPadding }: CardContainerProps) => {
  return (
    <>
        <div className="bg-white rounded-lg shadow-sm">
            <div className="px-4 md:px-6 py-3 md:py-4 border-b border-slate-200">
                <h3 className="text-base md:text-lg text-slate-800">{title}</h3>
            </div>
            {noPadding ? children : <div className="p-4 md:p-6 space-y-3 md:space-y-4">{children}</div>}
        </div>

    </>
  )
}

export default CardContainer
