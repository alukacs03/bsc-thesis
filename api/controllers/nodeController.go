package controllers

import (
	"encoding/json"
	"gluon-api/database"
	"gluon-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
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
	result := database.DB.Delete(&models.Node{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete node",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
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
