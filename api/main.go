package main

import (
	"gluon-api/controllers"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	logger.Init()
	logger.Info("Starting Gluon API server...")
	_, err := database.ConnectDB()
	if err != nil {
		logger.Error("Failed to connect to database:", err)
		panic("Failed to connect to database!")
	}

	controllers.AddDemoUser()

	logger.Info("Database connection successful")

	logger.Debug("Setting up Fiber app")
	app := fiber.New()

	logger.Debug("Setting up CORS middleware")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,Accept,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Access-Control-Allow-Methods,Access-Control-Expose-Headers,Access-Control-Max-Age,Access-Control-Allow-Credentials",
		AllowCredentials: true,
	}))

	logger.Debug("Setting up routes")
	routes.SetupRoutes(app)

	logger.Info("Starting server on port 3000")
	err = app.Listen(":3000")
	if err != nil {
		logger.Error("Failed to start server:", err)
		panic(err)
	}
}
