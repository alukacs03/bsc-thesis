package controllers

import (
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"gluon-api/utils"
	"time"

	"github.com/gofiber/fiber/v2"
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

	req := models.NodeEnrollmentRequest{
		Hostname:    raw["hostname"].(string),
		PublicIP:    c.IP(),
		Provider:    raw["provider"].(string),
		OS:          raw["os"].(string),
		DesiredRole: models.NodeRole(raw["desired_role"].(string)),
		RequestedAt: time.Now(),
		Status:      "pending",
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

	// save to db
	if err := database.DB.Create(&req).Error; err != nil {
		logger.Error("Failed to save enrollment request: ", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process enrollment request",
		})
	}
	logger.Info("Enrollment request saved", "request_id", req.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Enrollment request received",
		"request_id": req.ID,
	})
}

func AcceptAgentEnrollment(c *fiber.Ctx) error {
	type acceptEnrollmentInput struct {
		RequestID uint `json:"request_id"`
	}
	var input acceptEnrollmentInput
	if err := c.BodyParser(&input); err != nil {
		logger.Error("Failed to parse accept enrollment request: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}
	req_id := input.RequestID

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
	// audit log
	user, err := getUserFromToken(c)
	if err != nil {
		return err
	}
	uid := user.ID
	logger.Audit(c, "Accepted enrollment request", &uid, "accept_enrollment_request", "node_enrollment_request", map[string]any{
		"request_id": req_id,
	})
	// create node
	node := models.Node{
		Hostname:            request.Hostname,
		Role:                models.NodeRole("worker"),
		PublicIP:            request.PublicIP,
		Provider:            request.Provider,
		OS:                  request.OS,
		Status:              models.NodeStatusActive,
		EnrolledByID:        &user.ID,
		EnrollmentRequestID: req_id,
	}
	if err := database.DB.Create(&node).Error; err != nil {
		logger.Error("Failed to create node from enrollment request: ", "error", err)
		return err
	}
	// Safely set ApprovedBy and ApprovedAt
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

func CheckAgentEnrollmentStatus(c *fiber.Ctx) error {
	type enrollmentStatusInput struct {
		RequestID uint `json:"request_id"`
	}
	var input enrollmentStatusInput
	if err := c.BodyParser(&input); err != nil {
		logger.Error("Failed to parse enrollment status request: ", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}
	req_id := input.RequestID

	request := models.NodeEnrollmentRequest{}
	if err := database.DB.First(&request, req_id).Error; err != nil {
		logger.Error("Enrollment request not found: ", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Enrollment request not found",
		})
	}

	if request.Status == "accepted" && request.ConvertedNodeID != nil {
		// check if API key already exists for this node and if yes, do not return. we've already returned it once
		var existingAPIKey models.APIKey
		if err := database.DB.Where("node_id = ?", *request.ConvertedNodeID).First(&existingAPIKey).Error; err == nil {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"request_id": req_id,
				"status":     request.Status,
			})
		}

		// generate API key
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
			"request_id": req_id,
			"status":     request.Status,
			"api_key":    plainKey,
		})

	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"request_id": req_id,
		"status":     request.Status,
	})
}
