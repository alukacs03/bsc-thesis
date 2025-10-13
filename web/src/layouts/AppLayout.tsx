import Header from "../components/Header"
import Sidebar from "../components/Sidebar"
import type { ReactNode } from "react"

interface AppLayoutProps {
    children: ReactNode
}

const AppLayout = ({ children }: AppLayoutProps) => {
  return (
    <div className="min-h-screen bg-slate-100 flex">
      <Sidebar currentView="dashboard" onNavigate={() => {}} onLogout={() => {}} />
      <div className="flex-1 flex flex-col">
            <Header currentView="dashboard" onNavigate={() => {}} onLogout={() => {}} />
            <main className="flex-1 p-6">
                {children}
            </main>

      </div>

    </div>

  )
}

export default AppLayout
