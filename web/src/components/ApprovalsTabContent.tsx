import type NodeRequest from "@/types/Node"
import type UserRequest from "@/types/User"
import Node from "./approvals/Node"
import Users from "./approvals/Users"

interface ApprovalsTabContentProps {
    selectedTab: string;
    selectedCategory: string;
    nodeApprovals?: NodeRequest[];
    userApprovals?: UserRequest[];
    onUserStatusChange?: () => void;
}


const ApprovalsTabContent = ({ selectedTab, selectedCategory, nodeApprovals, userApprovals, onUserStatusChange }: ApprovalsTabContentProps) => {
    return (
        <>
            {selectedCategory === 'vps' && (
            <>
                {nodeApprovals && nodeApprovals.length > 0 ? (
                <Node
                    approvals={nodeApprovals.filter(approval =>
                    selectedTab === 'pending' ? approval.status === 'pending' :
                    selectedTab === 'approved' ? approval.status === 'approved' :
                    approval.status === 'rejected'
                    )}
                />
                ) : (
                <p className="text-sm text-slate-600">No VPS node requests available.</p>
                )}
            </>
            )} {selectedCategory === 'users' && (
            <>
                {userApprovals && userApprovals.length > 0 ? (
                <Users
                    requests={userApprovals.filter(approval =>
                    selectedTab === 'pending' ? approval.status === 'pending' :
                    selectedTab === 'approved' ? approval.status === 'approved' :
                    selectedTab === 'rejected' ? approval.status === 'rejected' :
                    approval.status === 'rejected'
                    )}
                    onStatusChange={onUserStatusChange}
                />
                ) : (
                <p className="text-sm text-slate-600">No user registration requests available.</p>
                )}
            </>
            )}
        </>
    )
}

export default ApprovalsTabContent