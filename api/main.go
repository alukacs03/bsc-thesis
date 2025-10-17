package main

import (
	"fmt"
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

	fmt.Println("Database connection successful")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,Accept,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Access-Control-Allow-Methods,Access-Control-Expose-Headers,Access-Control-Max-Age,Access-Control-Allow-Credentials",
		AllowCredentials: false,
	}))

	routes.SetupRoutes(app)

	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
