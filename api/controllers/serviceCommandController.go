package controllers

import (
	"encoding/json"
	"gluon-api/database"
	"gluon-api/models"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var systemdUnitNameRe = regexp.MustCompile(`^[A-Za-z0-9@._:-]+$`)

func QueueRestartService(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil || nodeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node id"})
	}

	var node models.Node
	if err := database.DB.Select("id").First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	name := strings.TrimSpace(input.Name)
	if name == "" || !systemdUnitNameRe.MatchString(name) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid service name"})
	}
	if !strings.Contains(name, ".") {
		
		name = name + ".service"
	}

	payload, _ := json.Marshal(fiber.Map{"name": name})
	cmd := models.NodeCommand{
		NodeID:  node.ID,
		Kind:    "restart_service",
		Payload: payload,
		Status:  models.NodeCommandStatusPending,
	}

	if err := database.DB.Create(&cmd).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to queue command"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"command_id": cmd.ID,
		"node_id":    cmd.NodeID,
		"kind":       cmd.Kind,
		"queued_at":  time.Now(),
	})
}

func ReportCommandResults(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var input struct {
		Results []struct {
			ID     uint   `json:"id"`
			Status string `json:"status"`
			Output string `json:"output"`
			Error  string `json:"error"`
		} `json:"results"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	now := time.Now()
	updated := 0

	for _, r := range input.Results {
		if r.ID == 0 {
			continue
		}

		status := models.NodeCommandStatusFailed
		switch strings.ToLower(strings.TrimSpace(r.Status)) {
		case "succeeded", "success", "ok":
			status = models.NodeCommandStatusSucceeded
		case "failed", "error":
			status = models.NodeCommandStatusFailed
		default:
			status = models.NodeCommandStatusFailed
		}

		tx := database.DB.Model(&models.NodeCommand{}).
			Where("id = ? AND node_id = ?", r.ID, nodeID).
			Updates(map[string]any{
				"status":       status,
				"completed_at": &now,
				"output":       r.Output,
				"error":        r.Error,
			})
		if tx.Error == nil && tx.RowsAffected > 0 {
			updated++
		}
	}

	return c.JSON(fiber.Map{"updated": updated})
}

