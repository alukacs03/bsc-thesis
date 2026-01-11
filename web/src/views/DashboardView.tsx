import CardWithIcon from "../components/CardWithIcon"
import CardContainer from "../components/CardContainer"
import CardAction from "../components/CardAction"
import CardActivity from "../components/CardActivity"
import { Checkmark, ClockMark, MetricsMark, XMark } from "../components/Icons"
import Table from "../components/Table"
import NodeTableRow from "../components/NodeTableRow"
import { AlertTriangle } from "lucide-react"
import { useNodes } from "../services/hooks/useNodes"
import { useEnrollments } from "../services/hooks/useEnrollments"
import { LoadingSpinner } from "../components/LoadingSpinner"
import { ErrorMessage } from "../components/ErrorMessage"
import { useNavigate } from "react-router-dom"

const DashboardView = () => {
  const navigate = useNavigate()
  const { data: nodes, loading: nodesLoading, error: nodesError, refetch: refetchNodes } = useNodes({ pollingInterval: 15000 })
  const { data: enrollments, loading: enrollmentsLoading, error: enrollmentsError, refetch: refetchEnrollments } = useEnrollments()

  if ((nodesLoading && !nodes) || (enrollmentsLoading && !enrollments)) {
    return <LoadingSpinner />
  }

  if (nodesError || enrollmentsError) {
    return (
      <ErrorMessage
        message={(nodesError ?? enrollmentsError)!.message}
        onRetry={() => {
          refetchNodes()
          refetchEnrollments()
        }}
      />
    )
  }

  const nodesList = nodes ?? []
  const enrollmentsList = enrollments ?? []

  const onlineNodes = nodesList.filter((node) => node.status === "active").length
  const offlineNodes = nodesList.filter((node) => node.status === "offline" || node.status === "decommissioned").length
  const maintenanceNodes = nodesList.filter((node) => node.status === "maintenance").length
  const pendingApprovals = enrollmentsList.filter((request) => request.status === "pending").length
  const activeAlerts = offlineNodes + maintenanceNodes

  const toTime = (value?: string) => (value ? new Date(value).getTime() : 0)
  const formatRelativeTime = (value?: string) => {
    if (!value) return "Never"
    const diff = Date.now() - new Date(value).getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 1) return "Just now"
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    return `${Math.floor(hours / 24)}d ago`
  }

  const nodeTimestamp = (node: typeof nodesList[number]) =>
    node.last_seen_at ?? node.updated_at ?? node.created_at

  type ActivityItem = {
    message: string
    timestamp: string
    alertLevel: "info" | "warning" | "error"
  }

  const activityItems: ActivityItem[] = []

  const pendingActivity = enrollmentsList
    .filter((request) => request.status === "pending")
    .sort((a, b) => toTime(b.requested_at) - toTime(a.requested_at))
    .map((request) => ({
      message: `Enrollment pending: ${request.hostname}`,
      timestamp: formatRelativeTime(request.requested_at),
      alertLevel: "warning" as const,
    }))

  const offlineActivity = nodesList
    .filter((node) => node.status === "offline" || node.status === "decommissioned")
    .sort((a, b) => toTime(nodeTimestamp(b)) - toTime(nodeTimestamp(a)))
    .map((node) => ({
      message: `Node ${node.hostname} is offline`,
      timestamp: formatRelativeTime(nodeTimestamp(node)),
      alertLevel: "error" as const,
    }))

  const maintenanceActivity = nodesList
    .filter((node) => node.status === "maintenance")
    .sort((a, b) => toTime(nodeTimestamp(b)) - toTime(nodeTimestamp(a)))
    .map((node) => ({
      message: `Node ${node.hostname} in maintenance`,
      timestamp: formatRelativeTime(nodeTimestamp(node)),
      alertLevel: "warning" as const,
    }))

  const onlineActivity = nodesList
    .filter((node) => node.status === "active")
    .sort((a, b) => toTime(nodeTimestamp(b)) - toTime(nodeTimestamp(a)))
    .map((node) => ({
      message: `Node ${node.hostname} online`,
      timestamp: formatRelativeTime(nodeTimestamp(node)),
      alertLevel: "info" as const,
    }))

  const addActivity = (items: ActivityItem[]) => {
    for (const item of items) {
      if (activityItems.length >= 4) {
        break
      }
      activityItems.push(item)
    }
  }

  addActivity(pendingActivity)
  addActivity(offlineActivity)
  addActivity(maintenanceActivity)
  addActivity(onlineActivity)

  const nodeRows = [...nodesList]
    .sort((a, b) => toTime(nodeTimestamp(b)) - toTime(nodeTimestamp(a)))
    .slice(0, 5)

  const statusToBadge = (status: typeof nodesList[number]["status"]) => {
    if (status === "active") return "online"
    if (status === "offline" || status === "decommissioned") return "offline"
    return "degraded"
  }

  return (
    <div className="space-y-4 md:space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
            <CardWithIcon
                title="Nodes Online"
                value={onlineNodes}
                hint="Click to view nodes"
                onClick={() => navigate("/nodes")}
                textColorClass="text-slate-600"
                valueColorClass="text-green-600"
                iconBGColorClass="bg-green-200"
                icon={<Checkmark />}
            />
            <CardWithIcon
                title="Nodes Offline"
                value={offlineNodes}
                hint="Click to view nodes"
                onClick={() => navigate("/nodes")}
                textColorClass="text-slate-600"
                valueColorClass="text-red-600"
                iconBGColorClass="bg-red-200"
                icon={<XMark />}
            />
            <CardWithIcon
                title="Pending Approvals"
                value={pendingApprovals}
                hint="Click to review approvals"
                onClick={() => navigate("/approvals")}
                textColorClass="text-slate-600"
                outlineColorClass={pendingApprovals > 0 ? "ring-2 ring-red-500" : ""}
                iconBGColorClass="bg-red-200"
                icon={<ClockMark />}               
            />
            <CardWithIcon
                title="Active Alerts"
                value={activeAlerts}
                hint="Click to view system health"
                onClick={() => navigate("/nodes")}
                textColorClass="text-slate-600"
                valueColorClass="text-yellow-600"
                iconBGColorClass="bg-yellow-200"
                icon={<AlertTriangle />}
            />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
                <CardContainer title="Recent Activity">
                    {activityItems.length === 0 ? (
                        <CardActivity message="No recent activity" timestamp="Just now" alertLevel="info" />
                    ) : (
                        activityItems.map((item, index) => (
                            <CardActivity
                                key={`${item.message}-${index}`}
                                message={item.message}
                                timestamp={item.timestamp}
                                alertLevel={item.alertLevel}
                            />
                        ))
                    )}
                </CardContainer>
                <CardContainer title="Quick Actions">
                    <CardAction 
                        title = "Add New Node"
                        subtitle = "Register a new VPS node"
                        onClick = {() => navigate("/approvals")}
                        icon={
                            <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                            </svg>
                        }
                    />
                    <CardAction 
                        title = "Review Approvals"
                        subtitle = {`${pendingApprovals} pending approvals`}
                        onClick = {() => navigate("/approvals")}
                        icon={<Checkmark />}
                    />
                    <CardAction 
                        title = "View Metrics"
                        subtitle = "System performance metrics"
                        onClick = {() => navigate("/kubernetes")}
                        icon={<MetricsMark />}
                    />
                </CardContainer>
        </div>
        <CardContainer title="Node Overview" noPadding={true}>
            <Table
                columns={['Node Name', 'IP Address', 'Status', 'Role', 'Last Seen']}
            >
                {nodeRows.length === 0 ? (
                    <tr>
                        <td className="py-4 px-6 text-sm text-slate-500" colSpan={5}>No nodes found</td>
                    </tr>
                ) : (
                    nodeRows.map((node) => (
                        <NodeTableRow
                            key={node.id}
                            handleNodeClick={(nodeId) => navigate(`/nodes/${nodeId}`)}
                            nodeId={node.id}
                            nodeName={node.hostname}
                            nodeIP={node.management_ip || node.public_ip}
                            nodeStatus={statusToBadge(node.status)}
                            nodeRole={node.role === "hub" ? "Hub" : "Worker"}
                            lastSeen={formatRelativeTime(nodeTimestamp(node))}
                        />
                    ))
                )}
            </Table>
        </CardContainer>
    </div>
  )
}

export default DashboardView
