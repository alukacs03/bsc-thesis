package controllers

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"gluon-api/services"
	"gluon-api/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func AgentStatus(c *fiber.Ctx) error {
	return c.SendString("Agent status endpoint")
}

func RequestEnrollment(c *fiber.Ctx) error {
	logger.Info("Enrollment request received from " + c.IP() + " with user_agent " + c.Get("User-Agent"))

	type enrollmentRequestInput struct {
		Hostname    string `json:"hostname"`
		Provider    string `json:"provider"`
		OS          string `json:"os"`
		DesiredRole string `json:"desired_role"`
	}

	var input enrollmentRequestInput
	if err := c.BodyParser(&input); err != nil {
		logger.Error("Failed to parse enrollment request: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	var raw map[string]any
	if err := c.BodyParser(&raw); err != nil {
		logger.Error("Failed to parse enrollment request: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	allowedFields := []string{"hostname", "provider", "os", "desired_role"}
	if len(raw) != len(allowedFields) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Request must contain exactly the required fields",
		})
	}
	for _, f := range allowedFields {
		v, ok := raw[f]
		if !ok || v == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Missing or empty required field: " + f,
			})
		}
	}

	plainSecret, err := utils.GenerateEnrollmentSecret()
	if err != nil {
		logger.Error("Failed to generate enrollment secret: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate enrollment secret",
		})
	}

	secretHash, secretHashIndex, err := utils.HashEnrollmentSecret(plainSecret)
	if err != nil {
		logger.Error("Failed to hash enrollment secret: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process enrollment secret",
		})
	}

	req := models.NodeEnrollmentRequest{
		Hostname:        raw["hostname"].(string),
		PublicIP:        c.IP(),
		Provider:        raw["provider"].(string),
		OS:              raw["os"].(string),
		DesiredRole:     models.NodeRole(raw["desired_role"].(string)),
		RequestedAt:     time.Now(),
		Status:          "pending",
		SecretHash:      secretHash,
		SecretHashIndex: secretHashIndex,
	}

	logger.Info("Enrollment request details", "hostname", req.Hostname, "public_ip", req.PublicIP, "provider", req.Provider, "os", req.OS, "desired_role", req.DesiredRole)

	var existingReq models.NodeEnrollmentRequest
	if err := database.DB.Where("public_ip = ? OR hostname = ?", req.PublicIP, req.Hostname).First(&existingReq).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "An enrollment request with this IP or hostname already exists",
		})
	}

	var existingNode models.Node
	if err := database.DB.Where("public_ip = ? OR hostname = ?", req.PublicIP, req.Hostname).First(&existingNode).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "A node with this IP or hostname already exists",
		})
	}

	if err := database.DB.Create(&req).Error; err != nil {
		logger.Error("Failed to save enrollment request: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process enrollment request",
		})
	}
	logger.Info("Enrollment request saved", "request_id", req.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":           "Enrollment request received",
		"request_id":        req.ID,
		"status":            req.Status,
		"enrollment_secret": plainSecret,
	})
}

func AcceptAgentEnrollment(c *fiber.Ctx) error {
	req_id_str := c.Params("id")
	req_id, err := strconv.Atoi(req_id_str)
	if err != nil {
		logger.Error("Invalid request ID: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	request := models.NodeEnrollmentRequest{}
	if err := database.DB.First(&request, req_id).Error; err != nil {
		logger.Error("Enrollment request not found: ", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Enrollment request not found",
		})
	}
	if request.Status != "pending" {
		logger.Warn("Enrollment request is not pending", "request_id", req_id, "status", request.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Enrollment request is not pending",
		})
	}

	request.Status = "accepted"
	if err := database.DB.Save(&request).Error; err != nil {
		logger.Error("Failed to update enrollment request status: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update enrollment request status",
		})
	}
	user, err := getUserFromToken(c)
	if err != nil {
		return err
	}
	uid := user.ID
	logger.Audit(c, "Accepted enrollment request", &uid, "accept_enrollment_request", "node_enrollment_request", map[string]any{
		"request_id": req_id,
	})
	node := models.Node{
		Hostname:            request.Hostname,
		Role:                request.DesiredRole,
		PublicIP:            request.PublicIP,
		Provider:            request.Provider,
		OS:                  request.OS,
		Status:              models.NodeStatusActive,
		EnrolledByID:        &user.ID,
		EnrollmentRequestID: uint(req_id),
	}
	if err := database.DB.Create(&node).Error; err != nil {
		logger.Error("Failed to create node from enrollment request: ", "error", err)
		return err
	}

	if err := services.SetupNodeNetworking(&node); err != nil {
		logger.Error("Failed to setup networking for node: ", "error", err, "node_id", node.ID)
	} else {
		logger.Info("Networking setup completed for node", "node_id", node.ID)
	}

	request.ApprovedBy = user
	now := time.Now()
	request.ApprovedAt = &now
	request.ConvertedNodeID = &node.ID
	if err := database.DB.Save(&request).Error; err != nil {
		logger.Error("Failed to update enrollment request approver: ", "error", err)
		return err
	}

	logger.Info("Enrollment request accepted", "request_id", req_id)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Enrollment request accepted",
		"node_id": node.ID,
	})
}

func RejectAgentEnrollment(c *fiber.Ctx) error {
	req_id_str := c.Params("id")
	req_id, err := strconv.Atoi(req_id_str)
	if err != nil {
		logger.Error("Invalid request ID: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	request := models.NodeEnrollmentRequest{}
	if err := database.DB.First(&request, req_id).Error; err != nil {
		logger.Error("Enrollment request not found: ", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Enrollment request not found",
		})
	}
	if request.Status != "pending" {
		logger.Warn("Enrollment request is not pending", "request_id", req_id, "status", request.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Enrollment request is not pending",
		})
	}

	request.Status = "rejected"
	if err := database.DB.Save(&request).Error; err != nil {
		logger.Error("Failed to update enrollment request status: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update enrollment request status",
		})
	}
	user, err := getUserFromToken(c)
	if err != nil {
		return err
	}
	uid := user.ID
	logger.Audit(c, "Rejected enrollment request", &uid, "reject_enrollment_request", "node_enrollment_request", map[string]any{
		"request_id": req_id,
	})

	logger.Info("Enrollment request rejected", "request_id", req_id)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Enrollment request rejected",
	})
}

func CheckAgentEnrollmentStatus(c *fiber.Ctx) error {
	type enrollmentStatusInput struct {
		RequestID        uint   `json:"request_id"`
		EnrollmentSecret string `json:"enrollment_secret"`
	}

	var input enrollmentStatusInput
	if err := c.BodyParser(&input); err != nil {
		logger.Error("Failed to parse enrollment status request: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	if input.RequestID == 0 || input.EnrollmentSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing request_id or enrollment_secret",
		})
	}
	if !strings.HasPrefix(input.EnrollmentSecret, "es_") || len(input.EnrollmentSecret) != 67 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid enrollment secret",
		})
	}

	sha := sha256.Sum256([]byte(input.EnrollmentSecret))
	hashIndex := hex.EncodeToString(sha[:8])

	request := models.NodeEnrollmentRequest{}
	if err := database.DB.Where("id = ? AND secret_hash_index = ?", input.RequestID, hashIndex).First(&request).Error; err != nil {
		logger.Warn("Enrollment status lookup failed", "request_id", input.RequestID, "reason", "not_found_or_secret_mismatch")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid enrollment secret",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(request.SecretHash), []byte(input.EnrollmentSecret)); err != nil {
		logger.Warn("Enrollment secret mismatch", "request_id", input.RequestID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid enrollment secret",
		})
	}

	if request.Status == "accepted" && request.ConvertedNodeID != nil {
		var existingAPIKey models.APIKey
		if err := database.DB.Where("node_id = ?", *request.ConvertedNodeID).First(&existingAPIKey).Error; err == nil {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"request_id": input.RequestID,
				"status":     request.Status,
			})
		}
		plainKey, err := utils.GenerateAPIKey()
		if err != nil {
			logger.Error("Failed to generate API key: ", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate API key",
			})
		}

		hashedKey, hashIndex, err := utils.HashAPIKey(plainKey)
		if err != nil {
			logger.Error("Failed to hash API key: ", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash API key",
			})
		}

		node := models.Node{}
		if err := database.DB.First(&node, *request.ConvertedNodeID).Error; err != nil {
			logger.Error("Failed to find node for API key generation: ", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to find node for API key generation",
			})
		}

		apiKey := models.APIKey{
			NodeID:    node.ID,
			Name:      node.Hostname + "_default",
			Hash:      hashedKey,
			HashIndex: hashIndex,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.DB.Create(&apiKey).Error; err != nil {
			logger.Error("Failed to store API key: ", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to store API key",
			})
		}
		logger.Info("API key generated for node", "node_id", node.ID, "api_key_id", apiKey.ID)
		logger.Audit(
			c,
			"Generated API key for enrolled node",
			nil,
			"generate_api_key",
			"api_key",
			map[string]any{
				"node_id":    node.ID,
				"api_key_id": apiKey.ID,
			},
		)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"request_id": input.RequestID,
			"status":     request.Status,
			"node_id":    node.ID,
			"api_key":    plainKey,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"request_id": input.RequestID,
		"status":     request.Status,
	})
}

func Heartbeat(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	type HeartbeatInput struct {
		AgentVersion string   `json:"agent_version"`
		DesiredRole  string   `json:"desired_role"`
		CPUUsage     *float64 `json:"cpu_usage"`
		MemoryUsage  *float64 `json:"memory_usage"`
		DiskUsage    *float64 `json:"disk_usage"`
		DiskTotalBytes *uint64 `json:"disk_total_bytes"`
		DiskUsedBytes  *uint64 `json:"disk_used_bytes"`
		UptimeSeconds  *uint64 `json:"uptime_seconds"`
		Logs         []string `json:"logs"`
		SystemUsers  []string `json:"system_users"`
		SystemServices []struct {
			Name          string `json:"name"`
			Description   string `json:"description"`
			ActiveState   string `json:"active_state"`
			SubState      string `json:"sub_state"`
			UnitFileState string `json:"unit_file_state"`
		} `json:"system_services"`
		WireGuardPeers []struct {
			Interface           string `json:"interface"`
			PeerPublicKey       string `json:"peer_public_key"`
			Endpoint            string `json:"endpoint"`
			AllowedIPs          string `json:"allowed_ips"`
			LatestHandshakeUnix int64  `json:"latest_handshake_unix"`
			RxBytes             uint64 `json:"rx_bytes"`
			TxBytes             uint64 `json:"tx_bytes"`
		} `json:"wireguard_peers"`
		OSPFNeighbors []struct {
			RouterID             string  `json:"router_id"`
			Area                 string  `json:"area"`
			State                string  `json:"state"`
			Interface            string  `json:"interface"`
			HelloIntervalSeconds *uint64 `json:"hello_interval_seconds"`
			DeadIntervalSeconds  *uint64 `json:"dead_interval_seconds"`
			Cost                 *uint64 `json:"cost"`
			Priority             *uint64 `json:"priority"`
		} `json:"ospf_neighbors"`
	}

	var input HeartbeatInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid heartbeat payload",
			})
		}
	}

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		logger.Error("Node not found for heartbeat: ", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	previousStatus := node.Status
	now := time.Now()
	node.LastSeenAt = &now
	if node.Status != models.NodeStatusActive {
		node.Status = models.NodeStatusActive
	}

	if input.AgentVersion == "" {
		ua := c.Get("User-Agent")
		if strings.HasPrefix(ua, "gluon-agent/") {
			rest := strings.TrimPrefix(ua, "gluon-agent/")
			if ver, _, ok := strings.Cut(rest, " "); ok {
				node.AgentVersion = ver
			} else if ver, _, ok := strings.Cut(rest, "("); ok {
				node.AgentVersion = strings.TrimSpace(ver)
			} else {
				node.AgentVersion = strings.TrimSpace(rest)
			}
		}
	} else {
		node.AgentVersion = input.AgentVersion
	}

	if input.DesiredRole != "" {
		role := strings.ToLower(strings.TrimSpace(input.DesiredRole))
		if role == string(models.NodeRoleHub) || role == string(models.NodeRoleWorker) {
			node.ReportedDesiredRole = role

			
			
			if role == string(models.NodeRoleHub) && node.Role != models.NodeRoleHub {
				var hubCount int64
				if err := database.DB.Model(&models.Node{}).Where("role = ?", models.NodeRoleHub).Count(&hubCount).Error; err != nil {
					logger.Error("Failed to count hubs for promotion", "error", err, "node_id", node.ID)
				} else if hubCount >= services.MaxHubs {
					logger.Warn("Refusing to promote node to hub (max hubs reached)", "node_id", node.ID, "hub_count", hubCount)
				} else {
					if err := database.DB.Model(&models.Node{}).Where("id = ?", node.ID).Update("role", models.NodeRoleHub).Error; err != nil {
						logger.Error("Failed to promote node to hub", "error", err, "node_id", node.ID)
					} else {
						node.Role = models.NodeRoleHub
						logger.Info("Promoted node to hub based on agent desired_role", "node_id", node.ID, "hostname", node.Hostname)

						
						if err := services.SetupNodeNetworking(&node); err != nil {
							logger.Error("Failed to setup networking after hub promotion", "error", err, "node_id", node.ID)
						}
					}
				}
			}
		}
	}

	if input.CPUUsage != nil {
		v := *input.CPUUsage
		node.CPUUsage = &v
	} else {
		node.CPUUsage = nil
	}
	if input.MemoryUsage != nil {
		v := *input.MemoryUsage
		node.MemoryUsage = &v
	} else {
		node.MemoryUsage = nil
	}
	if input.DiskUsage != nil {
		v := *input.DiskUsage
		node.DiskUsage = &v
	} else {
		node.DiskUsage = nil
	}

	if input.DiskTotalBytes != nil {
		v := *input.DiskTotalBytes
		node.DiskTotalBytes = &v
	} else {
		node.DiskTotalBytes = nil
	}
	if input.DiskUsedBytes != nil {
		v := *input.DiskUsedBytes
		node.DiskUsedBytes = &v
	} else {
		node.DiskUsedBytes = nil
	}
	if input.UptimeSeconds != nil {
		v := *input.UptimeSeconds
		node.UptimeSeconds = &v
	} else {
		node.UptimeSeconds = nil
	}

	if len(input.WireGuardPeers) > 0 {
		var ifaces []models.WireGuardInterface
		if err := database.DB.Where("node_id = ?", node.ID).Find(&ifaces).Error; err != nil {
			logger.Error("Failed to load interfaces for WG telemetry", "error", err, "node_id", node.ID)
		} else {
			ifaceByName := make(map[string]models.WireGuardInterface, len(ifaces))
			for _, iface := range ifaces {
				ifaceByName[iface.Name] = iface
			}

			seenIfaces := make(map[string]bool)
			for _, p := range input.WireGuardPeers {
				iface, ok := ifaceByName[p.Interface]
				if !ok {
					continue
				}
				seenIfaces[p.Interface] = true

				var handshakeAt *time.Time
				if p.LatestHandshakeUnix > 0 {
					t := time.Unix(p.LatestHandshakeUnix, 0)
					handshakeAt = &t
				}

				updates := map[string]any{
					"endpoint":          p.Endpoint,
					"allowed_ips":       p.AllowedIPs,
					"last_handshake_at": handshakeAt,
					"rx_bytes":          p.RxBytes,
					"tx_bytes":          p.TxBytes,
				}

				tx := database.DB.Model(&models.NodePeer{}).
					Where("interface_id = ? AND peer_public_key = ?", iface.ID, p.PeerPublicKey).
					Updates(updates)
				if tx.Error != nil {
					logger.Error("Failed to update WG peer telemetry", "error", tx.Error, "node_id", node.ID, "interface", p.Interface)
					continue
				}
				if tx.RowsAffected == 0 && p.Endpoint != "" && p.Endpoint != "(none)" {
					fallbackUpdates := make(map[string]any, len(updates)+1)
					for k, v := range updates {
						fallbackUpdates[k] = v
					}
					fallbackUpdates["peer_public_key"] = p.PeerPublicKey

					tx2 := database.DB.Model(&models.NodePeer{}).
						Where("interface_id = ? AND endpoint = ?", iface.ID, p.Endpoint).
						Updates(fallbackUpdates)
					if tx2.Error != nil {
						logger.Error("Failed to update WG peer telemetry (endpoint fallback)", "error", tx2.Error, "node_id", node.ID, "interface", p.Interface)
					}
				}
			}

			for name := range seenIfaces {
				iface := ifaceByName[name]
				if iface.Status != models.InterfaceStatusUp {
					_ = database.DB.Model(&models.WireGuardInterface{}).
						Where("id = ?", iface.ID).
						Update("status", models.InterfaceStatusUp).Error
				}
			}
		}
	}

	if input.OSPFNeighbors == nil {
		input.OSPFNeighbors = []struct {
			RouterID             string  `json:"router_id"`
			Area                 string  `json:"area"`
			State                string  `json:"state"`
			Interface            string  `json:"interface"`
			HelloIntervalSeconds *uint64 `json:"hello_interval_seconds"`
			DeadIntervalSeconds  *uint64 `json:"dead_interval_seconds"`
			Cost                 *uint64 `json:"cost"`
			Priority             *uint64 `json:"priority"`
		}{}
	}
	ospfJSON, err := json.Marshal(input.OSPFNeighbors)
	if err != nil {
		logger.Error("Failed to marshal OSPF neighbors", "error", err, "node_id", node.ID)
	} else {
		node.OSPFNeighbors = ospfJSON
	}

	logsJSON, err := json.Marshal(input.Logs)
	if err != nil {
		logger.Error("Failed to marshal heartbeat logs", "error", err, "node_id", node.ID)
	} else {
		node.HeartbeatLogs = logsJSON
	}

	if input.SystemUsers == nil {
		input.SystemUsers = []string{}
	}
	usersJSON, err := json.Marshal(input.SystemUsers)
	if err != nil {
		logger.Error("Failed to marshal system users", "error", err, "node_id", node.ID)
	} else {
		node.SystemUsers = usersJSON
	}

	if input.SystemServices == nil {
		input.SystemServices = []struct {
			Name          string `json:"name"`
			Description   string `json:"description"`
			ActiveState   string `json:"active_state"`
			SubState      string `json:"sub_state"`
			UnitFileState string `json:"unit_file_state"`
		}{}
	}
	servicesJSON, err := json.Marshal(input.SystemServices)
	if err != nil {
		logger.Error("Failed to marshal system services", "error", err, "node_id", node.ID)
	} else {
		node.SystemServices = servicesJSON
	}

	if err := database.DB.Save(&node).Error; err != nil {
		logger.Error("Failed to update node heartbeat: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update heartbeat",
		})
	}

	if previousStatus != models.NodeStatusActive && node.Status == models.NodeStatusActive {
		event := models.Event{
			Kind:    models.EventKindNodeOnline,
			NodeID:  &node.ID,
			Message: "Heartbeat received; node marked online",
		}
		if err := database.DB.Create(&event).Error; err != nil {
			logger.Error("Failed to create node online event", "error", err, "node_id", node.ID)
		}
	}

	commands := []models.NodeCommand{}
	if err := database.DB.
		Where("node_id = ? AND status = ?", node.ID, models.NodeCommandStatusPending).
		Order("id asc").
		Limit(10).
		Find(&commands).Error; err != nil {
		logger.Error("Failed to load pending node commands", "error", err, "node_id", node.ID)
		commands = []models.NodeCommand{}
	}

	if len(commands) > 0 {
		now := time.Now()
		for i := range commands {
			cmd := commands[i]
			_ = database.DB.Model(&models.NodeCommand{}).
				Where("id = ? AND status = ?", cmd.ID, models.NodeCommandStatusPending).
				Updates(map[string]any{
					"status":     models.NodeCommandStatusRunning,
					"started_at": &now,
				}).Error
			commands[i].Status = models.NodeCommandStatusRunning
			commands[i].StartedAt = &now
		}
	}

	logger.Debug("Heartbeat received from node", "node_id", nodeID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Heartbeat received",
		"commands": commands,
	})

}

func ListAgentEnrollmentRequests(c *fiber.Ctx) error {
	var requests []models.NodeEnrollmentRequest
	if err := database.DB.Preload("ApprovedBy").Preload("RejectedBy").Find(&requests).Error; err != nil {
		logger.Error("Failed to list enrollment requests: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list enrollment requests",
		})
	}

	return c.JSON(requests)
}
