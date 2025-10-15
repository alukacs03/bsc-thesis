import Terminal from "../Terminal";
import React from "react"

const LogsTab = () => {
    const [commandHistory, setCommandHistory] = React.useState<string[]>([]);

    const handleRunCommand = (command: string) => {
        // Placeholder for actual command execution logic
        if (command.trim() === 'clear') {
            setCommandHistory([]);
            return;
        }
        const newCommand = `$ ${command}`;
        const newOutput = `Executed command: ${command}`;
        setCommandHistory(prev => {
            const updated = [...prev, newCommand, newOutput];
            setTimeout(() => {
            const terminal = document.getElementById("terminal-scroll");
            if (terminal) terminal.scrollTop = terminal.scrollHeight;
            }, 0);
            return updated;
        });
    };

  return (
     <div className="space-y-6">
        <h3 className="text-lg text-slate-800">System Logs & Terminal</h3>
        
        <div className="bg-slate-900 rounded-lg p-4">
            <Terminal
                id="terminal-scroll"
                title="Node Terminal"
                commandHistory={commandHistory}
                refreshButtonAction={() => alert('Refresh functionality not implemented')}
                runCommandAction={handleRunCommand}
            />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
            <h4 className="text-md text-slate-800 mb-3">System Logs</h4>
            <div className="space-y-2">
            <button className="w-full text-left px-4 py-2 border border-slate-200 rounded hover:bg-slate-50 transition-colors">
                <div className="flex items-center justify-between">
                <span className="text-sm text-slate-800">System Log</span>
                <span className="text-xs text-slate-500">/var/log/syslog</span>
                </div>
            </button>
            <button className="w-full text-left px-4 py-2 border border-slate-200 rounded hover:bg-slate-50 transition-colors">
                <div className="flex items-center justify-between">
                <span className="text-sm text-slate-800">Auth Log</span>
                <span className="text-xs text-slate-500">/var/log/auth.log</span>
                </div>
            </button>
            <button className="w-full text-left px-4 py-2 border border-slate-200 rounded hover:bg-slate-50 transition-colors">
                <div className="flex items-center justify-between">
                <span className="text-sm text-slate-800">Gluon Agent</span>
                <span className="text-xs text-slate-500">/var/log/gluon-agent.log</span>
                </div>
            </button>
            </div>
        </div>

        <div>
            <h4 className="text-md text-slate-800 mb-3">Quick Actions</h4>
            <div className="space-y-2">
            <button
                onClick={() => handleRunCommand('systemctl restart gluon-agent')}
                className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm transition-colors"
            >
                Restart Gluon Agent
            </button>
            <button
                onClick={() => handleRunCommand('df -h')}
                className="w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded text-sm transition-colors"
            >
                Check Disk Space
            </button>
            <button
                onClick={() => handleRunCommand('free -h')}
                className="w-full px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded text-sm transition-colors"
            >
                Check Memory Usage
            </button>
            </div>
        </div>
        </div>
    </div>
  )
}

export default LogsTab
