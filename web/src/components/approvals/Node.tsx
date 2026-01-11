import { useState } from "react";
import type { NodeEnrollmentRequest } from "@/services/types/enrollment"
import { CheckCircle, Server, XCircle } from "lucide-react";
import { enrollmentsAPI } from "@/services/api/enrollments";
import { handleAPIError } from "@/utils/errorHandler";
import { toast } from "sonner";
import APIKeyModal from "./APIKeyModal";

interface NodeProps {
    approvals: NodeEnrollmentRequest[];
    onRefetch?: () => void;
}

const Node = ({ approvals, onRefetch }: NodeProps) => {
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [apiKeyNodeName, setApiKeyNodeName] = useState<string>("");
  const [showAPIKeyModal, setShowAPIKeyModal] = useState(false);
  const [rejectingId, setRejectingId] = useState<number | null>(null);
  const [rejectReason, setRejectReason] = useState("");
  const [processingId, setProcessingId] = useState<number | null>(null);

  const handleApprove = async (id: number, hostname: string) => {
    if (processingId) return; // Prevent double clicks

    setProcessingId(id);
    try {
      // First approve the enrollment
      await enrollmentsAPI.approve(id);

      // Then generate API key - need to get the converted node ID
      // Since we don't have it yet, we'll use the enrollment ID
      const response = await enrollmentsAPI.generateAPIKey(id);

      // Show the API key to user
      setApiKey(response.api_key);
      setApiKeyNodeName(hostname);
      setShowAPIKeyModal(true);

      toast.success('Node enrollment approved successfully');

      // Trigger parent refetch
      onRefetch?.();
    } catch (error) {
      const message = handleAPIError(error, 'approve enrollment');
      toast.error(message);
    } finally {
      setProcessingId(null);
    }
  };

  const handleReject = async (id: number) => {
    if (!rejectReason.trim()) {
      toast.error('Please provide a reason for rejection');
      return;
    }

    if (processingId) return;

    setProcessingId(id);
    try {
      await enrollmentsAPI.reject(id, { reason: rejectReason });
      toast.success('Node enrollment rejected');
      setRejectingId(null);
      setRejectReason("");
      onRefetch?.();
    } catch (error) {
      const message = handleAPIError(error, 'reject enrollment');
      toast.error(message);
    } finally {
      setProcessingId(null);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  return (
    <>
        <div className="space-y-4">
            {approvals.map((request) => (
            <div
              key={request.id}
              className={`border border-slate-200 rounded-lg p-6 hover:border-blue-300 transition-colors ${
                request.status === 'rejected' ? 'bg-red-50' : ''
              } ${(request.status === 'approved' || request.status === 'accepted') ? 'bg-green-50' : ''}`}
            >
                <div className="flex items-start justify-between">
                <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                      <Server className="w-6 h-6 text-blue-600" />
                      <h4 className="text-lg font-semibold text-slate-800">{request.hostname}</h4>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${
                        request.desired_role === 'hub'
                          ? 'bg-purple-100 text-purple-700 border border-purple-300'
                          : 'bg-blue-100 text-blue-700 border border-blue-300'
                      }`}>
                        {request.desired_role}
                      </span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                      <div>
                        <p className="text-sm text-slate-600">Public IP</p>
                        <p className="text-sm text-slate-800 font-mono">{request.public_ip}</p>
                      </div>
                      <div>
                        <p className="text-sm text-slate-600">Provider</p>
                        <p className="text-sm text-slate-800">{request.provider}</p>
                      </div>
                      <div>
                        <p className="text-sm text-slate-600">Operating System</p>
                        <p className="text-sm text-slate-800">{request.os}</p>
                      </div>
                      <div>
                        <p className="text-sm text-slate-600">Requested At</p>
                        <p className="text-sm text-slate-800">{formatDate(request.requested_at)}</p>
                      </div>
                      {request.approved_at && (
                        <>
                          <div>
                            <p className="text-sm text-slate-600">Approved At</p>
                            <p className="text-sm text-slate-800">{formatDate(request.approved_at)}</p>
                          </div>
                          <div>
                            <p className="text-sm text-slate-600">Approved By</p>
                            <p className="text-sm text-slate-800">{request.approved_by?.email || 'N/A'}</p>
                          </div>
                        </>
                      )}
                      {request.rejected_at && (
                        <>
                          <div>
                            <p className="text-sm text-slate-600">Rejected At</p>
                            <p className="text-sm text-slate-800">{formatDate(request.rejected_at)}</p>
                          </div>
                          <div>
                            <p className="text-sm text-slate-600">Rejected By</p>
                            <p className="text-sm text-slate-800">{request.rejected_by?.email || 'N/A'}</p>
                          </div>
                        </>
                      )}
                    </div>

                    {request.rejection_reason && (
                      <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-3">
                        <p className="text-sm text-slate-600 mb-1 font-medium">Reason for Rejection</p>
                        <p className="text-sm text-red-800">{request.rejection_reason}</p>
                      </div>
                    )}

                    {request.status === 'pending' && rejectingId === request.id && (
                      <div className="mb-4 p-4 border border-slate-200 rounded-lg bg-slate-50">
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                          Rejection Reason
                        </label>
                        <textarea
                          value={rejectReason}
                          onChange={(e) => setRejectReason(e.target.value)}
                          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                          rows={3}
                          placeholder="Provide a reason for rejecting this enrollment request..."
                        />
                        <div className="flex space-x-2 mt-2">
                          <button
                            onClick={() => handleReject(request.id)}
                            disabled={processingId === request.id}
                            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm transition-colors disabled:opacity-50"
                          >
                            {processingId === request.id ? 'Processing...' : 'Confirm Rejection'}
                          </button>
                          <button
                            onClick={() => {
                              setRejectingId(null);
                              setRejectReason("");
                            }}
                            className="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 rounded-lg text-sm transition-colors"
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    )}
                </div>

                {request.status === 'pending' && (
                  <div className="flex flex-col space-y-2 ml-6">
                    <button
                      onClick={() => handleApprove(request.id, request.hostname)}
                      disabled={processingId === request.id}
                      className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2 disabled:opacity-50"
                    >
                      <CheckCircle className="w-4 h-4" />
                      <span>{processingId === request.id ? 'Processing...' : 'Approve'}</span>
                    </button>
                    <button
                      onClick={() => setRejectingId(request.id)}
                      disabled={processingId === request.id || rejectingId === request.id}
                      className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2 disabled:opacity-50"
                    >
                      <XCircle className="w-4 h-4" />
                      <span>Reject</span>
                    </button>
                  </div>
                )}
                </div>
            </div>
            ))}
        </div>

        {showAPIKeyModal && apiKey && (
          <APIKeyModal
            apiKey={apiKey}
            nodeName={apiKeyNodeName}
            onClose={() => {
              setShowAPIKeyModal(false);
              setApiKey(null);
              setApiKeyNodeName("");
            }}
          />
        )}
    </>
  )
}

export default Node
