import type UserRequest from "@/types/User"
import type { NodeEnrollmentRequest } from "@/services/types/enrollment"
import Node from "./approvals/Node"
import Users from "./approvals/Users"

interface ApprovalsTabContentProps {
    selectedTab: string;
    selectedCategory: string;
    nodeApprovals?: NodeEnrollmentRequest[];
    userApprovals?: UserRequest[];
    onUserStatusChange?: () => void;
    onNodeStatusChange?: () => void;
}


const ApprovalsTabContent = ({ selectedTab, selectedCategory, nodeApprovals, userApprovals, onUserStatusChange, onNodeStatusChange }: ApprovalsTabContentProps) => {
    const vpsApprovals = nodeApprovals ?? []
    const userRequests = userApprovals ?? []

    const vpsFiltered = vpsApprovals.filter(approval =>
        selectedTab === 'pending' ? approval.status === 'pending' :
        selectedTab === 'approved' ? (approval.status === 'approved' || approval.status === 'accepted') :
        approval.status === 'rejected'
    )

    const usersFiltered = userRequests.filter(approval =>
        selectedTab === 'pending' ? approval.status === 'pending' :
        selectedTab === 'approved' ? approval.status === 'approved' :
        selectedTab === 'rejected' ? approval.status === 'rejected' :
        approval.status === 'rejected'
    )

    const vpsEmptyText = selectedTab === 'pending'
        ? 'No pending VPS node requests.'
        : selectedTab === 'approved'
            ? 'No approved VPS node requests.'
            : 'No rejected VPS node requests.'

    const usersEmptyText = selectedTab === 'pending'
        ? 'No pending user registration requests.'
        : selectedTab === 'approved'
            ? 'No approved user registration requests.'
            : 'No rejected user registration requests.'

    return (
        <>
            {selectedCategory === 'vps' && (
            <>
                {vpsFiltered.length > 0 ? (
                    <Node
                        approvals={vpsFiltered}
                        onRefetch={onNodeStatusChange}
                    />
                ) : (
                    <p className="text-sm text-slate-600">{vpsEmptyText}</p>
                )}
            </>
            )} {selectedCategory === 'users' && (
            <>
                {usersFiltered.length > 0 ? (
                    <Users
                        requests={usersFiltered}
                        onStatusChange={onUserStatusChange}
                    />
                ) : (
                    <p className="text-sm text-slate-600">{usersEmptyText}</p>
                )}
            </>
            )}
        </>
    )
}

export default ApprovalsTabContent
