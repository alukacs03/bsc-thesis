import CardWithIcon from "../components/CardWithIcon"

const DashboardView = () => {
  return (
    <div className="space-y-4 md:space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
            <CardWithIcon
                title="Nodes Online"
                value="4"
                hint="Click to view nodes"
                onClick={() => alert('Navigate to Nodes')}
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                icon={
                    <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                }
            />
            <CardWithIcon
                title="Nodes Offline"
                value="1"
                hint="Click to view nodes"
                onClick={() => alert('Navigate to Nodes')}
                textColorClass="text-slate-600"
                valueColorClass="text-red-600"
                icon={
                    <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                }
            />
            <CardWithIcon
                title="Pending Approvals"
                value="3"
                hint="Click to review approvals"
                onClick={() => alert('Navigate to Approvals')}
                textColorClass="text-slate-600"
                outlineColorClass="ring-2 ring-red-500"
                icon={
                    <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                }
            />
            <CardWithIcon
                title="Active Alerts"
                value="2"
                hint="Click to view system health"
                onClick={() => alert('Navigate to System Health')}
                textColorClass="text-purple-800"
                icon={
                    <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                }
            />
        </div>
    </div>
  )
}

export default DashboardView
