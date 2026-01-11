package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func APIKeyAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing API key",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
			})
		}

		providedKey := parts[1]

		if !strings.HasPrefix(providedKey, "glx_") || len(providedKey) != 68 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key format",
			})
		}

		sha := sha256.Sum256([]byte(providedKey))
		searchIndex := hex.EncodeToString(sha[:8])

		var candidates []models.APIKey
		if err := database.DB.
			Select("id", "node_id", "hash", "last_used_at").
			Where("hash_index = ? AND (expires_at IS NULL OR expires_at > ?) AND (revoked_at IS NULL)", searchIndex, time.Now()).
			Find(&candidates).Error; err != nil {
			logger.Error("Database error while fetching API keys: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service error",
			})
		}

		var matchedKey *models.APIKey
		for i := range candidates {
			err := bcrypt.CompareHashAndPassword([]byte(candidates[i].Hash), []byte(providedKey))
			if err == nil {
				matchedKey = &candidates[i]
				break
			}
		}

		if matchedKey == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired API key",
			})
		}

		now := time.Now()
		
		
		if matchedKey.LastUsedAt == nil || now.Sub(*matchedKey.LastUsedAt) >= 30*time.Second {
			if err := database.DB.Model(&models.APIKey{}).
				Where("id = ?", matchedKey.ID).
				UpdateColumn("last_used_at", now).Error; err != nil {
				logger.Error("Failed to update API key last used timestamp: ", err)
			} else {
				matchedKey.LastUsedAt = &now
			}
		}

		c.Locals("api_key", matchedKey)
		c.Locals("node_id", matchedKey.NodeID)
		return c.Next()
	}

}
