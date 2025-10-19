import Header from "../components/Header"
import Sidebar from "../components/Sidebar"
import { useOutlet } from "react-router-dom"

const AppLayout = () => {
  const outlet = useOutlet();

  return (
    <div className="min-h-screen bg-slate-100 flex">
      <Sidebar  />
      <div className="flex-1 flex flex-col">
            <Header />
            <main className="flex-1 p-6">
                {outlet}
            </main>
      </div>
    </div>
  )
}

export default AppLayout;
