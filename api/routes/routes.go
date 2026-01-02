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

	agent := app.Group("/api/agent")
	agent.Use(middleware.APIKeyAuth())
	agent.Post("heartbeat", controllers.Heartbeat)

	agent.Get("network/info", controllers.GetNetworkInfo)
	agent.Post("network/keys", controllers.UploadPublicKeys)
	agent.Get("config", controllers.GetConfig)
	agent.Post("config/applied", controllers.ReportConfigApplied)

	admin.Get("ipam/pools", controllers.ListIPPools)
	admin.Post("ipam/pools", controllers.AddIPPool)
	admin.Delete("ipam/pools/:id", controllers.DeleteIPPool)
	admin.Get("ipam/allocations", controllers.ListIPAllocations)
	admin.Post("ipam/allocations", controllers.AllocateIP)
	admin.Delete("ipam/allocations/:id", controllers.DeallocateIP)
	admin.Get("ipam/allocations/:id", controllers.GetIPAllocation)
	admin.Get("ipam/pools/:id/next", controllers.GetNextAvailableIP)
	admin.Post("ipam/pools/:id/allocate-next", controllers.AllocateNextAvailableIP)
}
