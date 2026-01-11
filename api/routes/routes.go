package routes

import (
	"gluon-api/config"
	"gluon-api/controllers"
	"gluon-api/middleware"

	"github.com/gofiber/fiber/v2"

	jwtware "github.com/gofiber/contrib/jwt"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/metrics", controllers.Metrics)
	app.Get("/api", controllers.Hello)
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.User)
	app.Post("/api/logout", controllers.Logout)
	app.Post("/api/agent/enroll", controllers.RequestEnrollment)
	app.Post("/api/agent/enroll/status", controllers.CheckAgentEnrollmentStatus)

	admin := app.Group("/api/admin")
	admin.Use(jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(config.Current().SecretKey)},
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
	admin.Get("nodes/:id/logs", controllers.ListNodeLogs)
	admin.Delete("nodes/:id", controllers.DeleteNode)
	admin.Post("revokeApiKey", controllers.RevokeAPIKey)
	admin.Get("network/wireguard/peers", controllers.ListWireGuardPeers)
	admin.Get("network/ospf/neighbors", controllers.ListOSPFNeighbors)
	admin.Get("nodes/:id/network/wireguard/peers", controllers.ListWireGuardPeersForNode)
	admin.Get("nodes/:id/network/ospf/neighbors", controllers.ListOSPFNeighborsForNode)
	admin.Get("nodes/:id/ssh-keys", controllers.ListNodeSSHKeys)
	admin.Post("nodes/:id/ssh-keys", controllers.CreateNodeSSHKey)
	admin.Post("nodes/:id/ssh-keys/generate", controllers.GenerateNodeSSHKey)
	admin.Delete("nodes/:id/ssh-keys/:keyId", controllers.DeleteNodeSSHKey)
	admin.Post("nodes/:id/services/restart", controllers.QueueRestartService)
	admin.Get("kubernetes/cluster", controllers.AdminGetKubernetesCluster)
	admin.Post("kubernetes/refresh-join", controllers.AdminRefreshKubernetesJoinCommands)
	admin.Get("kubernetes/workloads", controllers.AdminGetKubernetesWorkloads)
	admin.Post("kubernetes/apply", controllers.AdminApplyKubernetesManifest)
	admin.Get("kubernetes/resource", controllers.AdminGetKubernetesResourceYAML)
	admin.Delete("kubernetes/resource", controllers.AdminDeleteKubernetesResource)
	admin.Get("kubernetes/networking", controllers.AdminGetKubernetesNetworking)
	admin.Get("deployment/settings", controllers.AdminGetDeploymentSettings)
	admin.Put("deployment/settings", controllers.AdminUpdateDeploymentSettings)

	agent := app.Group("/api/agent")
	agent.Use(middleware.APIKeyAuth())
	agent.Post("heartbeat", controllers.Heartbeat)
	agent.Post("commands/report", controllers.ReportCommandResults)

	agent.Get("network/info", controllers.GetNetworkInfo)
	agent.Post("network/keys", controllers.UploadPublicKeys)
	agent.Get("config", controllers.GetConfig)
	agent.Post("config/applied", controllers.ReportConfigApplied)
	agent.Get("kubernetes/task", controllers.GetKubernetesTask)
	agent.Post("kubernetes/report", controllers.ReportKubernetes)

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
