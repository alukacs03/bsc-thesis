import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import { Network, Router, Activity, Wifi, AlertTriangle, Settings } from 'lucide-react';
import Table from "../components/Table"
import OSPFTableRow from "../components/OSPFTableRow";
import { useWireGuardPeers } from "@/services/hooks/useWireGuardPeers";
import { useOSPFNeighbors } from "@/services/hooks/useOSPFNeighbors";
import { useDeploymentSettings } from "@/services/hooks/useDeploymentSettings";
import { settingsAPI } from "@/services/api/settings";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import { ErrorMessage } from "@/components/ErrorMessage";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { formatBytes } from "@/utils/format";
import { handleAPIError } from "@/utils/errorHandler";
import { toast } from "sonner";
import { ChangeEvent, useEffect, useState } from "react";

type DeploymentSettingsForm = {
  loopbackCIDR: string;
  hubToHubCIDR: string;
  hub1WorkerCIDR: string;
  hub2WorkerCIDR: string;
  hub3WorkerCIDR: string;
  kubernetesPodCIDR: string;
  kubernetesServiceCIDR: string;
  ospfArea: string;
  ospfHelloInterval: string;
  ospfDeadInterval: string;
  ospfHubToHubCost: string;
  ospfHubToWorkerCost: string;
  ospfWorkerToHubCost: string;
};

const NetworkingView = () => {
  const { data: wgPeers, loading, error, refetch } = useWireGuardPeers({ pollingInterval: 30000 });
  const { data: ospfNeighbors, loading: ospfLoading, error: ospfError, refetch: refetchOSPF } = useOSPFNeighbors({ pollingInterval: 30000 });
  const { data: deploymentSettings, loading: settingsLoading, error: settingsError, refetch: refetchSettings } = useDeploymentSettings({ pollingInterval: 60000 });
  const [settingsForm, setSettingsForm] = useState<DeploymentSettingsForm>({
    loopbackCIDR: "",
    hubToHubCIDR: "",
    hub1WorkerCIDR: "",
    hub2WorkerCIDR: "",
    hub3WorkerCIDR: "",
    kubernetesPodCIDR: "",
    kubernetesServiceCIDR: "",
    ospfArea: "",
    ospfHelloInterval: "",
    ospfDeadInterval: "",
    ospfHubToHubCost: "",
    ospfHubToWorkerCost: "",
    ospfWorkerToHubCost: "",
  });
  const [savingSettings, setSavingSettings] = useState(false);
  const [showRebuildModal, setShowRebuildModal] = useState(false);
  const [forceRebuild, setForceRebuild] = useState(false);

  useEffect(() => {
    if (!deploymentSettings) return;
    setSettingsForm({
      loopbackCIDR: deploymentSettings.loopback_cidr,
      hubToHubCIDR: deploymentSettings.hub_to_hub_cidr,
      hub1WorkerCIDR: deploymentSettings.hub1_worker_cidr,
      hub2WorkerCIDR: deploymentSettings.hub2_worker_cidr,
      hub3WorkerCIDR: deploymentSettings.hub3_worker_cidr,
      kubernetesPodCIDR: deploymentSettings.kubernetes_pod_cidr,
      kubernetesServiceCIDR: deploymentSettings.kubernetes_service_cidr,
      ospfArea: deploymentSettings.ospf_area.toString(),
      ospfHelloInterval: deploymentSettings.ospf_hello_interval.toString(),
      ospfDeadInterval: deploymentSettings.ospf_dead_interval.toString(),
      ospfHubToHubCost: deploymentSettings.ospf_hub_to_hub_cost.toString(),
      ospfHubToWorkerCost: deploymentSettings.ospf_hub_to_worker_cost.toString(),
      ospfWorkerToHubCost: deploymentSettings.ospf_worker_to_hub_cost.toString(),
    });
  }, [deploymentSettings]);

  if (loading && !wgPeers && ospfLoading && !ospfNeighbors) return <LoadingSpinner />;
  if (error && !wgPeers && ospfError && !ospfNeighbors) {
    return <ErrorMessage message={error.message} onRetry={refetch} />;
  }

  const peers = wgPeers ?? [];
  const neighbors = ospfNeighbors ?? [];
  const online = peers.filter(p => p.ui_status === 'connected').length;
  const potentiallyFailing = peers.filter(p => p.ui_status === 'potentially_failing').length;
  const down = peers.filter(p => p.ui_status === 'down').length;
  const ospfFull = neighbors.filter(n => (n.state || '').toLowerCase().startsWith('full')).length;

  function formatHandshake(ts?: string): string {
    if (!ts) return 'Never';
    const diff = Date.now() - new Date(ts).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  }

  const handleSettingsChange = (key: keyof DeploymentSettingsForm) => (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setSettingsForm((prev) => ({ ...prev, [key]: value }));
  };

  const requiresRebuild = deploymentSettings && (
    settingsForm.loopbackCIDR.trim() !== deploymentSettings.loopback_cidr ||
    settingsForm.hubToHubCIDR.trim() !== deploymentSettings.hub_to_hub_cidr ||
    settingsForm.hub1WorkerCIDR.trim() !== deploymentSettings.hub1_worker_cidr ||
    settingsForm.hub2WorkerCIDR.trim() !== deploymentSettings.hub2_worker_cidr ||
    settingsForm.hub3WorkerCIDR.trim() !== deploymentSettings.hub3_worker_cidr
  );

  const submitSettings = async (rebuild: boolean) => {
    const cidrFields = [
      { label: "Loopback CIDR", value: settingsForm.loopbackCIDR },
      { label: "Hub-to-Hub CIDR", value: settingsForm.hubToHubCIDR },
      { label: "Hub 1 Worker CIDR", value: settingsForm.hub1WorkerCIDR },
      { label: "Hub 2 Worker CIDR", value: settingsForm.hub2WorkerCIDR },
      { label: "Hub 3 Worker CIDR", value: settingsForm.hub3WorkerCIDR },
      { label: "Kubernetes Pod CIDR", value: settingsForm.kubernetesPodCIDR },
      { label: "Kubernetes Service CIDR", value: settingsForm.kubernetesServiceCIDR },
    ];

    for (const field of cidrFields) {
      if (!field.value.trim()) {
        toast.error(`${field.label} is required`);
        return;
      }
    }

    const ospfArea = Number(settingsForm.ospfArea);
    const ospfHello = Number(settingsForm.ospfHelloInterval);
    const ospfDead = Number(settingsForm.ospfDeadInterval);
    const ospfHubToHubCost = Number(settingsForm.ospfHubToHubCost);
    const ospfHubToWorkerCost = Number(settingsForm.ospfHubToWorkerCost);
    const ospfWorkerToHubCost = Number(settingsForm.ospfWorkerToHubCost);

    const numberChecks = [
      { label: "OSPF Area", value: ospfArea },
      { label: "OSPF Hello Interval", value: ospfHello },
      { label: "OSPF Dead Interval", value: ospfDead },
      { label: "OSPF Hub-to-Hub Cost", value: ospfHubToHubCost },
      { label: "OSPF Hub-to-Worker Cost", value: ospfHubToWorkerCost },
      { label: "OSPF Worker-to-Hub Cost", value: ospfWorkerToHubCost },
    ];

    for (const field of numberChecks) {
      if (!Number.isFinite(field.value) || field.value <= 0) {
        toast.error(`${field.label} must be a positive number`);
        return;
      }
    }

    if (!deploymentSettings) {
      toast.error("Deployment settings are not loaded yet");
      return;
    }

    setSavingSettings(true);
    try {
      await settingsAPI.updateDeploymentSettings({
        loopback_cidr: settingsForm.loopbackCIDR.trim(),
        hub_to_hub_cidr: settingsForm.hubToHubCIDR.trim(),
        hub1_worker_cidr: settingsForm.hub1WorkerCIDR.trim(),
        hub2_worker_cidr: settingsForm.hub2WorkerCIDR.trim(),
        hub3_worker_cidr: settingsForm.hub3WorkerCIDR.trim(),
        kubernetes_pod_cidr: settingsForm.kubernetesPodCIDR.trim(),
        kubernetes_service_cidr: settingsForm.kubernetesServiceCIDR.trim(),
        ospf_area: ospfArea,
        ospf_hello_interval: ospfHello,
        ospf_dead_interval: ospfDead,
        ospf_hub_to_hub_cost: ospfHubToHubCost,
        ospf_hub_to_worker_cost: ospfHubToWorkerCost,
        ospf_worker_to_hub_cost: ospfWorkerToHubCost,
        rebuild,
      });
      toast.success("Deployment settings updated");
      refetchSettings();
    } catch (error) {
      const message = handleAPIError(error, "update deployment settings");
      toast.error(message);
    } finally {
      setSavingSettings(false);
    }
  };

  const handleSaveSettings = async () => {
    if (requiresRebuild) {
      setForceRebuild(false);
      setShowRebuildModal(true);
      return;
    }
    await submitSettings(false);
  };

  const handleConfirmRebuild = async () => {
    setShowRebuildModal(false);
    setForceRebuild(false);
    await submitSettings(true);
  };

  const handleForceRebuild = () => {
    setForceRebuild(true);
    setShowRebuildModal(true);
  };

  return (
    <>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <CardWithIcon
                title="WG Peers Online"
                value={online.toString()}
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-100"
                icon={<Wifi className="w-6 h-6 text-green-600"/>}
            />
            <CardWithIcon
                title="WG Potentially Failing"
                value={potentiallyFailing.toString()}
                textColorClass="text-slate-600"
                valueColorClass="text-yellow-600"
                iconBGColorClass="bg-yellow-100"
                icon={<Activity className="w-6 h-6 text-yellow-600"/>}
            />
            <CardWithIcon
                title="WG Peers Down"
                value={down.toString()}
                textColorClass="text-slate-600"
                valueColorClass="text-red-600"
                iconBGColorClass="bg-red-100"
                icon={<AlertTriangle className="w-6 h-6 text-red-600"/>}
            />
            <CardWithIcon
                title="OSPF Full Neighbors"
                value={ospfFull.toString()}
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-100"
                icon={<Network className="w-6 h-6 text-green-600"/>}
            />
        </div>
        <CardContainer title="Deployment Settings" icon={<Settings className="w-5 h-5" />}>
          {settingsError ? (
            <div className="text-sm text-red-600">
              {settingsError.message}
              <button
                className="ml-3 text-blue-600 hover:text-blue-800"
                onClick={() => refetchSettings()}
              >
                Retry
              </button>
            </div>
          ) : settingsLoading && !deploymentSettings ? (
            <div className="text-sm text-slate-600">Loading settings…</div>
          ) : (
            <div className="space-y-6">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div className="text-sm font-semibold text-slate-700">Addressing</div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="loopback-cidr">Management Loopback CIDR</Label>
                      <Input
                        id="loopback-cidr"
                        value={settingsForm.loopbackCIDR}
                        onChange={handleSettingsChange("loopbackCIDR")}
                        placeholder="10.255.0.0/22"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="hub-to-hub-cidr">Hub-to-Hub CIDR</Label>
                      <Input
                        id="hub-to-hub-cidr"
                        value={settingsForm.hubToHubCIDR}
                        onChange={handleSettingsChange("hubToHubCIDR")}
                        placeholder="10.255.4.0/24"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="hub1-worker-cidr">Hub 1 Worker CIDR</Label>
                      <Input
                        id="hub1-worker-cidr"
                        value={settingsForm.hub1WorkerCIDR}
                        onChange={handleSettingsChange("hub1WorkerCIDR")}
                        placeholder="10.255.8.0/22"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="hub2-worker-cidr">Hub 2 Worker CIDR</Label>
                      <Input
                        id="hub2-worker-cidr"
                        value={settingsForm.hub2WorkerCIDR}
                        onChange={handleSettingsChange("hub2WorkerCIDR")}
                        placeholder="10.255.12.0/22"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="hub3-worker-cidr">Hub 3 Worker CIDR</Label>
                      <Input
                        id="hub3-worker-cidr"
                        value={settingsForm.hub3WorkerCIDR}
                        onChange={handleSettingsChange("hub3WorkerCIDR")}
                        placeholder="10.255.16.0/22"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="k8s-pod-cidr">Kubernetes Pod CIDR</Label>
                      <Input
                        id="k8s-pod-cidr"
                        value={settingsForm.kubernetesPodCIDR}
                        onChange={handleSettingsChange("kubernetesPodCIDR")}
                        placeholder="10.244.0.0/16"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="k8s-service-cidr">Kubernetes Service CIDR</Label>
                      <Input
                        id="k8s-service-cidr"
                        value={settingsForm.kubernetesServiceCIDR}
                        onChange={handleSettingsChange("kubernetesServiceCIDR")}
                        placeholder="10.96.0.0/16"
                      />
                    </div>
                  </div>
                </div>
                <div className="space-y-4">
                  <div className="text-sm font-semibold text-slate-700">OSPF</div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="ospf-area">Area</Label>
                      <Input
                        id="ospf-area"
                        type="number"
                        min="1"
                        value={settingsForm.ospfArea}
                        onChange={handleSettingsChange("ospfArea")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="ospf-hello">Hello Interval (s)</Label>
                      <Input
                        id="ospf-hello"
                        type="number"
                        min="1"
                        value={settingsForm.ospfHelloInterval}
                        onChange={handleSettingsChange("ospfHelloInterval")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="ospf-dead">Dead Interval (s)</Label>
                      <Input
                        id="ospf-dead"
                        type="number"
                        min="1"
                        value={settingsForm.ospfDeadInterval}
                        onChange={handleSettingsChange("ospfDeadInterval")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="ospf-hub-hub">Hub-to-Hub Cost</Label>
                      <Input
                        id="ospf-hub-hub"
                        type="number"
                        min="1"
                        value={settingsForm.ospfHubToHubCost}
                        onChange={handleSettingsChange("ospfHubToHubCost")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="ospf-hub-worker">Hub-to-Worker Cost</Label>
                      <Input
                        id="ospf-hub-worker"
                        type="number"
                        min="1"
                        value={settingsForm.ospfHubToWorkerCost}
                        onChange={handleSettingsChange("ospfHubToWorkerCost")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="ospf-worker-hub">Worker-to-Hub Cost</Label>
                      <Input
                        id="ospf-worker-hub"
                        type="number"
                        min="1"
                        value={settingsForm.ospfWorkerToHubCost}
                        onChange={handleSettingsChange("ospfWorkerToHubCost")}
                      />
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex items-center justify-end">
                <div className="flex items-center gap-3">
                  <Button variant="outline" onClick={handleForceRebuild} disabled={savingSettings || settingsLoading}>
                    Force Rebuild
                  </Button>
                  <Button onClick={handleSaveSettings} disabled={savingSettings || settingsLoading}>
                  {savingSettings ? "Saving..." : "Save Settings"}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardContainer>
        <CardContainer title="WireGuard Peers" noPadding={true} icon={<Wifi className="w-5 h-5"/>}>
            <Table
                columns={[
                  'Local Node', 'Local IF', 'Peer Node', 'Peer IF', 'Peer Key', 'Peer Endpoint', 'Status', 'Last Handshake', 'Transfer'
                ]}>
                {peers.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="py-6 px-4 text-sm text-slate-600">
                      No WireGuard peers found yet.
                    </td>
                  </tr>
                ) : (
                  peers.map((p) => {
                    const statusLabel = p.ui_status.toUpperCase();
                    const statusClass =
                      p.ui_status === 'connected'
                        ? 'bg-green-600 text-white'
                        : p.ui_status === 'potentially_failing'
                          ? 'bg-yellow-600 text-white'
                          : p.ui_status === 'down'
                            ? 'bg-red-600 text-white'
                            : 'bg-slate-500 text-white';

                    return (
                      <tr key={p.id} className="border-b border-slate-100 hover:bg-blue-50">
                        <td className="py-3 px-4 text-slate-800">{p.local_node_hostname || `node-${p.local_node_id}`}</td>
                        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{p.local_interface_name}</td>
                        <td className="py-3 px-4 text-slate-800">{p.peer_hostname || `node-${p.peer_node_id}`}</td>
                        <td className="py-3 px-4 text-slate-600 font-mono text-sm">{p.peer_interface_name || '—'}</td>
                        <td className="py-3 px-4">
                          <code className="text-sm bg-slate-100 px-2 py-1 rounded text-slate-600">
                            {p.peer_public_key ? `${p.peer_public_key.slice(0, 6)}...${p.peer_public_key.slice(-4)}` : '—'}
                          </code>
                        </td>
                        <td className="py-3 px-4 text-slate-600">{p.peer_endpoint || '—'}</td>
                        <td className="py-3 px-4">
                          <span className={`px-2 py-1 rounded text-xs ${statusClass}`}>{statusLabel}</span>
                        </td>
                        <td className="py-3 px-4 text-slate-600">{formatHandshake(p.last_handshake_at)}</td>
                        <td className="py-3 px-4 text-slate-600">
                          <div className="text-sm">
                            <div className="flex items-center space-x-2">
                              <span>↓</span>
                              <span>{formatBytes(p.rx_bytes)}</span>
                            </div>
                            <div className="flex items-center space-x-2">
                              <span>↑</span>
                              <span>{formatBytes(p.tx_bytes)}</span>
                            </div>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
             </Table>
        </CardContainer>
        <CardContainer title="OSPF Neighbors" noPadding={true} icon={<Router className="w-5 h-5"/>}>
            <Table
                columns={[
                  'Router ID', 'Area', 'State', 'Interface', 'Hello Timer', 'Dead Timer', 'Cost', 'Priority'
                ]}
            >
                {ospfError ? (
                  <tr>
                    <td colSpan={8} className="py-6 px-4 text-sm text-red-600">
                      {ospfError.message}
                      <button
                        className="ml-3 text-blue-600 hover:text-blue-800"
                        onClick={() => refetchOSPF()}
                      >
                        Retry
                      </button>
                    </td>
                  </tr>
                ) : ospfLoading && !ospfNeighbors ? (
                  <tr>
                    <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                      Loading OSPF neighbors…
                    </td>
                  </tr>
                ) : neighbors.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                      No OSPF neighbors found yet.
                    </td>
                  </tr>
                ) : (
                  neighbors.map((n, idx) => (
                    <OSPFTableRow
                      key={`${n.node_id}-${n.router_id}-${n.interface}-${idx}`}
                      rowKey={`${n.node_id}-${n.router_id}-${n.interface}-${idx}`}
                      routerId={n.router_id}
                      area={n.area || '—'}
                      state={n.state || '—'}
                      interface={`${n.node_hostname}:${n.interface}`}
                      helloTimer={n.hello_interval_seconds != null ? `${n.hello_interval_seconds}s` : '—'}
                      deadTimer={n.dead_interval_seconds != null ? `${n.dead_interval_seconds}s` : '—'}
                      cost={n.cost != null ? n.cost : '—'}
                      priority={n.priority != null ? n.priority : '—'}
                    />
                  ))
                )}
            </Table>
        </CardContainer>
      </div>
      {showRebuildModal && (
        <div className="fixed inset-0 flex items-center justify-center z-50">
          <div className="fixed inset-0 bg-black/60"></div>
          <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 border border-red-200">
            <div className="p-6 border-b border-red-100">
              <div className="text-lg font-semibold text-red-700">
                {forceRebuild ? "Force rebuild networking?" : "Rebuild networking?"}
              </div>
              <div className="text-sm text-slate-600 mt-2">
                {forceRebuild
                  ? "This will rebuild WireGuard and loopback allocations for every node even if the CIDRs are unchanged."
                  : "You changed management or worker CIDRs. Applying this will rebuild WireGuard and loopback allocations for every node."}
              </div>
            </div>
            <div className="p-6 space-y-3 text-sm text-slate-700">
              <div>What happens:</div>
              <div className="bg-red-50 border border-red-100 rounded-md p-3 space-y-1">
                <div>All WireGuard interfaces and peers are regenerated</div>
                <div>Loopback IPs and link subnets are reallocated</div>
                <div>Agents will fetch new configs on their next poll</div>
                <div>Expect brief connectivity loss during the rebuild</div>
              </div>
            </div>
            <div className="flex items-center justify-end gap-3 p-6 border-t border-red-100">
              <Button
                variant="outline"
                onClick={() => {
                  setShowRebuildModal(false);
                  setForceRebuild(false);
                }}
                disabled={savingSettings}
              >
                Cancel
              </Button>
              <Button variant="destructive" onClick={handleConfirmRebuild} disabled={savingSettings}>
                Rebuild and Save
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default NetworkingView
