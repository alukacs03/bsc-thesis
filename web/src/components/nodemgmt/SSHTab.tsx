import { Plus } from "lucide-react"
import { Key, Copy } from "lucide-react"
import React from "react"

interface SSHTabProps {
    sshKeys? : {
        user: string;
        name: string;
        type: string;
        bits: number;
        fingerprint: string;
        publicKey: string;
        created: string;
        lastUsed: string;
        status: string;
        id: number;
    }[]
}

const SSHTab = ({ sshKeys }: SSHTabProps) => {
  return (
    <div className="space-y-6">
        <div className="flex items-center justify-between">
            <h3 className="text-lg text-slate-800">SSH Key Management</h3>
            <button
                onClick={() => alert('Generate New Key functionality not implemented')}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
            >
                <Plus className="w-4 h-4" />
                <span>Generate New Key</span>
            </button>
        </div>

        <div className="space-y-4">
        {!sshKeys || sshKeys.length === 0 ? (
            <p className="text-sm text-slate-600">No SSH keys available.</p>
        ) : (
        sshKeys?.map((key) => (
            <div key={key.id} className="border border-slate-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
                <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                    <Key className="w-5 h-5 text-blue-600" />
                    <h4 className="text-lg text-slate-800">{key.name}</h4>
                    <span className={`px-2 py-1 rounded text-xs ${
                    key.status === 'active' 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-slate-100 text-slate-600'
                    }`}>
                    {key.status}
                    </span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-3">
                    <div>
                    <p className="text-sm text-slate-600">User</p>
                    <p className="text-sm text-slate-800">{key.user}</p>
                    </div>
                    <div>
                    <p className="text-sm text-slate-600">Type</p>
                    <p className="text-sm text-slate-800">{key.type} {key.bits} bits</p>
                    </div>
                    <div>
                    <p className="text-sm text-slate-600">Created</p>
                    <p className="text-sm text-slate-800">{key.created}</p>
                    </div>
                    <div>
                    <p className="text-sm text-slate-600">Last Used</p>
                    <p className="text-sm text-slate-800">{key.lastUsed}</p>
                    </div>
                </div>
                <div className="mb-3">
                    <p className="text-sm text-slate-600 mb-1">Fingerprint</p>
                    <code className="text-xs text-slate-700 bg-slate-100 px-2 py-1 rounded">{key.fingerprint}</code>
                </div>
                <div className="mb-3">
                    <p className="text-sm text-slate-600 mb-1">Public Key</p>
                    <div className="flex items-center space-x-2">
                    <code className="text-xs text-slate-700 bg-slate-100 px-2 py-1 rounded flex-1 truncate">
                        {key.publicKey}...
                    </code>
                    <button className="p-1 text-slate-600 hover:text-slate-800"
                        onClick={() => {
                            navigator.clipboard.writeText(key.publicKey);
                        }}>
                        <Copy className="w-4 h-4 hover:text-blue-600" />
                    </button>
                    </div>
                </div>
                </div>
                <div className="flex flex-col space-y-2 ml-4">
                <button className="px-3 py-1 text-blue-600 hover:bg-blue-50 border border-blue-600 rounded text-sm transition-colors"
                    onClick={() => alert('Edit functionality not implemented')}
                >
                    Edit
                </button>
                <button 
                    onClick={() => alert('Delete functionality not implemented')}
                    className="px-3 py-1 text-red-600 hover:bg-red-50 border border-red-600 rounded text-sm transition-colors"
                >
                    Delete
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

export default SSHTab
