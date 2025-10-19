import { useState } from 'react'
import { ChevronDown, User, Settings, LogOut } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';


const VIEW_TITLES: Record<string, string> = {
  dashboard: 'Dashboard',
  approvals: 'Approvals',
  nodes: 'Nodes',
  kubernetes: 'Kubernetes',
  networking: 'Networking',
  health: 'System Health',
  users: 'User Management',
  settings: 'Settings',
};



const Header = () => {
  const [showUserMenu, setShowUserMenu] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const handleLogout = async () => {

    try {
      const response = await fetch('/api/logout', {
        method: 'POST',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to log out. Please try again.');
      }
      
      logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed', error);
      const message = error instanceof Error ? error.message : 'Unexpected error during logout.';
      alert(message);
    }
  };

  return (
  <header className="bg-white border-b border-slate-200 px-6 py-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl text-slate-800">{VIEW_TITLES[location.pathname.split('/')[1] as keyof typeof VIEW_TITLES]}</h2>
        
        <div className="relative">
          <button
            onClick={() => setShowUserMenu(!showUserMenu)}
            className="flex items-center space-x-2 text-slate-600 hover:text-slate-800 transition-colors p-2 rounded-lg hover:bg-slate-100"
          >
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
              <User className="w-4 h-4 text-white" />
            </div>
            <div className="text-left">
              <p className="text-sm text-slate-800">Admin User</p>
              <p className="text-xs text-slate-600">admin@example.com</p>
            </div>
            <ChevronDown className={`w-4 h-4 transition-transform ${showUserMenu ? 'rotate-180' : ''}`} />
          </button>
          
          {showUserMenu && (
            <>
              <div 
                className="fixed inset-0 z-40" 
                onClick={() => setShowUserMenu(false)}
              />
              <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-slate-200 py-2 z-50">
                <button
                  onClick={() => {
                    setShowUserMenu(false);
                    alert('Profile settings coming soon!');
                  }}
                  className="w-full flex items-center space-x-2 px-4 py-2 text-slate-700 hover:bg-slate-100 transition-colors"
                >
                  <User className="w-4 h-4" />
                  <span>Profile</span>
                </button>
                <button
                  onClick={() => {
                    setShowUserMenu(false);
                  }}
                  className="w-full flex items-center space-x-2 px-4 py-2 text-slate-700 hover:bg-slate-100 transition-colors"
                >
                  <Settings className="w-4 h-4" />
                  <span>Settings</span>
                </button>
                <hr className="my-2 border-slate-200" />
                <button
                  onClick={() => {
                    setShowUserMenu(false);
                    handleLogout();
                  }}
                  className="w-full flex items-center space-x-2 px-4 py-2 text-red-600 hover:bg-red-50 transition-colors"
                >
                  <LogOut className="w-4 h-4" />
                  <span>Sign Out</span>
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  )
}

export default Header
