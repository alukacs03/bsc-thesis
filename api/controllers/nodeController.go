package controllers

import (
	"encoding/json"
	"errors"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"gluon-api/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ListNodes(c *fiber.Ctx) error {
	var nodes []models.Node
	result := database.DB.Find(&nodes)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve nodes",
		})
	}

	return c.JSON(nodes)
}

func GetNode(c *fiber.Ctx) error {
	id := c.Params("id")
	var node models.Node
	result := database.DB.First(&node, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	return c.JSON(node)
}

func DeleteNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	_, _, _, err = performDecommission(nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Node not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decommission node",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func ListNodeLogs(c *fiber.Ctx) error {
	id := c.Params("id")

	limit := 200
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			if n < 1 {
				limit = 1
			} else if n > 1000 {
				limit = 1000
			} else {
				limit = n
			}
		}
	}

	var node models.Node
	result := database.DB.First(&node, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	var logs []string
	if len(node.HeartbeatLogs) > 0 {
		_ = json.Unmarshal(node.HeartbeatLogs, &logs)
	}
	if logs == nil {
		logs = []string{}
	}

	if len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}

	return c.JSON(fiber.Map{
		"window": "2 minutes",
		"logs":   logs,
	})
}

// DecommissionNode gracefully removes a node from the mesh network.
// It marks the node decommissioned, queues a decommission command for the agent,
// and triggers config rebuild for peer nodes.
func DecommissionNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	cmd, node, already, err := performDecommission(nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Node not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decommission node",
		})
	}
	if already {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Node is already decommissioned",
		})
	}

	logger.Info("Node decommissioned", "node_id", nodeID, "hostname", node.Hostname)

	return c.JSON(fiber.Map{
		"message":    "Node decommissioned successfully",
		"node_id":    nodeID,
		"command_id": cmd.ID,
	})
}

func performDecommission(nodeID int) (models.NodeCommand, *models.Node, bool, error) {
	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return models.NodeCommand{}, nil, false, err
	}

	if node.Status == models.NodeStatusDecommissioned {
		return models.NodeCommand{}, &node, true, nil
	}

	tx := database.DB.Begin()
	now := time.Now()
	if err := tx.Model(&node).Updates(map[string]any{
		"status":      models.NodeStatusDecommissioned,
		"last_seen_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return models.NodeCommand{}, &node, false, err
	}

	cmd := models.NodeCommand{
		NodeID: uint(nodeID),
		Kind:   models.CmdKindDecommission,
		Status: models.NodeCommandStatusPending,
	}
	if err := tx.Create(&cmd).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create decommission command", "error", err, "node_id", nodeID)
		return models.NodeCommand{}, &node, false, err
	}

	if err := tx.Commit().Error; err != nil {
		return models.NodeCommand{}, &node, false, err
	}

	go func() {
		if err := services.RebuildNetworking(); err != nil {
			logger.Error("Failed to rebuild networking after decommission", "error", err, "node_id", nodeID)
		}
	}()

	event := models.Event{
		Kind:    models.EventKindNodeDecommission,
		NodeID:  &node.ID,
		Message: "Node decommissioned",
	}
	if err := database.DB.Create(&event).Error; err != nil {
		logger.Error("Failed to create decommission event", "error", err)
	}

	return cmd, &node, false, nil
}
