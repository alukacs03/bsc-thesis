import { useState } from "react";
import CardContainer from "@/components/CardContainer";
import CardWithIcon from "@/components/CardWithIcon";
import { ErrorMessage } from "@/components/ErrorMessage";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import Badge from "@/components/Badge";
import Table from "@/components/Table";
import DeployModal, { type EditResourceInfo } from "@/components/kubernetes/DeployModal";
import ExposeServiceModal, { type ExposeTargetInfo } from "@/components/kubernetes/ExposeServiceModal";
import { useKubernetesWorkloads } from "@/services/hooks/useKubernetesWorkloads";
import { kubernetesAPI } from "@/services/api/kubernetes";
import type { KubernetesWorkloadNamespaceSummary, KubernetesWorkloadPodIssue } from "@/services/types/kubernetes";
import { AlertTriangle, Boxes, Layers, Pencil, Plus, RefreshCw, Server, Share2, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";

function formatAgeSeconds(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "—";
  const total = Math.floor(value);
  if (total < 60) return `${total}s`;
  const minutes = Math.floor(total / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
}

function formatSinceISO(iso?: string): string | null {
  if (!iso) return null;
  const t = new Date(iso).getTime();
  if (!Number.isFinite(t)) return null;
  const diffSeconds = Math.max(0, Math.floor((Date.now() - t) / 1000));
  return formatAgeSeconds(diffSeconds);
}

function getNamespaceHealth(ns: KubernetesWorkloadNamespaceSummary): { label: string; className: string } {
  const hasAny =
    ns.deployments_total +
      ns.daemonsets_total +
      ns.statefulsets_total +
      ns.jobs_total +
      ns.pods_total >
    0;

  if (!hasAny) return { label: "EMPTY", className: "bg-slate-200 text-slate-800" };

  const critical = ns.pods_failed > 0 || ns.jobs_failed > 0;
  const warning =
    ns.pods_pending > 0 ||
    ns.pods_unhealthy > 0 ||
    (ns.deployments_total > 0 && ns.deployments_ready < ns.deployments_total) ||
    (ns.daemonsets_total > 0 && ns.daemonsets_ready < ns.daemonsets_total) ||
    (ns.statefulsets_total > 0 && ns.statefulsets_ready < ns.statefulsets_total);

  if (critical) return { label: "CRITICAL", className: "bg-red-600 text-white" };
  if (warning) return { label: "DEGRADED", className: "bg-yellow-600 text-white" };
  return { label: "HEALTHY", className: "bg-green-600 text-white" };
}

function getPodStatusBadge(issue: KubernetesWorkloadPodIssue): { label: string; className: string } {
  const phase = (issue.phase || "").toLowerCase();
  const reason = (issue.reason || "").toLowerCase();

  if (phase === "failed") return { label: "FAILED", className: "bg-red-600 text-white" };
  if (reason.includes("crashloopbackoff")) return { label: "CRASHLOOP", className: "bg-red-600 text-white" };
  if (phase === "pending") return { label: "PENDING", className: "bg-yellow-600 text-white" };
  return { label: issue.phase?.toUpperCase() || "UNKNOWN", className: "bg-slate-600 text-white" };
}

export default function ClusterWorkloadsTab() {
  const { data, loading, error, refetch } = useKubernetesWorkloads({ pollingInterval: 30000 });
  const [showDeployModal, setShowDeployModal] = useState(false);
  const [editResource, setEditResource] = useState<EditResourceInfo | undefined>(undefined);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [exposeTarget, setExposeTarget] = useState<ExposeTargetInfo | undefined>(undefined);

  if (loading && !data) return <LoadingSpinner />;
  if (error && !data) return <ErrorMessage message={error.message} onRetry={() => void refetch()} />;

  const namespaces = data?.namespaces ?? [];
  const nodes = data?.nodes ?? [];
  const issues = data?.unhealthy_pods ?? [];
  const resources = data?.resources ?? [];

  const totals = namespaces.reduce(
    (acc, ns) => {
      acc.deployments_total += ns.deployments_total;
      acc.deployments_ready += ns.deployments_ready;
      acc.daemonsets_total += ns.daemonsets_total;
      acc.daemonsets_ready += ns.daemonsets_ready;
      acc.statefulsets_total += ns.statefulsets_total;
      acc.statefulsets_ready += ns.statefulsets_ready;
      acc.jobs_total += ns.jobs_total;
      acc.jobs_failed += ns.jobs_failed;
      acc.pods_total += ns.pods_total;
      acc.pods_running += ns.pods_running;
      acc.pods_pending += ns.pods_pending;
      acc.pods_failed += ns.pods_failed;
      acc.restarts_total += ns.restarts_total;
      return acc;
    },
    {
      deployments_total: 0,
      deployments_ready: 0,
      daemonsets_total: 0,
      daemonsets_ready: 0,
      statefulsets_total: 0,
      statefulsets_ready: 0,
      jobs_total: 0,
      jobs_failed: 0,
      pods_total: 0,
      pods_running: 0,
      pods_pending: 0,
      pods_failed: 0,
      restarts_total: 0,
    }
  );

  const crashloops = issues.filter((i) => (i.reason || "").toLowerCase().includes("crashloopbackoff")).length;
  const unscheduled = issues.filter((i) => !i.node || i.node.trim() === "").length;
  const updatedAgo = formatSinceISO(data?.generated_at);

  const handleDeleteResource = async (namespace: string, kind: string, name: string) => {
    const label = `${namespace}/${name}`;
    if (!window.confirm(`Delete ${kind} ${label}? This cannot be undone.`)) {
      return;
    }
    const key = `${namespace}/${kind}/${name}`;
    try {
      setDeletingKey(key);
      await kubernetesAPI.deleteResource(namespace, kind, name);
      toast.success(`${kind} deleted`, { description: label });
      void refetch();
    } catch (err) {
      toast.error(handleAPIError(err, "delete resource"));
    } finally {
      setDeletingKey((current) => (current === key ? null : current));
    }
  };

  const sortedNamespaces = [...namespaces].sort((a, b) => {
    const sa = getNamespaceHealth(a).label;
    const sb = getNamespaceHealth(b).label;
    const score = (label: string) => {
      switch (label) {
        case "CRITICAL":
          return 0;
        case "DEGRADED":
          return 1;
        case "HEALTHY":
          return 2;
        default:
          return 3;
      }
    };
    if (score(sa) !== score(sb)) return score(sa) - score(sb);
    return a.namespace.localeCompare(b.namespace);
  });

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h4 className="text-lg text-slate-800">Workloads</h4>
          <p className="text-xs text-slate-500">
            Aggregated from the live cluster via the API. Updates every ~30 seconds.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {updatedAgo && <div className="text-xs text-slate-500">Updated {updatedAgo} ago</div>}
          <button
            onClick={() => setShowDeployModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center space-x-2"
          >
            <Plus className="w-4 h-4" />
            <span>Deploy</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="flex items-center gap-2 text-sm text-yellow-800 bg-yellow-50 border border-yellow-200 rounded p-3">
          <AlertTriangle className="w-4 h-4" />
          <span className="flex-1">Last update failed: {error.message}</span>
          <button
            className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-yellow-600 text-white rounded hover:bg-yellow-700"
            onClick={() => void refetch()}
          >
            <RefreshCw className="w-3 h-3" />
            Retry
          </button>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <CardWithIcon
          title="Pods Running"
          value={`${totals.pods_running}/${totals.pods_total}`}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Boxes className="w-6 h-6 text-slate-800" />}
          noHover
        />
        <CardWithIcon
          title="Unhealthy Pods"
          value={issues.length.toString()}
          hint={crashloops > 0 ? `${crashloops} crashloops` : undefined}
          textColorClass="text-slate-600"
          valueColorClass={issues.length > 0 ? "text-yellow-700" : "text-green-700"}
          iconBGColorClass={issues.length > 0 ? "bg-yellow-100" : "bg-green-100"}
          icon={<AlertTriangle className={`w-6 h-6 ${issues.length > 0 ? "text-yellow-700" : "text-green-700"}`} />}
          noHover
        />
        <CardWithIcon
          title="Restarts (Total)"
          value={totals.restarts_total.toString()}
          hint={unscheduled > 0 ? `${unscheduled} unscheduled` : undefined}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Layers className="w-6 h-6 text-slate-800" />}
          noHover
        />
        <CardWithIcon
          title="Deployments Ready"
          value={`${totals.deployments_ready}/${totals.deployments_total}`}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Server className="w-6 h-6 text-slate-800" />}
          noHover
        />
      </div>

      <CardContainer title="Per-namespace Inventory" icon={<Boxes className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          <Table
            columns={[
              "Namespace",
              "Health",
              "Deployments",
              "DaemonSets",
              "StatefulSets",
              "Jobs (A/S/F)",
              "Pods (R/P/F/U/T)",
              "Restarts",
            ]}
          >
            {sortedNamespaces.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                  No workloads found.
                </td>
              </tr>
            ) : (
              sortedNamespaces.map((ns) => {
                const health = getNamespaceHealth(ns);
                return (
                  <tr key={ns.namespace} className="border-b border-slate-100 hover:bg-blue-50">
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{ns.namespace}</td>
                    <td className="py-3 px-4">
                      <Badge className={health.className}>{health.label}</Badge>
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ns.deployments_ready}/{ns.deployments_total}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ns.daemonsets_ready}/{ns.daemonsets_total}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ns.statefulsets_ready}/{ns.statefulsets_total}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ns.jobs_active}/{ns.jobs_succeeded}/{ns.jobs_failed}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ns.pods_running}/{ns.pods_pending}/{ns.pods_failed}/{ns.pods_unhealthy}/{ns.pods_total}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{ns.restarts_total}</td>
                  </tr>
                );
              })
            )}
          </Table>
          <p className="text-xs text-slate-500 pt-3">
            Health is derived from ready/desired counts, pending/failed pods, and failed jobs.
          </p>
        </div>
      </CardContainer>

      <CardContainer title="Unhealthy Pods (last 50)" icon={<AlertTriangle className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          <Table columns={["Namespace", "Pod", "Node", "Status", "Reason", "Restarts", "Age", "Images"]}>
            {issues.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                  No unhealthy pods detected.
                </td>
              </tr>
            ) : (
              issues.map((p) => {
                const status = getPodStatusBadge(p);
                const reason = p.reason?.trim() || "—";
                const images = (p.images ?? []).filter(Boolean);
                const imagesLabel = images.length === 0 ? "—" : images.length === 1 ? images[0] : `${images[0]} (+${images.length - 1})`;
                return (
                  <tr key={`${p.namespace}/${p.name}`} className="border-b border-slate-100 hover:bg-blue-50">
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{p.namespace}</td>
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{p.name}</td>
                    <td className="py-3 px-4 text-slate-600 font-mono text-sm">{p.node?.trim() ? p.node : "—"}</td>
                    <td className="py-3 px-4">
                      <Badge className={status.className}>{status.label}</Badge>
                    </td>
                    <td className="py-3 px-4 text-slate-700" title={p.message?.trim() || undefined}>
                      {reason}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{p.restarts}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{formatAgeSeconds(p.age_seconds)}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-xs" title={images.join("\n") || undefined}>
                      {imagesLabel}
                    </td>
                  </tr>
                );
              })
            )}
          </Table>
          <p className="text-xs text-slate-500 pt-3">
            Shows pending/failed/non-ready pods. Hover “Reason” to see the message. Images reflect the pod spec.
          </p>
        </div>
      </CardContainer>

      <CardContainer title="Workload Resources" icon={<Boxes className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          <Table columns={["Namespace", "Kind", "Name", "Ready", "Age", "Images", ""]}>
            {resources.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-6 px-4 text-sm text-slate-600">
                  No workload resources found.
                </td>
              </tr>
            ) : (
              resources.map((r) => {
                const images = (r.images ?? []).filter(Boolean);
                const imagesLabel = images.length === 0 ? "—" : images.length === 1 ? images[0] : `${images[0]} (+${images.length - 1})`;
                const canExpose = ["Deployment", "DaemonSet", "StatefulSet"].includes(r.kind);
                return (
                  <tr key={`${r.namespace}/${r.kind}/${r.name}`} className="border-b border-slate-100 hover:bg-blue-50">
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{r.namespace}</td>
                    <td className="py-3 px-4 text-slate-700 text-sm">{r.kind}</td>
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{r.name}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{r.ready}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{formatAgeSeconds(r.age_seconds)}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-xs" title={images.join("\n") || undefined}>
                      {imagesLabel}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => {
                            setEditResource({ namespace: r.namespace, kind: r.kind, name: r.name });
                            setShowDeployModal(true);
                          }}
                          className="p-1.5 text-slate-500 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title={`Edit ${r.kind}`}
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        {canExpose && (
                          <button
                            onClick={() => setExposeTarget({ namespace: r.namespace, kind: r.kind, name: r.name })}
                            className="p-1.5 text-slate-500 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                            title="Expose via NodePort"
                          >
                            <Share2 className="w-4 h-4" />
                          </button>
                        )}
                        <button
                          onClick={() => void handleDeleteResource(r.namespace, r.kind, r.name)}
                          className="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          title={`Delete ${r.kind}`}
                          disabled={deletingKey === `${r.namespace}/${r.kind}/${r.name}`}
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </Table>
          <p className="text-xs text-slate-500 pt-3">
            Deployments, DaemonSets, and StatefulSets. Click the edit button to modify a resource.
          </p>
        </div>
      </CardContainer>

      <CardContainer title="Node Placement" icon={<Server className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          <Table columns={["Node", "Pods", "Unhealthy Pods"]}>
            {nodes.length === 0 ? (
              <tr>
                <td colSpan={3} className="py-6 px-4 text-sm text-slate-600">
                  No node placement data yet.
                </td>
              </tr>
            ) : (
              nodes.map((n) => (
                <tr key={n.node} className="border-b border-slate-100 hover:bg-blue-50">
                  <td className="py-3 px-4 text-slate-800 font-mono text-sm">{n.node}</td>
                  <td className="py-3 px-4 text-slate-700 font-mono text-sm">{n.pods}</td>
                  <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                    {n.unhealthy_pods > 0 ? (
                      <span className="text-yellow-700 font-medium">{n.unhealthy_pods}</span>
                    ) : (
                      "0"
                    )}
                  </td>
                </tr>
              ))
            )}
          </Table>
          <p className="text-xs text-slate-500 pt-3">
            Quick view of pod distribution and where unhealthy pods are concentrated.
          </p>
        </div>
      </CardContainer>

      {showDeployModal && (
        <DeployModal
          onClose={() => {
            setShowDeployModal(false);
            setEditResource(undefined);
          }}
          onSuccess={() => {
            void refetch();
          }}
          editResource={editResource}
        />
      )}

      {exposeTarget && (
        <ExposeServiceModal
          target={exposeTarget}
          onClose={() => {
            setExposeTarget(undefined);
          }}
          onSuccess={() => {
            void refetch();
          }}
        />
      )}
    </div>
  );
}
