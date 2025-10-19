import type UserRequest from '@/types/User'
import React from "react";

interface ModalProps {
    isOpen: boolean;
    title: string;
    confirmLabel: string;
    cancelLabel: string;
    onConfirm: (reason?: string) => void;
    onCancel: () => void;
    type: 'approve' | 'reject';
    user?: UserRequest;
    reason?: string;
    onReasonChange?: (value: string) => void;
    children?: React.ReactNode;
}

const Modal = ({
    type,
    user,
    reason,
    isOpen,
    title,
    confirmLabel,
    cancelLabel,
    onConfirm,
    onCancel,
    onReasonChange,
    children
}: ModalProps) => {
    const handleConfirm = () => {
        if (type === 'reject') {
            onConfirm(reason);
        } else {
            onConfirm();
        }
    };

    return (
        <div className={`fixed inset-0 flex items-center justify-center z-50 ${isOpen ? '' : 'hidden'}`}>
            <div className="fixed inset-0 bg-black opacity-40"></div>
            <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-md mx-auto z-10">
                <h2 className="text-2xl font-semibold mb-4 text-gray-800">
                    {title}
                </h2>
                {user && (
                    <div className="mb-4">
                        <p className="text-gray-700">
                            <span className="font-medium">User:</span> {user.full_name}
                        </p>
                        <p className="text-gray-700">
                            <span className="font-medium">Email:</span> {user.email}
                        </p>
                    </div>
                )}
                {children && (
                    <div className="mb-4">
                        {children}
                    </div>
                )}
                {type === 'reject' ? (
                    <div className="mb-4">
                        <label className="block text-gray-700 font-medium mb-1">
                            Reason:
                        </label>
                        <textarea
                            value={reason ?? ''}
                            onChange={(e) => onReasonChange?.(e.target.value)}
                            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500"
                            placeholder="Provide a short reason"
                        />
                    </div>
                ) : reason !== undefined && (
                    <div className="mb-4">
                        <label className="block text-gray-700 font-medium mb-1">
                            Reason:
                        </label>
                        <input
                            type="text"
                            value={reason}
                            readOnly
                            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-700"
                        />
                    </div>
                )}
                <div className="flex justify-end gap-2 mt-6">
                    <button
                        className="px-4 py-2 rounded bg-gray-200 text-gray-800 hover:bg-gray-300"
                        onClick={onCancel}
                        type="button"
                    >
                        {cancelLabel}
                    </button>
                    <button
                        className={`px-4 py-2 rounded text-white ${type === 'approve' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700'}`}
                        onClick={handleConfirm}
                        type="button"
                    >
                        {confirmLabel}
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Modal
