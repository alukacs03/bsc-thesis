package controllers

import (
	"gluon-api/database"
	"gluon-api/models"

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
