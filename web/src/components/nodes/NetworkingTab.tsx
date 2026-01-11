import Table from "../Table";
import OSPFTableRow from "../OSPFTableRow";
import { useNodeOSPFNeighbors, useNodeWireGuardPeers } from "@/services/hooks/useNodeNetworking";
import { formatBytes } from "@/utils/format";

export default function NetworkingTab({ nodeId }: { nodeId: number }) {
  const { data: wgPeers, loading: wgLoading, error: wgError, refetch: refetchWG } = useNodeWireGuardPeers(nodeId, { pollingInterval: 30000 });
  const { data: ospfNeighbors, loading: ospfLoading, error: ospfError, refetch: refetchOSPF } = useNodeOSPFNeighbors(nodeId, { pollingInterval: 30000 });

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

  const peers = wgPeers ?? [];
  const neighbors = ospfNeighbors ?? [];

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg text-slate-800">WireGuard Peers</h3>
          {(wgError || ospfError) && (
            <div className="space-x-3 text-sm">
              {wgError && (
                <button className="text-blue-600 hover:text-blue-800" onClick={() => refetchWG()}>
                  Retry WG
                </button>
              )}
              {ospfError && (
                <button className="text-blue-600 hover:text-blue-800" onClick={() => refetchOSPF()}>
                  Retry OSPF
                </button>
              )}
            </div>
          )}
        </div>

        {wgLoading && !wgPeers ? (
          <p className="text-sm text-slate-600">Loading WireGuard peers…</p>
        ) : wgError ? (
          <p className="text-sm text-red-600">{wgError.message}</p>
        ) : (
          <Table columns={['Local IF', 'Peer Node', 'Peer IF', 'Peer Endpoint', 'Status', 'Last Handshake', 'Transfer']}>
            {peers.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-6 px-4 text-sm text-slate-600">
                  No WireGuard peers found for this node yet.
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
                    <td className="py-3 px-4 text-slate-600 font-mono text-sm">{p.local_interface_name}</td>
                    <td className="py-3 px-4 text-slate-800">{p.peer_hostname || `node-${p.peer_node_id}`}</td>
                    <td className="py-3 px-4 text-slate-600 font-mono text-sm">{p.peer_interface_name || '—'}</td>
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
        )}
      </div>

      <div>
        <h3 className="text-lg text-slate-800 mb-3">OSPF Neighbors</h3>
        {ospfLoading && !ospfNeighbors ? (
          <p className="text-sm text-slate-600">Loading OSPF neighbors…</p>
        ) : ospfError ? (
          <p className="text-sm text-red-600">{ospfError.message}</p>
        ) : (
          <Table columns={['Router ID', 'Area', 'State', 'Interface', 'Hello Timer', 'Dead Timer', 'Cost', 'Priority']}>
            {neighbors.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
                  No OSPF neighbors found for this node yet.
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
                  interface={n.interface}
                  helloTimer={n.hello_interval_seconds != null ? `${n.hello_interval_seconds}s` : '—'}
                  deadTimer={n.dead_interval_seconds != null ? `${n.dead_interval_seconds}s` : '—'}
                  cost={n.cost != null ? n.cost : '—'}
                  priority={n.priority != null ? n.priority : '—'}
                />
              ))
            )}
          </Table>
        )}
      </div>

      <p className="text-xs text-slate-500">Auto-updates every 30 seconds.</p>
    </div>
  );
}
