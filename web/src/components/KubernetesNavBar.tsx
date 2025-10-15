interface KubernetesNavBarProps {
    setSelectedTab: (tab: "Overview" | "Workloads" | "Networking" | "Storage" | "Logs") => void;
    selectedTab: "Overview" | "Workloads" | "Networking" | "Storage" | "Logs";
}

const KubernetesNavBar = ({ setSelectedTab, selectedTab }: KubernetesNavBarProps) => {
  return (
    <nav className="flex space-x-8 px-6">
        {["Overview", "Workloads", "Networking", "Storage", "Logs"].map((tab) => (
            <button 
                onClick={() => setSelectedTab(tab as "Overview" | "Workloads" | "Networking" | "Storage")}
                className={`py-4 px-1 border-b-2 text-sm ${
                    selectedTab === tab
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}>
                    {tab}
                </button>
        ))}
    </nav>
)
}

export default KubernetesNavBar
