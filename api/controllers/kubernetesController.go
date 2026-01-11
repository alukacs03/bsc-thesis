package controllers

import (
	"errors"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"gluon-api/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	defaultPodCIDR     = "10.244.0.0/16"
	defaultServiceCIDR = services.DefaultKubernetesServiceCIDR
	defaultK8sVersion  = "v1.29"
	joinRefreshWindow  = 30 * time.Minute
)

type kubernetesTask struct {
	Action               string `json:"action"`
	ControlPlaneEndpoint string `json:"control_plane_endpoint,omitempty"`
	PodCIDR              string `json:"pod_cidr,omitempty"`
	ServiceCIDR          string `json:"service_cidr,omitempty"`
	KubernetesVersion    string `json:"kubernetes_version,omitempty"`
	JoinCommand          string `json:"join_command,omitempty"`
	Note                 string `json:"note,omitempty"`
}

type kubernetesReport struct {
	State string `json:"state"`

	Message string `json:"message,omitempty"`

	ControlPlaneEndpoint string `json:"control_plane_endpoint,omitempty"`
	PodCIDR              string `json:"pod_cidr,omitempty"`
	ServiceCIDR          string `json:"service_cidr,omitempty"`
	KubernetesVersion    string `json:"kubernetes_version,omitempty"`

	WorkerJoinCommand       string `json:"worker_join_command,omitempty"`
	ControlPlaneJoinCommand string `json:"control_plane_join_command,omitempty"`
	JoinCommandExpiresAt    string `json:"join_command_expires_at,omitempty"` 
}

func GetKubernetesTask(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var bootstrapHub models.Node
	if err := database.DB.
		Where("role = ? OR reported_desired_role = ?", models.NodeRoleHub, string(models.NodeRoleHub)).
		Order("id asc").
		First(&bootstrapHub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(kubernetesTask{Action: "none", Note: "No hubs enrolled yet"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to determine bootstrap hub"})
	}

	cluster, err := getOrCreateKubernetesCluster(&bootstrapHub)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load cluster state"})
	}

	if cluster.InitializedAt == nil {
		if cluster.BootstrapNodeID != nil && node.ID == *cluster.BootstrapNodeID {
			endpoint := cluster.ControlPlaneEndpoint
			if endpoint == "" {
				if ip, err := services.GetNodeLoopbackIP(bootstrapHub.ID); err == nil && ip != "" {
					endpoint = ip
				} else {
					endpoint = bootstrapHub.PublicIP
				}
			}
			return c.JSON(kubernetesTask{
				Action:               "init",
				ControlPlaneEndpoint: endpoint,
				PodCIDR:              nonEmpty(cluster.PodCIDR, defaultPodCIDR),
				ServiceCIDR:          nonEmpty(cluster.ServiceCIDR, defaultServiceCIDR),
				KubernetesVersion:    nonEmpty(cluster.KubernetesVersion, defaultK8sVersion),
				Note:                 "Bootstrap control-plane with kubeadm init (single-cluster mode)",
			})
		}
		return c.JSON(kubernetesTask{Action: "wait", Note: "Waiting for bootstrap hub to initialize the cluster"})
	}

	
	if wantsControlPlane(&node) {
		isBootstrap := cluster.BootstrapNodeID != nil && node.ID == *cluster.BootstrapNodeID
		if isBootstrap && shouldRefreshJoinCommands(cluster) {
			endpoint := cluster.ControlPlaneEndpoint
			if endpoint == "" {
				if ip, err := services.GetNodeLoopbackIP(bootstrapHub.ID); err == nil && ip != "" {
					endpoint = ip
				} else {
					endpoint = bootstrapHub.PublicIP
				}
			}
			return c.JSON(kubernetesTask{
				Action:               "init",
				ControlPlaneEndpoint: endpoint,
				PodCIDR:              nonEmpty(cluster.PodCIDR, defaultPodCIDR),
				ServiceCIDR:          nonEmpty(cluster.ServiceCIDR, defaultServiceCIDR),
				KubernetesVersion:    nonEmpty(cluster.KubernetesVersion, defaultK8sVersion),
				Note:                 "Refresh Kubernetes join commands",
			})
		}

		if node.K8sState == "joined_control_plane" || (isBootstrap && node.K8sState == "cluster_initialized") {
			return c.JSON(kubernetesTask{Action: "none"})
		}
		if cluster.ControlPlaneJoinCommand == "" {
			return c.JSON(kubernetesTask{Action: "wait", Note: "Cluster initialized; join command not available yet"})
		}
		return c.JSON(kubernetesTask{
			Action:      "join_control_plane",
			JoinCommand: cluster.ControlPlaneJoinCommand,
			Note:        "Join as additional control-plane node",
		})
	}

	switch node.Role {
	case models.NodeRoleWorker:
		if node.K8sState == "joined_worker" {
			return c.JSON(kubernetesTask{Action: "none"})
		}
		if cluster.WorkerJoinCommand == "" {
			return c.JSON(kubernetesTask{Action: "wait", Note: "Cluster initialized; join command not available yet"})
		}
		return c.JSON(kubernetesTask{
			Action:      "join_worker",
			JoinCommand: cluster.WorkerJoinCommand,
			Note:        "Join as worker node",
		})
	default:
		return c.JSON(kubernetesTask{Action: "none"})
	}
}

func ReportKubernetes(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var input kubernetesReport
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	now := time.Now()
	updates := map[string]any{
		"k8s_last_attempt_at": &now,
	}

	switch input.State {
	case "cluster_initialized":
		updates["k8s_state"] = "cluster_initialized"
		updates["k8s_joined_at"] = &now
		updates["k8s_last_error"] = ""

		cluster, err := upsertClusterFromReport(nodeID, &input)
		if err != nil {
			logger.Error("Failed to update cluster from report", "error", err, "node_id", nodeID)
			updates["k8s_last_error"] = "failed to update cluster state on API"
		} else {
			logger.Info("Kubernetes cluster initialized", "bootstrap_node_id", nodeID, "cluster_id", cluster.ID)
		}

	case "joined_control_plane":
		updates["k8s_state"] = "joined_control_plane"
		updates["k8s_joined_at"] = &now
		updates["k8s_last_error"] = ""

	case "joined_worker":
		updates["k8s_state"] = "joined_worker"
		updates["k8s_joined_at"] = &now
		updates["k8s_last_error"] = ""

	case "error":
		updates["k8s_state"] = "error"
		errMsg := truncateString(nonEmpty(input.Message, "unknown error"), 4000)
		updates["k8s_last_error"] = errMsg

		
		
		low := strings.ToLower(errMsg)
		if strings.Contains(low, "kubeadm-certs") && strings.Contains(low, "secret") && strings.Contains(low, "not found") {
			var cluster models.KubernetesCluster
			if err := database.DB.Order("id asc").First(&cluster).Error; err == nil && cluster.InitializedAt != nil {
				expires := time.Now().Add(-1 * time.Minute)
				if err := database.DB.Model(&models.KubernetesCluster{}).
					Where("id = ?", cluster.ID).
					Updates(map[string]any{
						"worker_join_command":        "",
						"control_plane_join_command": "",
						"join_command_expires_at":    &expires,
					}).Error; err == nil {
					logger.Info("Kubernetes join cert secret missing; forcing join-command refresh", "cluster_id", cluster.ID, "node_id", nodeID)
				} else {
					logger.Error("Failed to mark join commands stale after kubeadm-certs missing", "error", err, "cluster_id", cluster.ID, "node_id", nodeID)
				}
			}
		}

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Unknown state"})
	}

	if err := database.DB.Model(&models.Node{}).Where("id = ?", nodeID).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update node state"})
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

func AdminGetKubernetesCluster(c *fiber.Ctx) error {
	var cluster models.KubernetesCluster
	if err := database.DB.Order("id asc").First(&cluster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{"cluster": nil})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load cluster"})
	}

	
	type clusterSummary struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`

		BootstrapNodeID *uint `json:"bootstrap_node_id,omitempty"`

		ControlPlaneEndpoint string `json:"control_plane_endpoint"`
		PodCIDR              string `json:"pod_cidr"`
		ServiceCIDR          string `json:"service_cidr"`
		KubernetesVersion    string `json:"kubernetes_version"`

		InitializedAt        *time.Time `json:"initialized_at,omitempty"`
		JoinCommandExpiresAt *time.Time `json:"join_command_expires_at,omitempty"`
	}

	return c.JSON(fiber.Map{
		"cluster": clusterSummary{
			ID:                   cluster.ID,
			CreatedAt:            cluster.CreatedAt,
			UpdatedAt:            cluster.UpdatedAt,
			BootstrapNodeID:      cluster.BootstrapNodeID,
			ControlPlaneEndpoint: cluster.ControlPlaneEndpoint,
			PodCIDR:              cluster.PodCIDR,
			ServiceCIDR:          cluster.ServiceCIDR,
			KubernetesVersion:    cluster.KubernetesVersion,
			InitializedAt:        cluster.InitializedAt,
			JoinCommandExpiresAt: cluster.JoinCommandExpiresAt,
		},
	})
}

func AdminRefreshKubernetesJoinCommands(c *fiber.Ctx) error {
	var cluster models.KubernetesCluster
	if err := database.DB.Order("id asc").First(&cluster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Cluster not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load cluster"})
	}

	
	now := time.Now()
	updates := map[string]any{
		"worker_join_command":        "",
		"control_plane_join_command": "",
		"join_command_expires_at":    &now,
	}
	if err := database.DB.Model(&models.KubernetesCluster{}).Where("id = ?", cluster.ID).Updates(updates).Error; err != nil {
		logger.Error("Failed to request join command refresh", "error", err, "cluster_id", cluster.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to request refresh"})
	}

	logger.Info("Requested Kubernetes join command refresh", "cluster_id", cluster.ID)
	return c.JSON(fiber.Map{"message": "refresh requested"})
}

func getOrCreateKubernetesCluster(bootstrapHub *models.Node) (*models.KubernetesCluster, error) {
	var cluster models.KubernetesCluster
	err := database.DB.Order("id asc").First(&cluster).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		endpoint := bootstrapHub.PublicIP
		if ip, err := services.GetNodeLoopbackIP(bootstrapHub.ID); err == nil && ip != "" {
			endpoint = ip
		}

		newCluster := models.KubernetesCluster{
			BootstrapNodeID:      &bootstrapHub.ID,
			ControlPlaneEndpoint: endpoint,
			PodCIDR:              defaultPodCIDR,
			ServiceCIDR:          defaultServiceCIDR,
			KubernetesVersion:    defaultK8sVersion,
		}
		if err := database.DB.Create(&newCluster).Error; err != nil {
			return nil, err
		}
		return &newCluster, nil
	}

	changed := false
	if cluster.BootstrapNodeID == nil {
		cluster.BootstrapNodeID = &bootstrapHub.ID
		changed = true
	}
	if cluster.ControlPlaneEndpoint == "" {
		endpoint := bootstrapHub.PublicIP
		if ip, err := services.GetNodeLoopbackIP(bootstrapHub.ID); err == nil && ip != "" {
			endpoint = ip
		}
		cluster.ControlPlaneEndpoint = endpoint
		changed = true
	}
	if cluster.PodCIDR == "" {
		cluster.PodCIDR = defaultPodCIDR
		changed = true
	}
	if cluster.ServiceCIDR == "" {
		cluster.ServiceCIDR = defaultServiceCIDR
		changed = true
	}
	if cluster.KubernetesVersion == "" {
		cluster.KubernetesVersion = defaultK8sVersion
		changed = true
	}
	if changed {
		if err := database.DB.Save(&cluster).Error; err != nil {
			return nil, err
		}
	}

	return &cluster, nil
}

func upsertClusterFromReport(nodeID uint, r *kubernetesReport) (*models.KubernetesCluster, error) {
	var cluster models.KubernetesCluster
	err := database.DB.Order("id asc").First(&cluster).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		cluster = models.KubernetesCluster{}
	}

	changed := false

	if cluster.BootstrapNodeID == nil {
		cluster.BootstrapNodeID = &nodeID
		changed = true
	}

	if r.ControlPlaneEndpoint != "" && r.ControlPlaneEndpoint != cluster.ControlPlaneEndpoint {
		cluster.ControlPlaneEndpoint = r.ControlPlaneEndpoint
		changed = true
	}
	if r.PodCIDR != "" && r.PodCIDR != cluster.PodCIDR {
		cluster.PodCIDR = r.PodCIDR
		changed = true
	}
	if r.ServiceCIDR != "" && r.ServiceCIDR != cluster.ServiceCIDR {
		cluster.ServiceCIDR = r.ServiceCIDR
		changed = true
	}
	if r.KubernetesVersion != "" && r.KubernetesVersion != cluster.KubernetesVersion {
		cluster.KubernetesVersion = r.KubernetesVersion
		changed = true
	}

	if r.WorkerJoinCommand != "" && r.WorkerJoinCommand != cluster.WorkerJoinCommand {
		cluster.WorkerJoinCommand = r.WorkerJoinCommand
		changed = true
	}
	if r.ControlPlaneJoinCommand != "" && r.ControlPlaneJoinCommand != cluster.ControlPlaneJoinCommand {
		cluster.ControlPlaneJoinCommand = r.ControlPlaneJoinCommand
		changed = true
	}

	if r.JoinCommandExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, r.JoinCommandExpiresAt); err == nil {
			cluster.JoinCommandExpiresAt = &t
			changed = true
		}
	}

	if cluster.InitializedAt == nil {
		now := time.Now()
		cluster.InitializedAt = &now
		changed = true
	}

	if cluster.ID == 0 {
		if err := database.DB.Create(&cluster).Error; err != nil {
			return nil, err
		}
		return &cluster, nil
	}
	if changed {
		if err := database.DB.Save(&cluster).Error; err != nil {
			return nil, err
		}
	}

	return &cluster, nil
}

func nonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

func wantsControlPlane(node *models.Node) bool {
	if node == nil {
		return false
	}
	if node.Role == models.NodeRoleHub {
		return true
	}
	return strings.ToLower(strings.TrimSpace(node.ReportedDesiredRole)) == string(models.NodeRoleHub)
}

func shouldRefreshJoinCommands(cluster *models.KubernetesCluster) bool {
	if cluster == nil {
		return false
	}
	if strings.TrimSpace(cluster.WorkerJoinCommand) == "" || strings.TrimSpace(cluster.ControlPlaneJoinCommand) == "" {
		return true
	}
	if cluster.JoinCommandExpiresAt == nil {
		return true
	}
	return time.Until(*cluster.JoinCommandExpiresAt) <= joinRefreshWindow
}
