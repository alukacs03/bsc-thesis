import { RefreshCw } from "lucide-react"

interface TerminalProps {
    id?: string;
    title?: string;
    commandHistory?: string[];
    refreshButtonAction?: () => void;
    runCommandAction?: (command: string) => void;
}

const Terminal = ({ id, title, commandHistory, refreshButtonAction, runCommandAction }: TerminalProps) => {
  return (
    <div className="bg-slate-900 rounded-lg p-4">
        <div className="flex items-center justify-between mb-4">
            <h4 className="text-white text-sm">{title}</h4>
            {refreshButtonAction && (
                <button className="text-slate-400 hover:text-white" onClick={refreshButtonAction}>
                    <RefreshCw className="w-4 h-4" />
                </button>
            )}
        </div>
        <div id={id} className="space-y-1 max-h-64 overflow-y-auto">
            {commandHistory?.map((line, index) => (
            <div key={index} className={`text-sm ${
                line.startsWith('$') ? 'text-green-400' : 'text-slate-300'
            }`}>
                {line}
            </div>
            ))}
        </div>
        {runCommandAction && (
        <div className="flex items-center mt-4">
            <span className="text-green-400 text-sm mr-2">$</span>
            <input
            type="text"
            placeholder="Enter command..."
            className="flex-1 bg-transparent text-white text-sm outline-none"
            onKeyPress={(e) => {
                if (e.key === 'Enter') {
                runCommandAction?.(e.currentTarget.value);
                    e.currentTarget.value = '';
                }
            }}
            />
        </div>
        )}
    </div>
  )
}

export default Terminal
