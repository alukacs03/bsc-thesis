import React from "react"
import { useNodeLogs } from "@/services/hooks/useNodeLogs";

const LogsTab = ({ nodeId, enabled }: { nodeId: number; enabled: boolean }) => {
    const { data: logsResponse, loading, error } = useNodeLogs(nodeId, {
      enabled,
      pollingInterval: 30000,
      limit: 200,
    });

    const logLines = React.useMemo(() => {
      return logsResponse?.logs ?? [];
    }, [logsResponse]);

    const scrollerRef = React.useRef<HTMLDivElement | null>(null);

    React.useEffect(() => {
        const el = scrollerRef.current;
        if (!el) return;
        el.scrollTop = el.scrollHeight;
    }, [logLines]);

  return (
     <div className="space-y-6">
        <h3 className="text-lg text-slate-800">Logs</h3>

        <div className="bg-white border border-slate-200 rounded-lg p-4">
            <h4 className="text-md text-slate-800 mb-1">Node Logs</h4>
            <p className="text-xs text-slate-500 mb-3">
                Showing only the last 2 minutes (from journald).
            </p>
            {enabled === false ? (
                <p className="text-sm text-slate-600">Open the Logs tab to start live polling.</p>
            ) : error ? (
                <p className="text-sm text-red-600">{error.message}</p>
            ) : loading && !logsResponse ? (
                <p className="text-sm text-slate-600">Loading logs…</p>
            ) : logLines.length === 0 ? (
                <p className="text-sm text-slate-600">No logs yet.</p>
            ) : (
                <div
                    ref={scrollerRef}
                    className="bg-slate-900 text-slate-100 rounded-lg p-3 font-mono text-xs max-h-64 overflow-auto"
                >
                    {logLines.map((line, idx) => (
                        <div key={idx} className="whitespace-pre-wrap break-words">
                            {line}
                        </div>
                    ))}
                </div>
            )}
            <p className="text-xs text-slate-500 mt-2">Auto-updates every 30 seconds.</p>
        </div>
    </div>
  )
}

export default LogsTab
