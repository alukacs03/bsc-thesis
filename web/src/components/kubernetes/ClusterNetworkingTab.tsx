import { useState } from "react";
import CardContainer from "@/components/CardContainer";
import CardWithIcon from "@/components/CardWithIcon";
import { ErrorMessage } from "@/components/ErrorMessage";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import Badge from "@/components/Badge";
import Table from "@/components/Table";
import DeployModal, { type EditResourceInfo } from "@/components/kubernetes/DeployModal";
import { useKubernetesNetworking } from "@/services/hooks/useKubernetesNetworking";
import { kubernetesAPI } from "@/services/api/kubernetes";
import { AlertTriangle, Globe, Lock, Network, Pencil, Plus, RefreshCw, Server, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";

function formatAgeSeconds(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "-";
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

function getServiceTypeBadge(type: string): { label: string; className: string } {
  switch (type) {
    case "LoadBalancer":
      return { label: "LOAD BALANCER", className: "bg-purple-600 text-white" };
    case "NodePort":
      return { label: "NODE PORT", className: "bg-blue-600 text-white" };
    case "ExternalName":
      return { label: "EXTERNAL", className: "bg-orange-600 text-white" };
    case "ClusterIP":
    default:
      return { label: "CLUSTER IP", className: "bg-slate-600 text-white" };
  }
}

export default function ClusterNetworkingTab() {
  const { data, loading, error, refetch } = useKubernetesNetworking({ pollingInterval: 30000 });
  const [showDeployModal, setShowDeployModal] = useState(false);
  const [editResource, setEditResource] = useState<EditResourceInfo | undefined>(undefined);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);

  if (loading && !data) return <LoadingSpinner />;
  if (error && !data) return <ErrorMessage message={error.message} onRetry={() => void refetch()} />;

  const services = data?.services ?? [];
  const ingresses = data?.ingresses ?? [];

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

  // Count service types
  const servicesByType = services.reduce(
    (acc, s) => {
      acc[s.type] = (acc[s.type] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>
  );

  const loadBalancerCount = servicesByType["LoadBalancer"] || 0;
  const nodePortCount = servicesByType["NodePort"] || 0;
  const clusterIPCount = servicesByType["ClusterIP"] || 0;
  const ingressWithTLS = ingresses.filter((i) => i.tls).length;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h4 className="text-lg text-slate-800">Services & Ingress</h4>
          <p className="text-xs text-slate-500">
            Kubernetes networking resources. Services expose pods internally, Ingress provides external HTTP routing.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {updatedAgo && <div className="text-xs text-slate-500">Updated {updatedAgo} ago</div>}
          <button
            onClick={() => setShowDeployModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center space-x-2"
          >
            <Plus className="w-4 h-4" />
            <span>Create</span>
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
          title="Total Services"
          value={services.length.toString()}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Server className="w-6 h-6 text-slate-800" />}
          noHover
        />
        <CardWithIcon
          title="Ingress Resources"
          value={ingresses.length.toString()}
          hint={ingressWithTLS > 0 ? `${ingressWithTLS} with TLS` : undefined}
          textColorClass="text-slate-600"
          valueColorClass={ingresses.length > 0 ? "text-slate-800" : "text-slate-500"}
          iconBGColorClass={ingresses.length > 0 ? "bg-green-100" : "bg-slate-100"}
          icon={<Globe className={`w-6 h-6 ${ingresses.length > 0 ? "text-green-700" : "text-slate-500"}`} />}
          noHover
        />
        <CardWithIcon
          title="LoadBalancer SVCs"
          value={loadBalancerCount.toString()}
          hint={nodePortCount > 0 ? `${nodePortCount} NodePort` : undefined}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-purple-100"
          icon={<Network className="w-6 h-6 text-purple-700" />}
          noHover
        />
        <CardWithIcon
          title="ClusterIP SVCs"
          value={clusterIPCount.toString()}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-slate-100"
          icon={<Server className="w-6 h-6 text-slate-600" />}
          noHover
        />
      </div>

      <CardContainer title="Services" icon={<Server className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          <Table columns={["Namespace", "Name", "Type", "Cluster IP", "External IP", "Ports", "Age", ""]}>
            {services.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                  No services found.
                </td>
              </tr>
            ) : (
              services.map((s) => {
                const typeBadge = getServiceTypeBadge(s.type);
                const portsLabel = s.ports.length === 0 ? "-" : s.ports.length <= 2 ? s.ports.join(", ") : `${s.ports[0]} (+${s.ports.length - 1})`;
                return (
                  <tr key={`${s.namespace}/${s.name}`} className="border-b border-slate-100 hover:bg-blue-50">
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{s.namespace}</td>
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{s.name}</td>
                    <td className="py-3 px-4">
                      <Badge className={typeBadge.className}>{typeBadge.label}</Badge>
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{s.cluster_ip || "-"}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{s.external_ip || "-"}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-xs" title={s.ports.join("\n") || undefined}>
                      {portsLabel}
                    </td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{formatAgeSeconds(s.age_seconds)}</td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => {
                            setEditResource({ namespace: s.namespace, kind: "Service", name: s.name });
                            setShowDeployModal(true);
                          }}
                          className="p-1.5 text-slate-500 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title="Edit Service"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => void handleDeleteResource(s.namespace, "Service", s.name)}
                          className="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          title="Delete Service"
                          disabled={deletingKey === `${s.namespace}/Service/${s.name}`}
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
            Services expose pods within the cluster or externally. Hover on ports to see all port mappings.
          </p>
        </div>
      </CardContainer>

      <CardContainer title="Ingress Resources" icon={<Globe className="w-5 h-5" />} noPadding>
        <div className="p-4 md:p-6">
          {ingresses.length === 0 ? (
            <div className="py-6 text-center">
              <Globe className="w-12 h-12 text-slate-300 mx-auto mb-3" />
              <p className="text-sm text-slate-600 mb-2">No Ingress resources found.</p>
              <p className="text-xs text-slate-500">
                Install an Ingress controller (like nginx-ingress) and create Ingress resources to expose services externally.
              </p>
            </div>
          ) : (
            <>
              <Table columns={["Namespace", "Name", "Class", "TLS", "Hosts", "Address", "Age", ""]}>
                {ingresses.map((ing) => {
                  const hostsFromRules = [...new Set(ing.rules.map((r) => r.host).filter(Boolean))];
                  const hostsLabel =
                    hostsFromRules.length === 0
                      ? "*"
                      : hostsFromRules.length === 1
                        ? hostsFromRules[0]
                        : `${hostsFromRules[0]} (+${hostsFromRules.length - 1})`;
                  return (
                    <tr key={`${ing.namespace}/${ing.name}`} className="border-b border-slate-100 hover:bg-blue-50">
                      <td className="py-3 px-4 text-slate-700 font-mono text-sm">{ing.namespace}</td>
                      <td className="py-3 px-4 text-slate-800 font-mono text-sm">{ing.name}</td>
                      <td className="py-3 px-4 text-slate-700 text-sm">{ing.ingress_class || "-"}</td>
                      <td className="py-3 px-4">
                        {ing.tls ? (
                          <span className="inline-flex items-center gap-1 text-green-700">
                            <Lock className="w-3.5 h-3.5" />
                            <span className="text-xs font-medium">HTTPS</span>
                          </span>
                        ) : (
                          <span className="text-slate-400 text-xs">HTTP</span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-slate-700 font-mono text-sm" title={hostsFromRules.join("\n") || undefined}>
                        {hostsLabel}
                      </td>
                      <td className="py-3 px-4 text-slate-700 font-mono text-sm">{ing.address || "-"}</td>
                      <td className="py-3 px-4 text-slate-700 font-mono text-sm">{formatAgeSeconds(ing.age_seconds)}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => {
                              setEditResource({ namespace: ing.namespace, kind: "Ingress", name: ing.name });
                              setShowDeployModal(true);
                            }}
                            className="p-1.5 text-slate-500 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                            title="Edit Ingress"
                          >
                            <Pencil className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => void handleDeleteResource(ing.namespace, "Ingress", ing.name)}
                            className="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            title="Delete Ingress"
                            disabled={deletingKey === `${ing.namespace}/Ingress/${ing.name}`}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </Table>
              <p className="text-xs text-slate-500 pt-3">
                Ingress resources route external HTTP/HTTPS traffic to services based on host and path rules.
              </p>
            </>
          )}
        </div>
      </CardContainer>

      {ingresses.length > 0 && (
        <CardContainer title="Ingress Rules" icon={<Network className="w-5 h-5" />} noPadding>
          <div className="p-4 md:p-6">
            <Table columns={["Ingress", "Host", "Path", "Path Type", "Backend Service", "Port"]}>
              {ingresses.flatMap((ing) =>
                ing.rules.map((rule, idx) => (
                  <tr key={`${ing.namespace}/${ing.name}/${idx}`} className="border-b border-slate-100 hover:bg-blue-50">
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">
                      {ing.namespace}/{ing.name}
                    </td>
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{rule.host || "*"}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{rule.path || "/"}</td>
                    <td className="py-3 px-4 text-slate-600 text-sm">{rule.path_type || "Prefix"}</td>
                    <td className="py-3 px-4 text-slate-800 font-mono text-sm">{rule.service_name}</td>
                    <td className="py-3 px-4 text-slate-700 font-mono text-sm">{rule.service_port}</td>
                  </tr>
                ))
              )}
            </Table>
            <p className="text-xs text-slate-500 pt-3">
              Each rule maps a host/path combination to a backend service. Requests matching these rules are routed accordingly.
            </p>
          </div>
        </CardContainer>
      )}

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
    </div>
  );
}
