import CardContainer from "@/components/CardContainer";
import CardWithIcon from "@/components/CardWithIcon";
import type { KubernetesClusterSummary } from "@/services/types/kubernetes";
import type { Node } from "@/services/types/node";
import { formatBytes } from "@/utils/format";
import { CheckCircle2, AlertTriangle, Server, Network, Boxes } from "lucide-react";

export default function ClusterOverviewTab({
  cluster,
  nodes,
  onRefreshJoinTokens,
  refreshingJoinTokens,
}: {
  cluster: KubernetesClusterSummary | null;
  nodes: Node[];
  onRefreshJoinTokens?: () => void;
  refreshingJoinTokens?: boolean;
}) {
  const total = nodes.length;
  const online = nodes.filter((n) => n.status === "active").length;
  const offline = nodes.filter((n) => n.status === "offline").length;
  const workers = nodes.filter((n) => n.role === "worker").length;

  let diskTotal = 0;
  let diskUsed = 0;
  for (const n of nodes) {
    if (typeof n.disk_total_bytes === "number" && n.disk_total_bytes > 0) diskTotal += n.disk_total_bytes;
    if (typeof n.disk_used_bytes === "number" && n.disk_used_bytes >= 0) diskUsed += n.disk_used_bytes;
  }

  const healthLabel = offline > 0 ? "Degraded" : online > 0 ? "Healthy" : "Unknown";
  const initialized = Boolean(cluster?.initialized_at);

  const bootstrapNode =
    cluster?.bootstrap_node_id ? nodes.find((n) => n.id === cluster.bootstrap_node_id) : undefined;

  const controlPlaneDesired = nodes.filter((n) => n.role === "hub" || n.reported_desired_role === "hub");
  const controlPlaneJoined = controlPlaneDesired.filter(
    (n) => n.k8s_state === "joined_control_plane" || n.k8s_state === "cluster_initialized"
  );

  const joinExpiryLabel = (() => {
    if (!cluster?.join_command_expires_at) return null;
    const diff = new Date(cluster.join_command_expires_at).getTime() - Date.now();
    const minutes = Math.floor(diff / 60000);
    if (!Number.isFinite(minutes)) return null;
    if (minutes <= 0) return "expired";
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    if (hours < 48) return `${hours}h`;
    return `${Math.floor(hours / 24)}d`;
  })();

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <CardWithIcon
          title="Cluster Health"
          value={healthLabel}
          textColorClass="text-slate-600"
          valueColorClass={offline > 0 ? "text-yellow-700" : "text-green-700"}
          iconBGColorClass={offline > 0 ? "bg-yellow-100" : "bg-green-100"}
          icon={offline > 0 ? <AlertTriangle className="w-6 h-6 text-yellow-700" /> : <CheckCircle2 className="w-6 h-6 text-green-700" />}
        />
        <CardWithIcon
          title="Nodes Online"
          value={`${online}/${total}`}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Server className="w-6 h-6 text-slate-800" />}
        />
        <CardWithIcon
          title="Control Plane Joined"
          value={`${controlPlaneJoined.length}/${controlPlaneDesired.length}`}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Network className="w-6 h-6 text-slate-800" />}
        />
        <CardWithIcon
          title="Workers"
          value={workers.toString()}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Boxes className="w-6 h-6 text-slate-800" />}
        />
      </div>

      <CardContainer
        title="Kubernetes Cluster"
        icon={<Server className="w-5 h-5" />}
        button={
          cluster && onRefreshJoinTokens ? (
            <button
              className="bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition disabled:opacity-60"
              onClick={onRefreshJoinTokens}
              disabled={refreshingJoinTokens}
              title="Requests the bootstrap hub to regenerate join tokens (tokens are never shown in the UI)."
            >
              {refreshingJoinTokens ? "Refreshing..." : "Refresh Join Tokens"}
            </button>
          ) : null
        }
      >
        {cluster ? (
          <div className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Status</span>
              <span className={`font-medium ${initialized ? "text-green-700" : "text-yellow-700"}`}>
                {initialized ? "Initialized" : "Not initialized"}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Kubernetes version</span>
              <span className="text-slate-800 font-mono">{cluster.kubernetes_version || "unknown"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Control-plane endpoint</span>
              <span className="text-slate-800 font-mono">{cluster.control_plane_endpoint || "unknown"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Pod CIDR</span>
              <span className="text-slate-800 font-mono">{cluster.pod_cidr || "unknown"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Service CIDR</span>
              <span className="text-slate-800 font-mono">{cluster.service_cidr || "unknown"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Bootstrap node</span>
              <span className="text-slate-800">{bootstrapNode ? bootstrapNode.hostname : cluster.bootstrap_node_id ?? "unknown"}</span>
            </div>
            {joinExpiryLabel && (
              <div className="flex items-center justify-between">
                <span className="text-slate-600">Join token</span>
                <span className={`font-medium ${joinExpiryLabel === "expired" ? "text-red-700" : "text-slate-800"}`}>
                  {joinExpiryLabel === "expired" ? "Expired" : `Expires in ${joinExpiryLabel}`}
                </span>
              </div>
            )}
            {!joinExpiryLabel && (
              <div className="flex items-center justify-between">
                <span className="text-slate-600">Join token</span>
                <span className="text-slate-700">Unknown</span>
              </div>
            )}
            <p className="text-xs text-slate-500 pt-2">
              Managed by Gluon agents. This view updates every ~30 seconds.
            </p>
          </div>
        ) : (
          <p className="text-sm text-slate-600">Cluster state not available yet.</p>
        )}
      </CardContainer>

      <CardContainer title="Single Cluster Mode" icon={<Server className="w-5 h-5" />}>
        <div className="space-y-2">
          <p className="text-sm text-slate-700">
            Gluon currently manages exactly one Kubernetes cluster: the Gluon cluster itself.
          </p>
          <p className="text-sm text-slate-600">
            Kubernetes objects (namespaces/workloads/services/events) will be added to the backend later.
          </p>
        </div>
      </CardContainer>

      <CardContainer title="Useful Now" icon={<CheckCircle2 className="w-5 h-5" />}>
        <ul className="text-sm text-slate-600 list-disc pl-5">
          <li>Use Nodes for host health/telemetry and agent logs</li>
          <li>Use Networking for WireGuard + OSPF visibility</li>
          <li>Disk telemetry: {formatBytes(diskUsed)} / {formatBytes(diskTotal)} used (aggregated)</li>
          <li>Control-plane should match hubs (auto-healed)</li>
        </ul>
      </CardContainer>
    </div>
  );
}
