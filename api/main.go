package main

import (
	"crypto/tls"
	"gluon-api/certs"
	"gluon-api/config"
	"gluon-api/controllers"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/metrics"
	"gluon-api/middleware"
	"gluon-api/models"
	"gluon-api/routes"
	"gluon-api/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	logger.Init()
	logger.Info("Starting Gluon API server...")

	if err := config.Load(); err != nil {
		logger.Error("Failed to load config", "error", err)
		panic(err)
	}
	_, err := database.ConnectDB()
	if err != nil {
		logger.Error("Failed to connect to database:", err)
		panic("Failed to connect to database!")
	}
	if err := services.LoadDeploymentSettings(); err != nil {
		logger.Error("Failed to load deployment settings", "error", err)
	}

	controllers.AddDemoUser()
	startWorkerOfflineMonitor()
	metrics.StartDatabaseMetrics(30 * time.Second)
	if err := services.AssignHubNumbers(); err != nil {
		logger.Error("Failed to assign hub numbers", "error", err)
	}

	logger.Info("Database connection successful")

	logger.Debug("Setting up Fiber app")
	app := fiber.New()

	logger.Debug("Setting up CORS middleware")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173",
		AllowOriginsFunc: func(origin string) bool {
			// Allow any origin during development - restrict in production
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,Accept,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Access-Control-Allow-Methods,Access-Control-Expose-Headers,Access-Control-Max-Age,Access-Control-Allow-Credentials",
		AllowCredentials: true,
	}))

	logger.Debug("Setting up Prometheus middleware")
	app.Use(middleware.PrometheusMetrics())

	logger.Debug("Setting up routes")
	routes.SetupRoutes(app)

	cfg := config.Current()
	if cfg.TLSEnabled {
		logger.Info("TLS is enabled, setting up certificates...")
		if err := setupTLS(app, cfg); err != nil {
			logger.Error("Failed to setup TLS", "error", err)
			panic(err)
		}
	} else {
		logger.Info("Starting server on port 3000 (TLS disabled)")
		err = app.Listen(":3000")
		if err != nil {
			logger.Error("Failed to start server:", err)
			panic(err)
		}
	}
}

func setupTLS(app *fiber.App, cfg config.Settings) error {
	ca, caKey, err := certs.EnsureCA(cfg.CACertPath, cfg.CAKeyPath)
	if err != nil {
		return err
	}
	logger.Info("CA certificate ready", "path", cfg.CACertPath)

	// Generate server certificate if needed
	serverCertPEM, serverKeyPEM, err := certs.GenerateServerCert(ca, caKey, cfg.TLSHosts)
	if err != nil {
		return err
	}

	// Save server certificate
	if err := certs.SaveServerCert(serverCertPEM, serverKeyPEM, cfg.TLSCertPath, cfg.TLSKeyPath); err != nil {
		return err
	}
	logger.Info("Server certificate ready", "path", cfg.TLSCertPath, "hosts", cfg.TLSHosts)

	// Load certificate for TLS config
	cert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create TLS listener
	ln, err := tls.Listen("tcp", ":3000", tlsConfig)
	if err != nil {
		return err
	}

	logger.Info("Starting server on port 3000 with TLS")
	return app.Listener(ln)
}

func startWorkerOfflineMonitor() {
	const (
		checkInterval = 30 * time.Second
		offlineAfter  = 2 * time.Minute
	)

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for range ticker.C {
			cutoff := time.Now().Add(-offlineAfter)
			var nodes []models.Node
			result := database.DB.
				Where(
					`role = ? AND status = ? AND (
						(last_seen_at IS NOT NULL AND last_seen_at < ?) OR
						(last_seen_at IS NULL AND created_at < ?)
					)`,
					models.NodeRoleWorker,
					models.NodeStatusActive,
					cutoff,
					cutoff,
				).
				Find(&nodes)

			if result.Error != nil {
				logger.Error("Failed to find stale workers", "error", result.Error)
				continue
			}

			var marked int64
			for i := range nodes {
				node := nodes[i]
				update := database.DB.
					Model(&models.Node{}).
					Where("id = ? AND status = ?", node.ID, models.NodeStatusActive).
					Updates(map[string]any{
						"status": models.NodeStatusOffline,
					})
				if update.Error != nil {
					logger.Error("Failed to mark worker offline", "error", update.Error, "node_id", node.ID)
					continue
				}
				if update.RowsAffected == 0 {
					continue
				}

				marked++
				event := models.Event{
					Kind:    models.EventKindNodeOffline,
					NodeID:  &node.ID,
					Message: "Node not seen for >2m; marked offline",
				}
				if err := database.DB.Create(&event).Error; err != nil {
					logger.Error("Failed to create node offline event", "error", err, "node_id", node.ID)
				}
			}

			if marked > 0 {
				logger.Info("Marked workers offline", "count", marked)
			}
		}
	}()
}
