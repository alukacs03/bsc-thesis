package routes

import (
	"gluon-api/controllers"
	"os"

	"github.com/gofiber/fiber/v2"

	jwtware "github.com/gofiber/contrib/jwt"
)

var secretKey = func() string {
	s := os.Getenv("SECRET_KEY")
	if s == "" {
		// default for development only; set SECRET_KEY in environment in production
		s = "default_secret"
	}
	return s
}()

func SetupRoutes(app *fiber.App) {
	app.Get("/api", controllers.Hello)
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.User)
	app.Post("/api/logout", controllers.Logout)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(secretKey)},
		TokenLookup: "cookie:jwt",
	}))

	app.Post("/api/modifyRegistrationRequest", controllers.ModifyUserRegistration)
	app.Post("/api/deleteUser", controllers.DeleteUser)
	app.Get("/api/userRegRequests", controllers.ListUserRegRequests)
}
