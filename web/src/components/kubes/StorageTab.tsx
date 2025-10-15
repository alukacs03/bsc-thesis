import { HardDrive } from "lucide-react";

interface StorageTabProps {
    storageList? : { name: string; type: string; namespace: string; capacity: string; accessModes: string; age: string; cluster: string; }[];
}

const StorageTab = ({ storageList }: StorageTabProps) => {
  return (
    (storageList && storageList.length > 0) ? (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <h4 className="text-lg text-slate-800">Storage Resources</h4>
                <button className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm transition-colors"
                    onClick={() => alert('Create Storage clicked (would open create storage form)')}
                >
                Create Storage
                </button>
            </div>
            <div className="overflow-x-auto">
                <table className="w-full">
                <thead>
                    <tr className="border-b border-slate-200">
                    <th className="text-left py-3 px-4 text-slate-700">Name</th>
                    <th className="text-left py-3 px-4 text-slate-700">Type</th>
                    <th className="text-left py-3 px-4 text-slate-700">Namespace</th>
                    <th className="text-left py-3 px-4 text-slate-700">Capacity</th>
                    <th className="text-left py-3 px-4 text-slate-700">Access Modes</th>
                    <th className="text-left py-3 px-4 text-slate-700">Age</th>
                    </tr>
                </thead>
                <tbody>
                    {!storageList || storageList.length === 0 ? (
                        <tr>
                            <td colSpan={6} className="py-3 px-4 text-slate-500 text-center">No storage resources</td>
                        </tr>
                    ) : (
                        storageList.map((storage, index) => (
                            <tr key={index} className="border-b border-slate-100 hover:bg-blue-50">
                                <td className="py-3 px-4 text-slate-700">{storage.name ? storage.name : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{storage.type ? storage.type : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{storage.namespace ? storage.namespace : '-'}</td>
                                <td className="py-3 px-4 text!slate-600">{storage.capacity ? storage.capacity : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{storage.accessModes ? storage.accessModes : '-'}</td>
                                <td className="py-3 px-4 text-slate-600">{storage.age ? storage.age : '-'}</td>
                            </tr>
                        ))
                    )}
                </tbody>
                </table>
            </div>
        </div>
    ) : (
    <div className="text-center py-8">
    <div className="w-16 h-16 bg-slate-100 rounded-lg mx-auto mb-4 flex items-center justify-center">
        <HardDrive className="w-8 h-8 text-slate-500" />
    </div>
    <h4 className="text-lg text-slate-800 mb-2">Storage Management</h4>
    <p className="text-slate-600">Persistent volumes, storage classes, and volume claims would be displayed here.</p>
    </div>
  ))
}

export default StorageTab
