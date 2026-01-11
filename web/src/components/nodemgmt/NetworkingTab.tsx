import CardContainer from "../CardContainer";
import Table from "../Table";
import { useNodeOSPFNeighbors, useNodeWireGuardPeers } from "@/services/hooks/useNodeNetworking";
import { formatBytes } from "@/utils/format";
import OSPFTableRow from "../OSPFTableRow";
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
      <CardContainer title="WireGuard Peers" icon={undefined} noPadding={true}>
        {wgError ? (
          <div className="p-6 text-sm text-red-600">
            {wgError.message}{' '}
            <button className="ml-2 text-blue-600 hover:text-blue-800" onClick={() => refetchWG()}>
              Retry
            </button>
          </div>
        ) : wgLoading && !wgPeers ? (
          <div className="p-6 text-sm text-slate-600">Loading WireGuard peers…</div>
        ) : (
          <Table
            columns={[
              'Local IF',
              'Peer Node',
              'Peer IF',
              'Peer Key',
              'Peer Endpoint',
              'Status',
              'Last Handshake',
              'Transfer',
            ]}
          >
            {peers.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-6 px-4 text-sm text-slate-600">
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
        )}
      </CardContainer>

      <CardContainer title="OSPF Neighbors" icon={undefined} noPadding={true}>
        {ospfError ? (
          <div className="p-6 text-sm text-red-600">
            {ospfError.message}{' '}
            <button className="ml-2 text-blue-600 hover:text-blue-800" onClick={() => refetchOSPF()}>
              Retry
            </button>
          </div>
        ) : ospfLoading && !ospfNeighbors ? (
          <div className="p-6 text-sm text-slate-600">Loading OSPF neighbors…</div>
        ) : (
          <Table
            columns={[
              'Router ID',
              'Area',
              'State',
              'Interface',
              'Hello Timer',
              'Dead Timer',
              'Cost',
              'Priority',
            ]}
          >
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
      </CardContainer>

      <p className="text-xs text-slate-500">Auto-updates every 30 seconds.</p>
    </div>
  );
}
