import { Server, Users } from "lucide-react"
import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import ApprovalsTabContent from "../components/ApprovalsTabContent"
import React from "react"
import DetailsNavBar from "../components/DetailsNavBar"
import { useState, useEffect } from "react"
import type UserRequest from "../types/User"
import { useSearchParams } from "react-router-dom"
import { useEnrollments } from "@/services/hooks/useEnrollments"
import { LoadingSpinner } from "@/components/LoadingSpinner"
import { ErrorMessage } from "@/components/ErrorMessage"


const ApprovalsView = () => {
  const [selectedTab, setSelectedTab] = useState<'pending' | 'approved' | 'rejected'>('pending');
  const [userApprovals, setUserApprovals] = React.useState<UserRequest[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedCategory: "vps" | "users" =
    searchParams.get("tab") === "users" ? "users" : "vps";

  // Use the enrollment hook for VPS nodes
  const {
    data: enrollments,
    loading: enrollmentsLoading,
    error: enrollmentsError,
    refetch: refetchEnrollments
  } = useEnrollments();

  useEffect(() => {
    if (selectedCategory === 'users') {
      initUserApprovals();
    }
  }, [selectedCategory]);

  const initUserApprovals = () => {
    fetch('/api/admin/userRegRequests', { method: 'GET', credentials: 'include' })
      .then(res => res.json())
      .then((data: UserRequest[]) => {
        setUserApprovals(data);
      })
      .catch(err => {
        console.error('Failed to fetch user registration requests', err);
        setUserApprovals([]);
      });
  };

  const changeCategory = (category: "vps" | "users") => {
    const next = new URLSearchParams(searchParams);
    if (category === "users") next.set("tab", "users");
    else next.delete("tab");
    setSearchParams(next, { replace: true });
  }

  const nodeApprovals = enrollments || [];

  return (
    <>
        <div className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm">
                <div className="flex border-b border-slate-200">
                <button
                    onClick={() => changeCategory('vps')}
                    className={`flex-1 flex items-center justify-center space-x-3 px-6 py-4 transition-colors ${
                    selectedCategory === 'vps'
                        ? 'bg-blue-50 border-b-2 border-blue-500 text-blue-600'
                        : 'text-slate-600 hover:text-slate-700 hover:bg-slate-50'
                    }`}
                >
                    <Server className="w-5 h-5" />
                    <span>VPS Node Approvals</span>
                    {!enrollmentsLoading && (
                      <span className="bg-red-500 text-white text-xs px-2 py-1 rounded-full">
                        {nodeApprovals.filter(n => n.status === 'pending').length}
                      </span>
                    )}
                </button>
                <button
                    onClick={() => {
                        changeCategory('users');
                    }}
                    className={`flex-1 flex items-center justify-center space-x-3 px-6 py-4 transition-colors ${
                    selectedCategory === 'users'
                        ? 'bg-blue-50 border-b-2 border-blue-500 text-blue-600'
                        : 'text-slate-600 hover:text-slate-700 hover:bg-slate-50'
                    }`}
                >
                    <Users className="w-5 h-5" />
                    <span>User Registration Approvals</span>
                    <span className="bg-red-500 text-white text-xs px-2 py-1 rounded-full">
                      {userApprovals.filter(u => u.status === 'pending').length}
                    </span>
                </button>
                </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-1 gap-6">
            {(selectedCategory === 'vps' && (
                <CardWithIcon
                    icon={<Server className="w-6 h-6 text-blue-600" />}
                    title="Total VPS Node Requests"
                    value={enrollmentsLoading ? '...' : nodeApprovals.length}
                    noHover={true}
                />
            )) || (selectedCategory === 'users' && (
                  <CardWithIcon
                      icon={<Users className="w-6 h-6 text-blue-600" />}
                      title="Total User Registration Requests"
                      value={userApprovals.length}
                      noHover={true}
                  />
            ))}
            </div>

            {selectedCategory === 'vps' && enrollmentsLoading && <LoadingSpinner />}
            {selectedCategory === 'vps' && enrollmentsError && (
              <ErrorMessage
                message={enrollmentsError.message}
                onRetry={refetchEnrollments}
              />
            )}

            {(selectedCategory === 'users' || !enrollmentsLoading) && !enrollmentsError && (
              <CardContainer
                  title={selectedCategory === 'vps' ? 'VPS Node Approvals' : 'User Registration Approvals'}
                  noPadding={true}
              >
                  <DetailsNavBar
                      tabs={['pending', 'approved', 'rejected']}
                      setSelectedTab={(tab) => setSelectedTab(tab as 'pending' | 'approved' | 'rejected')}
                      selectedTab={selectedTab}
                  />
                  <div className="p-6">
                      <ApprovalsTabContent
                          selectedCategory={selectedCategory}
                          selectedTab={selectedTab}
                          nodeApprovals={nodeApprovals}
                          userApprovals={userApprovals}
                          onUserStatusChange={initUserApprovals}
                          onNodeStatusChange={refetchEnrollments}
                      />
                  </div>

              </CardContainer>
            )}

        </div>
    </>
  )
}

export default ApprovalsView
