import Logo from './Logo'
import { NavLink, useLocation } from 'react-router-dom'
import { useMemo } from 'react'
import { useEnrollments } from '@/services/hooks/useEnrollments'
import { useWireGuardPeers } from '@/services/hooks/useWireGuardPeers'
import { useOSPFNeighbors } from '@/services/hooks/useOSPFNeighbors'

interface NavItem {
  id: string;
  label: string;
  badge?: string | number;
  badgeColor?: string;
  path : string;
}

const NAV_ITEMS: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', path: '/dashboard' },
  { id: 'approvals', label: 'Approvals', path: '/approvals' },
  { id: 'nodes', label: 'Nodes', path: '/nodes' },
  { id: 'kubernetes', label: 'Kubernetes', path: '/kubernetes' },
  { id: 'networking', label: 'Networking', path: '/networking' },
];

function NavButton({ item, isActive }: { item: NavItem; isActive: boolean }) {
  return (
    <NavLink 
      to={item.path}
      className={`w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors ${
        isActive ? 'bg-blue-600 text-white' : 'text-slate-700 hover:bg-blue-50'
      }`}
    >
      {/* TODO: ICONS */}
      <span>{item.label}</span>
      {item.badge && (
        <span className={`ml-auto ${item.badgeColor || 'bg-blue-500'} text-white text-xs px-2 py-1 rounded-full`}>
          {item.badge}
        </span>
      )}
    </NavLink>
  );

}

const Sidebar = () => {
  const location = useLocation();
  const { data: enrollments } = useEnrollments({ pollingInterval: 30000 });
  const { data: wireGuardPeers } = useWireGuardPeers({ pollingInterval: 30000 });
  const { data: ospfNeighbors } = useOSPFNeighbors({ pollingInterval: 30000 });

  const pendingApprovals = useMemo(() => {
    return (enrollments ?? []).filter((request) => request.status === "pending").length;
  }, [enrollments]);

  const networkingIssues = useMemo(() => {
    const wgIssues = (wireGuardPeers ?? []).filter((peer) => peer.ui_status !== "connected").length;
    const ospfIssues = (ospfNeighbors ?? []).filter((neighbor) => {
      const state = (neighbor.state || "").toLowerCase();
      return !state.startsWith("full");
    }).length;
    return wgIssues + ospfIssues;
  }, [wireGuardPeers, ospfNeighbors]);

  const navItems = useMemo(() => {
    return NAV_ITEMS.map((item) => {
      if (item.id === "approvals" && pendingApprovals > 0) {
        return { ...item, badge: pendingApprovals, badgeColor: "bg-red-500" };
      }
      if (item.id === "networking" && networkingIssues > 0) {
        return { ...item, badge: networkingIssues, badgeColor: "bg-yellow-500" };
      }
      return item;
    });
  }, [pendingApprovals, networkingIssues]);

  return (
    <div className="w-56 bg-white border-r border-slate-200 flex flex-col sticky top-0 h-screen">
      <div className="p-6 border-b border-slate-200">
        <Logo />
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {navItems.map((item) => (
          <NavButton
            key={item.id}
            item={item}
            isActive={location.pathname.split('/')[1] === item.id}
          />
        ))} 
        <div className="border-t border-slate-200 my-4"></div>
      </nav>
      <div className="p-4 border-t border-slate-200">
        <button 
          onClick={() => {}}
          className="w-full flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-600 hover:text-slate-700 hover:bg-blue-50"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          <span>Sign Out</span>
        </button>
      </div>
    </div>
  )
}

export default Sidebar
