import Logo from './Logo'

interface SidebarProps {
  currentView: string;
  onNavigate: (view: string) => void;
  onLogout: () => void;
}

interface NavItem {
  id: string;
  label: string;
  badge?: string | number;
  badgeColor?: string;
}

// badge is for demonstration purposes currently
const NAV_ITEMS: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'approvals', label: 'Approvals', badge: 3, badgeColor: 'bg-red-500' },
  { id: 'nodes', label: 'Nodes' },
  { id: 'kubernetes', label: 'Kubernetes' },
  { id: 'networking', label: 'Networking', badge: 2, badgeColor: 'bg-yellow-500' },
];

function NavButton({ item, isActive, onClick }: { item: NavItem; isActive: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
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
    </button>
  );

}

const Sidebar = ({ currentView, onNavigate, onLogout }: SidebarProps) => {
  return (
    <div className="w-56 bg-white border-r border-slate-200 flex flex-col">
      <div className="p-6 border-b border-slate-200">
        <Logo />
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {NAV_ITEMS.map((item) => (
          <NavButton 
            key={item.id}
            item={item}
            isActive={currentView === item.id}
            onClick={() => onNavigate(item.id)}
          />
        ))} 
        <div className="border-t border-slate-200 my-4"></div>
      </nav>
      <div className="p-4 border-t border-slate-200">
        <button 
          onClick={onLogout}
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
