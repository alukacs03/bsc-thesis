import React from 'react'

interface TableProps {
  children? : React.ReactNode
  columns ?: string[]
}

const Table = ({ columns, children }: TableProps) => {
  return (
    <>
      <div className = "overflow-x-auto">
        <table className="w-full min-w-[640px]">
            <thead>
              <tr className="border-b border-slate-200">
                {columns?.map((col, index) => (
                  <th 
                    key={index}
                    className="text-left py-2 md:py-3 px-3 md:px-6 text-sm text-slate-700"
                  >
                    {col}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {children}
            </tbody>
        </table>
      </div>
    </>
  )
}

export default Table
