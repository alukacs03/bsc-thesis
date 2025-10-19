package main

import (
	"fmt"
	"gluon-api/controllers"
	"gluon-api/database"
	"gluon-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	_, err := database.ConnectDB()
	if err != nil {
		fmt.Println("Database connection failed:", err)
		panic("Failed to connect to database!")
	}

	controllers.AddDemoUser()

	fmt.Println("Database connection successful")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,Accept,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Access-Control-Allow-Methods,Access-Control-Expose-Headers,Access-Control-Max-Age,Access-Control-Allow-Credentials",
		AllowCredentials: true,
	}))

	routes.SetupRoutes(app)

	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
