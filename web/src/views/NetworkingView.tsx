import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import { Network, Router, Activity, Wifi, AlertTriangle } from 'lucide-react';
import Table from "../components/Table"
import WGTableRow from "../components/WGTableRow";
import OSPFTableRow from "../components/OSPFTableRow";

const NetworkingView = () => {
  return (
    <>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <CardWithIcon
                title="WG Peers Online"
                value="3"
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-100"
                icon={<Wifi className="w-6 h-6 text-green-600"/>}
            />
            <CardWithIcon
                title="WG Peers Flapping"
                value="2"
                textColorClass="text-slate-600"
                valueColorClass="text-yellow-600"
                iconBGColorClass="bg-yellow-100"
                icon={<Activity className="w-6 h-6 text-yellow-600"/>}
            />
            <CardWithIcon
                title="WG Peers Down"
                value="1"
                textColorClass="text-slate-600"
                valueColorClass="text-red-600"
                iconBGColorClass="bg-red-100"
                icon={<AlertTriangle className="w-6 h-6 text-red-600"/>}
            />
            <CardWithIcon
                title="OSPF Full Neighbors"
                value="5"
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-100"
                icon={<Network className="w-6 h-6 text-green-600"/>}
            />
        </div>
        <CardContainer title="WireGuard Peers" noPadding={true} icon={<Wifi className="w-5 h-5"/>}>
            <Table
                columns={[
                  'Peer Name', 'Public Key', 'Endpoint', 'Status', 'Last Handshake', 'Transfer', 'Actions'
                ]}>
                <WGTableRow 
                  rowKey="peer1"
                  status="connected"
                  name="gluon-worker-01"
                  pubKey="abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx5678yzab9012cdef3456"
                  endpoint="peer1.example.com"
                  latestHandshake="2024-10-01 12:34:56"
                  transferRx="1.2 GB"
                  transferTx="800 MB"
                />
                <WGTableRow
                  rowKey="peer2"
                  status="flapping"
                  name="gluon-worker-02"
                  pubKey="mnop3456qrst7890uvwx5678yzab9012cdef3456abcd1234efgh5678ijkl9012"
                  endpoint="peer2.example.com"
                  latestHandshake="2024-10-01 12:30:00"
                  transferRx="500 MB"
                  transferTx="300 MB"
                />
                <WGTableRow
                  rowKey="peer3"
                  status="down"
                  name="gluon-controller-01"
                  pubKey="yzab9012cdef3456abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx5678"
                  endpoint="peer3.example.com"
                  latestHandshake="N/A"
                  transferRx="0 B"
                  transferTx="0 B"
                />
                <WGTableRow
                  rowKey="peer4"
                  status="connected"
                  name="gluon-worker-03"
                  pubKey="efgh5678ijkl9012mnop3456qrst7890uvwx5678yzab9012cdef3456abcd1234"
                  endpoint="peer4.example.com"
                  latestHandshake="2024-10-01 12:35:10"
                  transferRx="2.5 GB"
                  transferTx="1.5 GB"
                />
             </Table>
        </CardContainer>
        <CardContainer title="OSPF Neighbors" noPadding={true} icon={<Router className="w-5 h-5"/>}>
            <Table
                columns={[
                  'Router ID', 'Area', 'State', 'Interface', 'Hello Timer', 'Dead Timer', 'Cost', 'Priority'
                ]}
            >
                <OSPFTableRow
                  rowKey="ospf1"
                  routerId="1.1.1.1"
                  area="0.0.0.0"
                  state="Full"
                  interface="eth0"
                  helloTimer="10s"
                  deadTimer="40s"
                  cost={10}
                  priority={100}
                />
                <OSPFTableRow
                  rowKey="ospf2"
                  routerId="1.1.1.2"
                  area="0.0.0.0"
                  state="Down"
                  interface="eth1"
                  helloTimer="10s"
                  deadTimer="40s"
                  cost={20}
                  priority={90}
                />
                <OSPFTableRow
                  rowKey="ospf3"
                  routerId="1.1.1.3"
                  area="0.0.0.0"
                  state="Full"
                  interface="eth2"
                  helloTimer="10s"
                  deadTimer="40s"
                  cost={30}
                  priority={80}
                />
                <OSPFTableRow
                  rowKey="ospf4"
                  routerId="1.1.1.4"
                  area="0.0.0.0"
                  state="Full"
                  interface="eth3"
                  helloTimer="10s"
                  deadTimer="40s"
                  cost={40}
                  priority={70}
                />
                <OSPFTableRow
                    rowKey="ospf5"
                    routerId="1.1.1.5"
                    area="0.0.0.0"
                    state="Full"
                    interface="eth4"
                    helloTimer="10s"
                    deadTimer="40s"
                    cost={50}
                    priority={60}
                  />
                <OSPFTableRow
                    rowKey="ospf6"
                    routerId="1.1.1.6"
                    area="0.0.0.0"
                    state="Full"
                    interface="eth5"
                    helloTimer="10s"
                    deadTimer="40s"
                    cost={60}
                    priority={50}
                  />
            </Table>
        </CardContainer>
      </div>
    </>
  )
}

export default NetworkingView
