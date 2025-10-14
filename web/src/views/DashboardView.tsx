import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import CardAction from "../components/CardAction"
import CardActivity from "../components/CardActivity"
import { Checkmark, ClockMark, MetricsMark, XMark } from "../components/Icons"
import Table from "../components/Table"
import NodeTableRow from "../components/NodeTableRow"
import { AlertTriangle } from "lucide-react"

const DashboardView = () => {
  return (
    <div className="space-y-4 md:space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
            <CardWithIcon
                title="Nodes Online"
                value="4"
                hint="Click to view nodes"
                onClick={() => alert('Navigate to Nodes')}
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-200"
                icon={<Checkmark />}
            />
            <CardWithIcon
                title="Nodes Offline"
                value="1"
                hint="Click to view nodes"
                onClick={() => alert('Navigate to Nodes')}
                textColorClass="text-slate-600"
                valueColorClass="text-red-600"
                iconBGColorClass="bg-red-200"
                icon={<XMark />}
            />
            <CardWithIcon
                title="Pending Approvals"
                value="3"
                hint="Click to review approvals"
                onClick={() => alert('Navigate to Approvals')}
                textColorClass="text-slate-600"
                outlineColorClass="ring-2 ring-red-500"
                iconBGColorClass="bg-red-200"
                icon={<ClockMark />}               
            />
            <CardWithIcon
                title="Active Alerts"
                value="2"
                hint="Click to view system health"
                onClick={() => alert('Navigate to System Health')}
                textColorClass="text-slate-600"
                valueColorClass="text-yellow-600"
                iconBGColorClass="bg-yellow-200"
                icon={<AlertTriangle />}
            />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
                <CardContainer title="Recent Activity">
                    <CardActivity 
                        message="Node gluon-worker-03 went offline"
                        timestamp="2 hours ago"
                        alertLevel="error"
                    />
                    <CardActivity
                        message="New node gluon-worker-04 registered"
                        timestamp="5 hours ago"
                        alertLevel="info"
                    />
                    <CardActivity
                        message="3 new approvals pending review"
                        timestamp="1 day ago"
                        alertLevel="warning"
                    />
                    <CardActivity
                        message="System health check passed"
                        timestamp="2 days ago"
                        alertLevel="info"
                    />
                </CardContainer>
                <CardContainer title="Quick Actions">
                    <CardAction 
                        title = "Add New Node"
                        subtitle = "Register a new VPS node"
                        onClick = {() => alert('Navigate to Add Node')}
                        icon={
                            <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                            </svg>
                        }
                    />
                    <CardAction 
                        title = "Review Approvals"
                        subtitle = "3 pending approvals"
                        onClick = {() => alert('Navigate to Approvals')}
                        icon={<Checkmark />}
                    />
                    <CardAction 
                        title = "View Metrics"
                        subtitle = "System performance metrics"
                        onClick = {() => alert('Navigate to System Health')}
                        icon={<MetricsMark />}
                    />
                </CardContainer>
        </div>
        <CardContainer title="Node Overview" noPadding={true}>
            <Table
                columns={['Node Name', 'IP Address', 'Status', 'Role', 'Last Heartbeat']}
            >
                <NodeTableRow
                    handleNodeClick={() => alert('Node gluon-worker-01 clicked')}
                    nodeName="gluon-worker-01"
                    nodeIP="10.0.1.10"
                    nodeStatus="online"
                    nodeRole="Control Plane"
                    lastHeartbeat="2 minutes ago"
                />
                <NodeTableRow
                    handleNodeClick={() => alert('Node gluon-worker-02 clicked')}
                    nodeName="gluon-worker-02"
                    nodeIP="10.0.1.11"
                    nodeStatus="offline"
                    nodeRole="Worker"
                    lastHeartbeat="5 minutes ago"
                />
                <NodeTableRow
                    handleNodeClick={() => alert('Node gluon-worker-03 clicked')}
                    nodeName="gluon-worker-03"
                    nodeIP="10.0.1.12"
                    nodeStatus="degraded"
                    nodeRole="Worker"
                    lastHeartbeat="10 minutes ago"
                />
                <NodeTableRow
                    handleNodeClick={() => alert('Node gluon-worker-04 clicked')}
                    nodeName="gluon-worker-04"
                    nodeIP="10.0.1.13"
                    nodeStatus="offline"
                    nodeRole="Worker"
                    lastHeartbeat="5 minutes ago"
                />
            </Table>
        </CardContainer>
    </div>
  )
}

export default DashboardView
