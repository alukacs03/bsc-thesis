package routes

import (
	"gluon-api/controllers"
	"gluon-api/middleware"
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
	app.Post("/api/agent/enroll", controllers.RequestEnrollment)
	app.Post("/api/agent/enroll/status", controllers.CheckAgentEnrollmentStatus)

	admin := app.Group("/api/admin")
	// Protected routes for admin users
	admin.Use(jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(secretKey)},
		TokenLookup: "cookie:jwt",
	}))

	admin.Post("modifyRegistrationRequest", controllers.ModifyUserRegistration)
	admin.Post("deleteUser", controllers.DeleteUser)
	admin.Get("userRegRequests", controllers.ListUserRegRequests)
	admin.Post("generateAPIKey", controllers.GenerateAPIKey)
	admin.Post("enrollments/:id/approve", controllers.AcceptAgentEnrollment)
	admin.Post("enrollments/:id/reject", controllers.RejectAgentEnrollment)
	admin.Get("enrollments", controllers.ListAgentEnrollmentRequests)
	admin.Get("nodes", controllers.ListNodes)
	admin.Get("nodes/:id", controllers.GetNode)
	admin.Delete("nodes/:id", controllers.DeleteNode)
	admin.Post("revokeApiKey", controllers.RevokeAPIKey)

	// API key protected routes for agents
	agent := app.Group("/api/agent")
	agent.Use(middleware.APIKeyAuth())
	agent.Post("heartbeat", controllers.Heartbeat)
}
