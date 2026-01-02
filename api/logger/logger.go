package logger

import (
	"encoding/json"
	"gluon-api/database"
	"gluon-api/models"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
)

var Logger *slog.Logger

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Audit(c *fiber.Ctx, msg string, actorID *uint, action string, entity string, args ...any) {
	Logger.Info("AUDIT: "+msg, args...)
	db := database.DB
	if db != nil {
		details := map[string]interface{}{
			"msg":    msg,
			"args":   args,
			"entity": entity,
		}
		detailsJSON, err := json.Marshal(details)
		if err != nil {
			Logger.Error("Failed to marshal audit log details:", "error", err)
			return
		}
		auditLog := models.AuditLog{
			Action:    action,
			Entity:    entity,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			ActorID:   actorID,
			Details:   detailsJSON,
		}
		if err := db.Create(&auditLog).Error; err != nil {
			Logger.Error("Failed to save audit log to database:", "error", err)
		}
	} else {
		Logger.Warn("Database connection is nil; cannot save audit log")
	}
}
