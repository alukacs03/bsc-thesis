package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	nodesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_total",
			Help: "Total nodes grouped by role and status.",
		},
		[]string{"role", "status"},
	)

	// WireGuard metrics
	wireguardHandshakeAge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_wireguard_handshake_age_seconds",
			Help: "Seconds since last WireGuard handshake per peer.",
		},
		[]string{"node_id", "interface", "peer_node_id"},
	)
	wireguardRxBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_wireguard_rx_bytes_total",
			Help: "Total bytes received per WireGuard peer.",
		},
		[]string{"node_id", "interface", "peer_node_id"},
	)
	wireguardTxBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_wireguard_tx_bytes_total",
			Help: "Total bytes transmitted per WireGuard peer.",
		},
		[]string{"node_id", "interface", "peer_node_id"},
	)
	wireguardPeersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_wireguard_peers_total",
			Help: "Total WireGuard peers per node.",
		},
		[]string{"node_id", "status"},
	)
	wireguardStaleHandshakes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gluon_wireguard_stale_handshakes_total",
			Help: "Total WireGuard peers with stale handshakes (>3 minutes).",
		},
	)

	// OSPF metrics
	ospfNeighborsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_ospf_neighbors_total",
			Help: "Total OSPF neighbors grouped by state.",
		},
		[]string{"state"},
	)
	ospfNeighborsByNode = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_ospf_neighbors_by_node",
			Help: "OSPF neighbor count per node.",
		},
		[]string{"node_id", "state"},
	)

	// Config sync metrics
	configVersionByNode = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_config_version",
			Help: "Current config version per node.",
		},
		[]string{"node_id"},
	)
	configApplySuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gluon_config_apply_success_total",
			Help: "Total successful config applications.",
		},
	)
	configApplyFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gluon_config_apply_failure_total",
			Help: "Total failed config applications.",
		},
	)

	// Enrollment metrics
	enrollmentRequestsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_enrollment_requests_total",
			Help: "Total enrollment requests grouped by status.",
		},
		[]string{"status"},
	)
	nodesOfflineTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_offline_total",
			Help: "Total offline nodes grouped by role.",
		},
		[]string{"role"},
	)
	nodesK8sStateTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_k8s_state_total",
			Help: "Total nodes grouped by Kubernetes state.",
		},
		[]string{"state"},
	)
	nodesByProviderTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_by_provider_total",
			Help: "Total nodes grouped by provider, role, and status.",
		},
		[]string{"provider", "role", "status"},
	)
	nodesLastSeenMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_last_seen_seconds_max",
			Help: "Max seconds since last heartbeat for nodes grouped by role.",
		},
		[]string{"role"},
	)
	nodesLastSeenAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_last_seen_seconds_avg",
			Help: "Avg seconds since last heartbeat for nodes grouped by role.",
		},
		[]string{"role"},
	)
	nodesCPUUsageAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_cpu_usage_avg",
			Help: "Average CPU usage percentage grouped by role.",
		},
		[]string{"role"},
	)
	nodesMemoryUsageAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_memory_usage_avg",
			Help: "Average memory usage percentage grouped by role.",
		},
		[]string{"role"},
	)
	nodesDiskUsageAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_disk_usage_avg",
			Help: "Average disk usage percentage grouped by role.",
		},
		[]string{"role"},
	)
	nodesUptimeSecondsAvg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_nodes_uptime_seconds_avg",
			Help: "Average uptime seconds grouped by role.",
		},
		[]string{"role"},
	)
	commandsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_commands_total",
			Help: "Total node commands grouped by status.",
		},
		[]string{"status"},
	)
	commandsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gluon_commands_in_flight",
			Help: "Total node commands currently pending or running.",
		},
	)
	k8sPodsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_k8s_pods_total",
			Help: "Total Kubernetes pods grouped by namespace and phase.",
		},
		[]string{"namespace", "phase"},
	)
	k8sWorkloadsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gluon_k8s_workloads_total",
			Help: "Total Kubernetes workloads grouped by namespace and kind.",
		},
		[]string{"namespace", "kind"},
	)
)

func init() {
	prometheus.MustRegister(
		nodesTotal,
		nodesOfflineTotal,
		nodesK8sStateTotal,
		nodesByProviderTotal,
		nodesLastSeenMax,
		nodesLastSeenAvg,
		nodesCPUUsageAvg,
		nodesMemoryUsageAvg,
		nodesDiskUsageAvg,
		nodesUptimeSecondsAvg,
		commandsTotal,
		commandsInFlight,
		k8sPodsTotal,
		k8sWorkloadsTotal,
		// WireGuard metrics
		wireguardHandshakeAge,
		wireguardRxBytesTotal,
		wireguardTxBytesTotal,
		wireguardPeersTotal,
		wireguardStaleHandshakes,
		// OSPF metrics
		ospfNeighborsTotal,
		ospfNeighborsByNode,
		// Config metrics
		configVersionByNode,
		configApplySuccess,
		configApplyFailure,
		// Enrollment metrics
		enrollmentRequestsTotal,
	)
}

type nodeCountRow struct {
	Role   string
	Status string
	Count  int64
}

type nodeProviderCountRow struct {
	Provider string
	Role     string
	Status   string
	Count    int64
}

type k8sStateCountRow struct {
	State string
	Count int64
}

type commandCountRow struct {
	Status string
	Count  int64
}

type nodeLastSeenRow struct {
	Role       string
	LastSeenAt *time.Time
}

type nodeUsageRow struct {
	Role        string
	CPUUsage    *float64
	MemoryUsage *float64
	DiskUsage   *float64
	Uptime      *uint64
}

type k8sList[T any] struct {
	Items []T `json:"items"`
}

type k8sPod struct {
	Metadata struct {
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

type k8sWorkload struct {
	Metadata struct {
		Namespace string `json:"namespace"`
	} `json:"metadata"`
}

func StartDatabaseMetrics(interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			updateNodeCounts()
			updateProviderCounts()
			updateK8sStateCounts()
			updateCommandCounts()
			updateLastSeenMetrics()
			updateUsageMetrics()
			updateKubernetesMetrics()
			// New metrics
			updateWireGuardMetrics()
			updateOSPFMetrics()
			updateConfigMetrics()
			updateEnrollmentMetrics()
		}
	}()
}

func updateNodeCounts() {
	var rows []nodeCountRow
	err := database.DB.
		Model(&models.Node{}).
		Select("role, status, count(*) as count").
		Group("role, status").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect node counts", "error", err)
		return
	}

	nodesTotal.Reset()
	nodesOfflineTotal.Reset()
	for _, row := range rows {
		nodesTotal.WithLabelValues(row.Role, row.Status).Set(float64(row.Count))
		if row.Status == string(models.NodeStatusOffline) {
			nodesOfflineTotal.WithLabelValues(row.Role).Set(float64(row.Count))
		}
	}
}

func updateProviderCounts() {
	var rows []nodeProviderCountRow
	err := database.DB.
		Model(&models.Node{}).
		Select("provider, role, status, count(*) as count").
		Group("provider, role, status").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect provider counts", "error", err)
		return
	}

	nodesByProviderTotal.Reset()
	for _, row := range rows {
		nodesByProviderTotal.WithLabelValues(row.Provider, row.Role, row.Status).Set(float64(row.Count))
	}
}

func updateK8sStateCounts() {
	var rows []k8sStateCountRow
	err := database.DB.
		Model(&models.Node{}).
		Select("k8s_state as state, count(*) as count").
		Group("k8s_state").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect k8s state counts", "error", err)
		return
	}

	nodesK8sStateTotal.Reset()
	for _, row := range rows {
		state := row.State
		if state == "" {
			state = "unknown"
		}
		nodesK8sStateTotal.WithLabelValues(state).Set(float64(row.Count))
	}
}

func updateCommandCounts() {
	var rows []commandCountRow
	err := database.DB.
		Model(&models.NodeCommand{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect command counts", "error", err)
		return
	}

	commandsTotal.Reset()
	commandsInFlight.Set(0)
	var inFlight float64
	for _, row := range rows {
		commandsTotal.WithLabelValues(row.Status).Set(float64(row.Count))
		if row.Status == string(models.NodeCommandStatusPending) || row.Status == string(models.NodeCommandStatusRunning) {
			inFlight += float64(row.Count)
		}
	}
	commandsInFlight.Set(inFlight)
}

func updateLastSeenMetrics() {
	var rows []nodeLastSeenRow
	err := database.DB.
		Model(&models.Node{}).
		Select("role, last_seen_at").
		Where("last_seen_at IS NOT NULL").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect last seen timestamps", "error", err)
		return
	}

	type agg struct {
		max float64
		sum float64
		n   int
	}

	byRole := map[string]*agg{}
	now := time.Now()
	for _, row := range rows {
		if row.LastSeenAt == nil {
			continue
		}
		age := now.Sub(*row.LastSeenAt).Seconds()
		entry := byRole[row.Role]
		if entry == nil {
			entry = &agg{}
			byRole[row.Role] = entry
		}
		if age > entry.max {
			entry.max = age
		}
		entry.sum += age
		entry.n++
	}

	nodesLastSeenMax.Reset()
	nodesLastSeenAvg.Reset()
	for role, entry := range byRole {
		nodesLastSeenMax.WithLabelValues(role).Set(entry.max)
		avg := 0.0
		if entry.n > 0 {
			avg = entry.sum / float64(entry.n)
		}
		nodesLastSeenAvg.WithLabelValues(role).Set(avg)
	}
}

func updateUsageMetrics() {
	var rows []nodeUsageRow
	err := database.DB.
		Model(&models.Node{}).
		Select("role, cpu_usage, memory_usage, disk_usage, uptime_seconds").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect usage metrics", "error", err)
		return
	}

	type agg struct {
		sum float64
		n   int
	}

	cpuByRole := map[string]*agg{}
	memByRole := map[string]*agg{}
	diskByRole := map[string]*agg{}
	uptimeByRole := map[string]*agg{}

	for _, row := range rows {
		role := row.Role
		if row.CPUUsage != nil {
			entry := cpuByRole[role]
			if entry == nil {
				entry = &agg{}
				cpuByRole[role] = entry
			}
			entry.sum += *row.CPUUsage
			entry.n++
		}
		if row.MemoryUsage != nil {
			entry := memByRole[role]
			if entry == nil {
				entry = &agg{}
				memByRole[role] = entry
			}
			entry.sum += *row.MemoryUsage
			entry.n++
		}
		if row.DiskUsage != nil {
			entry := diskByRole[role]
			if entry == nil {
				entry = &agg{}
				diskByRole[role] = entry
			}
			entry.sum += *row.DiskUsage
			entry.n++
		}
		if row.Uptime != nil {
			entry := uptimeByRole[role]
			if entry == nil {
				entry = &agg{}
				uptimeByRole[role] = entry
			}
			entry.sum += float64(*row.Uptime)
			entry.n++
		}
	}

	nodesCPUUsageAvg.Reset()
	for role, entry := range cpuByRole {
		if entry.n == 0 {
			continue
		}
		nodesCPUUsageAvg.WithLabelValues(role).Set(entry.sum / float64(entry.n))
	}

	nodesMemoryUsageAvg.Reset()
	for role, entry := range memByRole {
		if entry.n == 0 {
			continue
		}
		nodesMemoryUsageAvg.WithLabelValues(role).Set(entry.sum / float64(entry.n))
	}

	nodesDiskUsageAvg.Reset()
	for role, entry := range diskByRole {
		if entry.n == 0 {
			continue
		}
		nodesDiskUsageAvg.WithLabelValues(role).Set(entry.sum / float64(entry.n))
	}

	nodesUptimeSecondsAvg.Reset()
	for role, entry := range uptimeByRole {
		if entry.n == 0 {
			continue
		}
		nodesUptimeSecondsAvg.WithLabelValues(role).Set(entry.sum / float64(entry.n))
	}
}

func updateKubernetesMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	podBytes, err := kubectlJSON(ctx, []string{"get", "pods", "-A", "-o", "json"})
	if err != nil {
		logger.Error("metrics: failed to collect k8s pods", "error", err)
	} else {
		var pods k8sList[k8sPod]
		if err := json.Unmarshal(podBytes, &pods); err != nil {
			logger.Error("metrics: failed to parse k8s pods", "error", err)
		} else {
			k8sPodsTotal.Reset()
			for _, pod := range pods.Items {
				ns := pod.Metadata.Namespace
				phase := pod.Status.Phase
				if phase == "" {
					phase = "Unknown"
				}
				k8sPodsTotal.WithLabelValues(ns, phase).Inc()
			}
		}
	}

	workloadKinds := []string{"deployments", "daemonsets", "statefulsets", "jobs"}
	k8sWorkloadsTotal.Reset()
	for _, kind := range workloadKinds {
		payload, err := kubectlJSON(ctx, []string{"get", kind, "-A", "-o", "json"})
		if err != nil {
			continue
		}
		var items k8sList[k8sWorkload]
		if err := json.Unmarshal(payload, &items); err != nil {
			continue
		}
		for _, item := range items.Items {
			ns := item.Metadata.Namespace
			if ns == "" {
				ns = "default"
			}
			k8sWorkloadsTotal.WithLabelValues(ns, kind).Inc()
		}
	}
}

func kubectlJSON(ctx context.Context, args []string) ([]byte, error) {
	kubeconfig := strings.TrimSpace(os.Getenv("GLUON_KUBECONFIG"))
	if kubeconfig == "" {
		kubeconfig = "/etc/kubernetes/admin.conf"
	}
	if _, err := os.Stat(kubeconfig); err != nil {
		return nil, fmt.Errorf("kubeconfig not found at %s", kubeconfig)
	}
	cmdArgs := append([]string{"--kubeconfig", kubeconfig}, args...)
	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("kubectl failed: %s", msg)
	}
	return out, nil
}

// WireGuard peer row for metrics collection
type wireguardPeerRow struct {
	NodeID         uint
	InterfaceName  string
	PeerNodeID     uint
	LastHandshakeAt *time.Time
	RxBytes        uint64
	TxBytes        uint64
	Status         string
}

func updateWireGuardMetrics() {
	var rows []wireguardPeerRow
	err := database.DB.
		Table("node_peers").
		Select("wire_guard_interfaces.node_id as node_id, wire_guard_interfaces.name as interface_name, node_peers.peer_node_id, node_peers.last_handshake_at, node_peers.rx_bytes, node_peers.tx_bytes, node_peers.status").
		Joins("JOIN wire_guard_interfaces ON node_peers.interface_id = wire_guard_interfaces.id").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect wireguard peer data", "error", err)
		return
	}

	wireguardHandshakeAge.Reset()
	wireguardRxBytesTotal.Reset()
	wireguardTxBytesTotal.Reset()
	wireguardPeersTotal.Reset()

	now := time.Now()
	staleCount := 0.0
	peerCounts := map[string]map[string]int{} // node_id -> status -> count

	for _, row := range rows {
		nodeIDStr := fmt.Sprintf("%d", row.NodeID)
		peerNodeIDStr := fmt.Sprintf("%d", row.PeerNodeID)

		// Handshake age
		if row.LastHandshakeAt != nil {
			age := now.Sub(*row.LastHandshakeAt).Seconds()
			wireguardHandshakeAge.WithLabelValues(nodeIDStr, row.InterfaceName, peerNodeIDStr).Set(age)
			if age > 180 { // 3 minutes
				staleCount++
			}
		}

		// Bytes
		wireguardRxBytesTotal.WithLabelValues(nodeIDStr, row.InterfaceName, peerNodeIDStr).Set(float64(row.RxBytes))
		wireguardTxBytesTotal.WithLabelValues(nodeIDStr, row.InterfaceName, peerNodeIDStr).Set(float64(row.TxBytes))

		// Peer counts
		if peerCounts[nodeIDStr] == nil {
			peerCounts[nodeIDStr] = map[string]int{}
		}
		peerCounts[nodeIDStr][row.Status]++
	}

	wireguardStaleHandshakes.Set(staleCount)

	for nodeID, statusCounts := range peerCounts {
		for status, count := range statusCounts {
			wireguardPeersTotal.WithLabelValues(nodeID, status).Set(float64(count))
		}
	}
}

// OSPF neighbor info stored as JSON on node
type ospfNeighborInfo struct {
	NeighborID string `json:"neighbor_id"`
	State      string `json:"state"`
	Interface  string `json:"interface"`
}

func updateOSPFMetrics() {
	var nodes []models.Node
	err := database.DB.
		Select("id, ospf_neighbors").
		Where("ospf_neighbors IS NOT NULL").
		Find(&nodes).Error
	if err != nil {
		logger.Error("metrics: failed to collect OSPF data", "error", err)
		return
	}

	ospfNeighborsTotal.Reset()
	ospfNeighborsByNode.Reset()

	totalByState := map[string]int{}

	for _, node := range nodes {
		if node.OSPFNeighbors == nil {
			continue
		}

		var neighbors []ospfNeighborInfo
		if err := json.Unmarshal(node.OSPFNeighbors, &neighbors); err != nil {
			continue
		}

		nodeIDStr := fmt.Sprintf("%d", node.ID)
		nodeStateCount := map[string]int{}

		for _, neighbor := range neighbors {
			state := neighbor.State
			if state == "" {
				state = "unknown"
			}
			totalByState[state]++
			nodeStateCount[state]++
		}

		for state, count := range nodeStateCount {
			ospfNeighborsByNode.WithLabelValues(nodeIDStr, state).Set(float64(count))
		}
	}

	for state, count := range totalByState {
		ospfNeighborsTotal.WithLabelValues(state).Set(float64(count))
	}
}

type configVersionRow struct {
	NodeID         uint
	AppliedVersion int
}

func updateConfigMetrics() {
	var rows []configVersionRow
	err := database.DB.
		Table("nodes").
		Select("nodes.id as node_id, COALESCE(applied_configs.version, 0) as applied_version").
		Joins("LEFT JOIN (SELECT node_id, MAX(version) as version FROM node_configs WHERE applied_at IS NOT NULL GROUP BY node_id) AS applied_configs ON nodes.id = applied_configs.node_id").
		Scan(&rows).Error
	if err != nil {
		// Fallback: just use zero for all nodes
		var nodes []models.Node
		if err := database.DB.Select("id").Find(&nodes).Error; err == nil {
			configVersionByNode.Reset()
			for _, n := range nodes {
				configVersionByNode.WithLabelValues(fmt.Sprintf("%d", n.ID)).Set(0)
			}
		}
		return
	}

	configVersionByNode.Reset()
	for _, row := range rows {
		configVersionByNode.WithLabelValues(fmt.Sprintf("%d", row.NodeID)).Set(float64(row.AppliedVersion))
	}
}

type enrollmentCountRow struct {
	Status string
	Count  int64
}

func updateEnrollmentMetrics() {
	var rows []enrollmentCountRow
	err := database.DB.
		Model(&models.NodeEnrollmentRequest{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error
	if err != nil {
		logger.Error("metrics: failed to collect enrollment counts", "error", err)
		return
	}

	enrollmentRequestsTotal.Reset()
	for _, row := range rows {
		enrollmentRequestsTotal.WithLabelValues(row.Status).Set(float64(row.Count))
	}
}

// IncrementConfigApplySuccess increments the config apply success counter
func IncrementConfigApplySuccess() {
	configApplySuccess.Inc()
}

// IncrementConfigApplyFailure increments the config apply failure counter
func IncrementConfigApplyFailure() {
	configApplyFailure.Inc()
}
