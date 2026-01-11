package metrics

import (
	"context"
	"encoding/json"
	"os/exec"
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
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}
