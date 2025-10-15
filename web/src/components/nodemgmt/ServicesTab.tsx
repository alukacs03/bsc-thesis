interface ServicesTabProps {
   services?: {
         name: string;
         status: string;
         enabled: boolean;
         description: string;
     }[]
}


const ServicesTab = ({ services }: ServicesTabProps) => {
  return (
    <div className="space-y-6">
        <h3 className="text-lg text-slate-800">System Services</h3>
        <div className="space-y-3">
        {!services || services.length === 0 ? (
            <p className="text-sm text-slate-600">No services available.</p>
        ) : (
        services?.map((service) => (
            <div key={service.name} className="flex items-center justify-between p-4 border border-slate-200 rounded-lg">
            <div className="flex items-center space-x-4">
                <div className={`w-3 h-3 rounded-full ${
                service.status === 'running' ? 'bg-green-500' : 'bg-red-500'
                }`}></div>
                <div>
                <h4 className="text-slate-800">{service.name}</h4>
                <p className="text-sm text-slate-600">{service.description}</p>
                </div>
            </div>
            <div className="flex items-center space-x-3">
                <span className={`px-2 py-1 rounded text-xs ${
                service.status === 'running' 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800'
                }`}>
                {service.status}
                </span>
                <span className={`px-2 py-1 rounded text-xs ${
                service.enabled 
                    ? 'bg-blue-100 text-blue-800' 
                    : 'bg-slate-100 text-slate-600'
                }`}>
                {service.enabled ? 'enabled' : 'disabled'}
                </span>
                <div className="flex space-x-1">
                <button 
                    onClick={() => alert(`Toggling enable/disable for service ${service.name} functionality not implemented`)}
                    className="px-2 py-1 text-purple-600 hover:bg-purple-50 border border-purple-600 rounded text-xs transition-colors"
                >
                    {service.enabled ? 'Disable' : 'Enable'}
                </button>
                {service.status === 'running' ? (
                    <button
                    onClick={() => alert(`Stopping service ${service.name} functionality not implemented`)}
                    className="px-2 py-1 text-red-600 hover:bg-red-50 border border-red-600 rounded text-xs transition-colors"
                    >
                    Stop
                    </button>
                ) : (
                    <button
                    onClick={() => alert(`Starting service ${service.name} functionality not implemented`)}
                    className="px-2 py-1 text-green-600 hover:bg-green-50 border border-green-600 rounded text-xs transition-colors"
                    >
                    Start
                    </button>
                )}
                <button
                    onClick={() => alert(`Restarting service ${service.name} functionality not implemented`)}
                    className="px-2 py-1 text-blue-600 hover:bg-blue-50 border border-blue-600 rounded text-xs transition-colors"
                >
                    Restart
                </button>
                </div>
            </div>
            </div>
        )))
    }
        </div>
    </div>
  )
}

export default ServicesTab
