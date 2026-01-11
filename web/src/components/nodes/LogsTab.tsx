import React from "react";
import { useNodeLogs } from "@/services/hooks/useNodeLogs";

interface LogsTabProps {
    nodeId: number;
    enabled: boolean;
}

const LogsTab = ({ nodeId, enabled }: LogsTabProps) => {
  const { data: logsResponse, loading, error } = useNodeLogs(nodeId, {
    enabled,
    pollingInterval: 30000,
    limit: 200,
  });

  const logLines = React.useMemo(() => logsResponse?.logs ?? [], [logsResponse]);
  const scrollerRef = React.useRef<HTMLDivElement | null>(null);

  React.useEffect(() => {
    const el = scrollerRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [logLines]);

  return (
        <div className="space-y-3">
            <p className="text-xs text-slate-500">
                Showing only the last 2 minutes (from journald). Auto-updates every 30 seconds.
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
        </div>
    )
}

export default LogsTab
