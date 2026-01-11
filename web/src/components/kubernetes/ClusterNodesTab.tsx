import Table from "@/components/Table";
import type { Node } from "@/services/types/node";
import { clampPercent, formatBytes, formatPercent, formatUptimeSeconds } from "@/utils/format";
import { getStatusColor } from "@/utils/Helpers";
import { useNavigate } from "react-router-dom";

export default function ClusterNodesTab({ nodes }: { nodes: Node[] }) {
  const navigate = useNavigate();

  function formatLastSeen(timestamp?: string): string {
    if (!timestamp) return "Never";
    const diff = Date.now() - new Date(timestamp).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h4 className="text-lg text-slate-800">Cluster Nodes</h4>
        <p className="text-xs text-slate-500">Shows Gluon nodes + Kubernetes automation status.</p>
      </div>

      <Table columns={["Hostname", "K8s", "Gluon Role", "Status", "IP", "CPU", "MEM", "Disk", "Uptime", "Last Seen", "Actions"]}>
        {nodes.length === 0 ? (
          <tr>
            <td colSpan={11} className="py-6 px-4 text-sm text-slate-600">
              No nodes found.
            </td>
          </tr>
        ) : (
          nodes.map((n) => {
            const cpu = clampPercent(n.cpu_usage ?? Number.NaN);
            const mem = clampPercent(n.memory_usage ?? Number.NaN);
            const disk = clampPercent(n.disk_usage ?? Number.NaN);
            const wantsCP = n.role === "hub" || n.reported_desired_role === "hub";
            const k8sLabel = wantsCP ? "control-plane" : "worker";
            const k8sState = (n.k8s_state ?? "unknown").toString();
            const k8sStateColor =
              k8sState === "joined_control_plane" || k8sState === "joined_worker" || k8sState === "cluster_initialized"
                ? "bg-green-100 text-green-800"
                : k8sState === "error"
                  ? "bg-red-100 text-red-800"
                  : "bg-slate-100 text-slate-700";
            return (
              <tr key={n.id} className="border-b border-slate-100 hover:bg-blue-50">
                <td className="py-3 px-4 text-slate-800">{n.hostname}</td>
                <td className="py-3 px-4">
                  <div className="space-y-1">
                    <div className="text-slate-800 text-sm">{k8sLabel}</div>
                    <div className={`inline-flex px-2 py-0.5 rounded text-xs ${k8sStateColor}`} title={n.k8s_last_error ?? undefined}>
                      {k8sState}
                    </div>
                  </div>
                </td>
                <td className="py-3 px-4 text-slate-600">{n.role}</td>
                <td className="py-3 px-4">
                  <span className={`px-2 py-1 rounded text-xs ${getStatusColor(n.status === "active" ? "online" : n.status)}`}>
                    {n.status}
                  </span>
                </td>
                <td className="py-3 px-4 text-slate-600 font-mono text-sm">{n.public_ip}</td>
                <td className="py-3 px-4 text-slate-600">{formatPercent(cpu)}</td>
                <td className="py-3 px-4 text-slate-600">{formatPercent(mem)}</td>
                <td className="py-3 px-4 text-slate-600">
                  {formatPercent(disk)} ({formatBytes(n.disk_used_bytes)} / {formatBytes(n.disk_total_bytes)})
                </td>
                <td className="py-3 px-4 text-slate-600">{formatUptimeSeconds(n.uptime_seconds)}</td>
                <td className="py-3 px-4 text-slate-600">{formatLastSeen(n.last_seen_at)}</td>
                <td className="py-3 px-4">
                  <button
                    className="text-blue-600 hover:text-blue-800 text-sm"
                    onClick={() => navigate(`/nodes/${n.id}`)}
                  >
                    Manage
                  </button>
                </td>
              </tr>
            );
          })
        )}
      </Table>
    </div>
  );
}
