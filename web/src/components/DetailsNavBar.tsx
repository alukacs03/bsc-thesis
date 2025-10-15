interface DetailsNavBarProps {
    tabs: string[];
    icons? : React.ReactNode[];
    setSelectedTab: (tab: string) => void;
    selectedTab: string;
}

const DetailsNavBar = ({ tabs, setSelectedTab, selectedTab, icons }: DetailsNavBarProps) => {
  return (
    <nav className="flex space-x-8 px-6">
        {tabs.map((tab, index) => (
            <div key={tab} className={`flex items-center border-b-2 ${selectedTab === tab ? 'border-blue-500 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}`}>
            {icons && icons[index]}
            <button 
                key={tab}
                onClick={() => setSelectedTab(tab)}
                className={`py-4 px-1  text-sm`}>
                    {tab}
                </button>
            </div>
        ))}
    </nav>
    )
}

export default DetailsNavBar
