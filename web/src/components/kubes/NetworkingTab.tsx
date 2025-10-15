interface NetworkingTabProps {
    servicesList? : { name: string; type: string; namespace: string; clusterIP: string; externalIP?: string; ports?: string; age?: string; cluster: string; }[];
}

const NetworkingTab = ({ servicesList } : NetworkingTabProps) => {
  return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <h4 className="text-lg text-slate-800">Services</h4>
                <button className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm transition-colors">
                Create Service
                </button>
            </div>
            <div className="overflow-x-auto">
                <table className="w-full">
                <thead>
                    <tr className="border-b border-slate-200">
                    <th className="text-left py-3 px-4 text-slate-700">Name</th>
                    <th className="text-left py-3 px-4 text-slate-700">Type</th>
                    <th className="text-left py-3 px-4 text-slate-700">Cluster IP</th>
                    <th className="text-left py-3 px-4 text-slate-700">External IP</th>
                    <th className="text-left py-3 px-4 text-slate-700">Ports</th>
                    <th className="text-left py-3 px-4 text-slate-700">Age</th>
                    </tr>
                </thead>
                <tbody>
                    {!servicesList || servicesList.length === 0 ? (
                        <tr>
                            <td colSpan={6} className="py-3 px-4 text-slate-500 text-center">No services</td>
                        </tr>
                    ) : (
                        servicesList.map((service, index) => (
                            <tr key={index} className="border-b border-slate-100 hover:bg-blue-50">
                                <td className="py-3 px-4 text-slate-700">{service.name ? service.name : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{service.type ? service.type : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{service.clusterIP ? service.clusterIP : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{service.externalIP ? service.externalIP : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{service.ports ? service.ports : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{service.age ? service.age : '-'}</td>
                            </tr>
                        ))
                    )}
                </tbody>
                </table>
            </div>
        </div>
  )
}

export default NetworkingTab
