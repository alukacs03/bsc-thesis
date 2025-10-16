import type NodeRequest from "../../types/Node"
import { AlertTriangle, CheckCircle, Cpu, HardDrive, MapPin, Network, Server, XCircle } from "lucide-react";
import { getPriorityColor } from "../../utils/Helpers";


interface NodeProps {
    approvals: NodeRequest[];

}

const Node = ({ approvals }: NodeProps ) => {
  return (
    <>
        <div className="space-y-4">
            {approvals.map((request) => (
            <div key={request.id} className={`border border-slate-200 rounded-lg p-6 hover:border-blue-300 transition-colors ${request.rejectedAt ? 'bg-red-50' : ''} ${request.approvedAt ? 'bg-green-50' : ''}`}>
                <div className="flex items-start justify-between">
                <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                    <Server className="w-6 h-6 text-blue-600" />
                    <h4 className="text-lg text-slate-800">{request.hostname}</h4>
                    <span className={`px-2 py-1 rounded text-xs border ${getPriorityColor(request.priority)}`}>
                        {request.priority}
                    </span>
                    {request.priority === 'urgent' && (
                        <div className="flex items-center space-x-1 text-red-600">
                        <AlertTriangle className="w-4 h-4" />
                        <span className="text-xs">URGENT</span>
                        </div>
                    )}
                    </div>
                    
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                    <div>
                        <p className="text-sm text-slate-600">IP Address</p>
                        <p className="text-sm text-slate-800 font-mono">{request.ipAddress}</p>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Location</p>
                        <div className="flex items-center space-x-1">
                        <MapPin className="w-3 h-3 text-slate-500" />
                        <p className="text-sm text-slate-800">{request.location}</p>
                        </div>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Provider</p>
                        <p className="text-sm text-slate-800">{request.provider}</p>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Instance Type</p>
                        <p className="text-sm text-slate-800">{request.instanceType}</p>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Requested by</p>
                        <p className="text-sm text-slate-800">{request.requestedBy}</p>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Requested at</p>
                        <p className="text-sm text-slate-800">{request.requestedAt}</p>
                    </div>
                    </div>
                    
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                        <div className="mb-4">
                        <p className="text-sm text-slate-600 mb-2">Purpose</p>
                        <p className="text-sm text-slate-800">{request.purpose}</p>
                        </div>

                        {(request.approvedAt) && (
                            <>
                                <div className="mb-4">
                                    <p className="text-sm text-slate-600 mb-2">Approved At</p>
                                    <p className="text-sm text-slate-800">{request.approvedAt}</p>
                                </div>
                                <div className="mb-4">
                                    <p className="text-sm text-slate-600 mb-2">Approved By</p>
                                    <p className="text-sm text-slate-800">{request.approvedBy}</p>
                                </div>
                            </>
                        )}

                        {(request.rejectedAt) && (
                            <>
                                <div className="mb-4">
                                    <p className="text-sm text-slate-600 mb-2">Rejected At</p>
                                    <p className="text-sm text-slate-800">{request.rejectedAt}</p>
                                </div>
                                <div className="mb-4">
                                    <p className="text-sm text-slate-600 mb-2">Rejected By</p>
                                    <p className="text-sm text-slate-800">{request.rejectedBy}</p>
                                </div>
                            </>
                        )}
                    </div>

                    {(request.reason) && (
                    <div className="mb-4">
                        <p className="text-sm text-slate-600 mb-2">Reason for Rejection</p>
                        <p className="text-sm text-slate-800">{request.reason}</p>
                    </div>
                    )}

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                    <div className="bg-slate-50 rounded-lg p-4">
                        <p className="text-sm text-slate-600 mb-2">Server Specifications</p>
                        <div className="space-y-2">
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center space-x-2">
                            <Cpu className="w-4 h-4 text-slate-500" />
                            <span className="text-slate-600">CPU:</span>
                            </div>
                            <span className="text-slate-800">{request.specs.cpu}</span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center space-x-2">
                            <Server className="w-4 h-4 text-slate-500" />
                            <span className="text-slate-600">Memory:</span>
                            </div>
                            <span className="text-slate-800">{request.specs.memory}</span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center space-x-2">
                            <HardDrive className="w-4 h-4 text-slate-500" />
                            <span className="text-slate-600">Storage:</span>
                            </div>
                            <span className="text-slate-800">{request.specs.storage}</span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                            <div className="flex items-center space-x-2">
                            <Network className="w-4 h-4 text-slate-500" />
                            <span className="text-slate-600">Bandwidth:</span>
                            </div>
                            <span className="text-slate-800">{request.specs.bandwidth}</span>
                        </div>
                        </div>
                    </div>
                    
                    <div className="bg-slate-50 rounded-lg p-4">
                        <p className="text-sm text-slate-600 mb-2">Security & Access</p>
                        <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                            <span className="text-slate-600">Security Group:</span>
                            <span className="text-slate-800">{request.securityGroup}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-slate-600">SSH Key:</span>
                            <span className="text-slate-800">{request.sshKey}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-slate-600">Est. Cost:</span>
                            <span className="text-slate-800">{request.estimatedCost}</span>
                        </div>
                        </div>
                    </div>
                    </div>
                </div>
                <div className="flex flex-col space-y-2 ml-6">
                    <button
                    onClick={() => handleApprove(request.id)}
                    className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
                    >
                    <CheckCircle className="w-4 h-4" />
                    <span>Approve</span>
                    </button>
                    <button
                    onClick={() => handleReject(request.id)}
                    className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
                    >
                    <XCircle className="w-4 h-4" />
                    <span>Reject</span>
                    </button>
                </div>
                </div>
            </div>
            ))}
        </div>
    </>
  )
}

export default Node
