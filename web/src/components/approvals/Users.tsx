import type UserRequest from "../../types/User"
import { CheckCircle, User, XCircle } from "lucide-react";
import React from "react";
import Modal from "./Modal";

interface UsersProps {
    requests: UserRequest[];
    onStatusChange?: () => void;
}

const Users = ({ requests, onStatusChange }: UsersProps) => {
    const [activeModal, setActiveModal] = React.useState<'approve' | 'reject' | null>(null);
    const [activeUser, setActiveUser] = React.useState<UserRequest | null>(null);
    const [rejectReason, setRejectReason] = React.useState('');
    const [validationError, setValidationError] = React.useState<string | null>(null);
    console.log(requests)

    const closeModal = () => {
        setActiveModal(null);
        setActiveUser(null);
        setRejectReason('');
        setValidationError(null);
    };

    const openApproveModal = (user: UserRequest) => {
        setActiveUser(user);
        setActiveModal('approve');
        setValidationError(null);
    };

    const openRejectModal = (user: UserRequest) => {
        setActiveUser(user);
        setRejectReason('');
        setActiveModal('reject');
        setValidationError(null);
    };

    const confirmApprove = async () => {
        if (!activeUser) {
            return;
        }

        const response = await fetch(`/api/modifyRegistrationRequest`, {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                request_id: String(activeUser.id),
                status: 'approved',
            }),
        });

        if (!response.ok) {
            throw new Error('Failed to approve user request');
        }

        onStatusChange?.();
        closeModal();
    };

    const confirmReject = async (providedReason?: string) => {
        if (!activeUser) {
            return;
        }

        const finalReason = (providedReason ?? rejectReason).trim();

        if (!finalReason) {
            setValidationError("Reason is required");
            return;
        }

        const response = await fetch("/api/modifyRegistrationRequest", {
            method: "POST",
            credentials: "include",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                request_id: String(activeUser.id),
                status: "rejected",
                rejection_reason: finalReason,
            }),
        });

        if (!response.ok) {
            throw new Error("Failed to reject user request");
        }

        onStatusChange?.();
        closeModal();
    };

    function formatDate(dateString: string) {
        const options: Intl.DateTimeFormatOptions = {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        };
      
        const date = new Date(dateString);
        return date.toLocaleDateString(undefined, options);
      }

    return (
        <>
            <div className="space-y-4">
            {requests.map((user) => (
                <div key={user.id} className={`border border-slate-200 rounded-lg p-6 hover:border-blue-300 transition-colors ${user.status === 'approved' ? 'bg-green-50' : user.status === 'rejected' ? 'bg-red-50' : ''}`}>
                    <div className="flex items-start justify-between">
                    <div className="flex-1">
                        <div className="flex items-center space-x-3 mb-4">
                        <User className="w-6 h-6 text-blue-600" />
                        <h4 className="text-lg text-slate-800">{user.full_name}</h4>
                        </div>
                        
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <p className="text-sm text-slate-600">Email</p>
                            <p className="text-sm text-slate-800">{user.email}</p>
                        </div>
                        <div>
                            <p className="text-sm text-slate-600">Requested at</p>
                            <p className="text-sm text-slate-800">{formatDate(user.requested_at)}</p>
                        </div>
                        </div>
                        {user.approved_at && (
                        <div className="mt-4">
                            <p className="text-sm text-green-800">Approved by {user.approved_by?.name} ({user.approved_by?.email}) on {formatDate(user.approved_at)}</p>
                        </div>
                        )}
                        {user.rejected_at && (
                        <div className="mt-4">
                            <p className="text-sm text-red-800">Rejected by {user.rejected_by?.name} ({user.rejected_by?.email}) on {formatDate(user.rejected_at)}</p>
                        </div>
                        )}
                        {user.rejection_reason && (
                        <div className="mt-4">
                            <p className="text-sm text-slate-600">Reason for rejection:</p>
                            <p className="text-sm text-slate-800">{user.rejection_reason}</p>
                        </div>
                        )}
                    </div>
                    <div className="flex flex-col space-y-2 ml-6">
                        <button
                        onClick={() => openApproveModal(user)}
                        className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
                        >
                        <CheckCircle className="w-4 h-4" />
                        <span>Approve</span>
                        </button>
                        <button
                        onClick={() => openRejectModal(user)}
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
        {activeModal === 'approve' && activeUser && (
            <Modal
                type="approve"
                user={activeUser}
                isOpen
                title="Approve User"
                confirmLabel="Approve"
                cancelLabel="Cancel"
                onConfirm={confirmApprove}
                onCancel={closeModal}
            >
                <p>Are you sure you want to approve the registration request for <strong>{activeUser.full_name}</strong>?</p>
            </Modal>
        )}
        {activeModal === 'reject' && activeUser && (
            <Modal
                type="reject"
                user={activeUser}
                isOpen
                title="Reject User"
                confirmLabel="Reject"
                cancelLabel="Cancel"
                reason={rejectReason}
                onConfirm={confirmReject}
                onReasonChange={(value) => {
                    setRejectReason(value);
                    if (validationError) {
                        setValidationError(null);
                    }
                }}
                onCancel={closeModal}
            >
                <p>Are you sure you want to reject the registration request for <strong>{activeUser.full_name}</strong>?</p>
                {validationError && (
                    <p className="text-sm text-red-600 mt-2">{validationError}</p>
                )}
            </Modal>
        )}
        </>
      )
}

export default Users;
