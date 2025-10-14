import Header from "../components/Header"
import Sidebar from "../components/Sidebar"
import { Outlet } from "react-router-dom"

const AppLayout = () => {
  return (
    <div className="min-h-screen bg-slate-100 flex">
      <Sidebar  />
      <div className="flex-1 flex flex-col">
            <Header />
            <main className="flex-1 p-6">
                <Outlet />
            </main>

      </div>

    </div>

  )
}

export default AppLayout
