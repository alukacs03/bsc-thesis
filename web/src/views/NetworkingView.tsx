import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import { Network, Router, Activity, Wifi, AlertTriangle } from 'lucide-react';
import Table from "../components/Table"
import OSPFTableRow from "../components/OSPFTableRow";
import { useWireGuardPeers } from "@/services/hooks/useWireGuardPeers";
import { useOSPFNeighbors } from "@/services/hooks/useOSPFNeighbors";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import { ErrorMessage } from "@/components/ErrorMessage";
import { formatBytes } from "@/utils/format";

const NetworkingView = () => {
  const { data: wgPeers, loading, error, refetch } = useWireGuardPeers({ pollingInterval: 30000 });
  const { data: ospfNeighbors, loading: ospfLoading, error: ospfError, refetch: refetchOSPF } = useOSPFNeighbors({ pollingInterval: 30000 });

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
    </>
  )
}

export default NetworkingView
