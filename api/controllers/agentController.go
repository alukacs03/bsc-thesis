package controllers

import "github.com/gofiber/fiber/v2"

func AgentStatus(c *fiber.Ctx) error {
	return c.SendString("Agent status endpoint")
}
