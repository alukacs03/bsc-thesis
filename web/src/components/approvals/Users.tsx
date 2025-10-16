import type UserRequest from "../../types/User"
import { CheckCircle, User, XCircle } from "lucide-react";

interface UsersProps {
    requests: UserRequest[];
}

const Users = ({ requests }: UsersProps) => {
  return (
    <>
        <div className="space-y-4">
        {requests.map((user) => (
            <div key={user.id} className={`border border-slate-200 rounded-lg p-6 hover:border-blue-300 transition-colors ${user.status === 'approved' ? 'bg-green-50' : user.status === 'rejected' ? 'bg-red-50' : ''}`}>
                <div className="flex items-start justify-between">
                <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                    <User className="w-6 h-6 text-blue-600" />
                    <h4 className="text-lg text-slate-800">{user.fullName} ({user.username})</h4>
                    </div>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <p className="text-sm text-slate-600">Email</p>
                        <p className="text-sm text-slate-800">{user.email}</p>
                    </div>
                    <div>
                        <p className="text-sm text-slate-600">Requested at</p>
                        <p className="text-sm text-slate-800">{user.requestedAt}</p>
                    </div>
                    </div>
                    {user.approvedAt && (
                    <div className="mt-4">
                        <p className="text-sm text-green-800">Approved by {user.approvedBy} at {user.approvedAt}</p>
                    </div>
                    )}
                    {user.rejectedAt && (
                    <div className="mt-4">
                        <p className="text-sm text-red-800">Rejected by {user.rejectedBy} at {user.rejectedAt}</p>
                    </div>
                    )}
                    {user.reason && (
                    <div className="mt-4">
                        <p className="text-sm text-slate-600">Reason for Rejection:</p>
                        <p className="text-sm text-slate-800">{user.reason}</p>
                    </div>
                    )}
                </div>
                <div className="flex flex-col space-y-2 ml-6">
                    <button
                    onClick={() => handleApprove(user.id)}
                    className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
                    >
                    <CheckCircle className="w-4 h-4" />
                    <span>Approve</span>
                    </button>
                    <button
                    onClick={() => handleReject(user.id)}
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

export default Users
