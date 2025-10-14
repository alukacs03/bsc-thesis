import AppLayout from './layouts/AppLayout'
import { useState } from 'react';
import DashboardView from './views/DashboardView';

function App() {
    const [currentView, setCurrentView] = useState<string>('dashboard');
    
    const renderPage = () => {
        switch (currentView) {
            case 'dashboard':
                return <DashboardView />;
            default:
                return <div>Not Found</div>;
        }
    };

    const handleNavigate = (view: string) => {
        setCurrentView(view);
    };

    const handleLogout = () => {
        alert('Logging out...');
    };

  return (
    <>
      <AppLayout
        currentView={currentView === 'node-management' ? 'nodes' : currentView}
        onNavigate={handleNavigate}
        onLogout={handleLogout}
      >
        {renderPage()}
      </AppLayout>
    </>
  )
}

export default App
