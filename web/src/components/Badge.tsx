interface BadgeProps {
    className?: string
    children: React.ReactNode
}

const Badge = ({ className, children }: BadgeProps) => {
  return (
    <>
        <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full text-slate-800 ${className}`}>
            {children}
        </span>
    </>
  )
}

export default Badge

