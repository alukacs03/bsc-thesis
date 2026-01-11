package controllers

import (
	"fmt"
	"gluon-api/models"
	"gluon-api/services"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type deploymentSettingsInput struct {
	LoopbackCIDR          string `json:"loopback_cidr"`
	HubToHubCIDR          string `json:"hub_to_hub_cidr"`
	Hub1WorkerCIDR        string `json:"hub1_worker_cidr"`
	Hub2WorkerCIDR        string `json:"hub2_worker_cidr"`
	Hub3WorkerCIDR        string `json:"hub3_worker_cidr"`
	KubernetesPodCIDR     string `json:"kubernetes_pod_cidr"`
	KubernetesServiceCIDR string `json:"kubernetes_service_cidr"`
	OSPFArea              int    `json:"ospf_area"`
	OSPFHelloInterval     int    `json:"ospf_hello_interval"`
	OSPFDeadInterval      int    `json:"ospf_dead_interval"`
	OSPFHubToHubCost      int    `json:"ospf_hub_to_hub_cost"`
	OSPFHubToWorkerCost   int    `json:"ospf_hub_to_worker_cost"`
	OSPFWorkerToHubCost   int    `json:"ospf_worker_to_hub_cost"`
	Rebuild               bool   `json:"rebuild"`
}

func AdminGetDeploymentSettings(c *fiber.Ctx) error {
	settings, err := services.GetDeploymentSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load deployment settings",
		})
	}

	return c.JSON(settings)
}

func AdminUpdateDeploymentSettings(c *fiber.Ctx) error {
	var input deploymentSettingsInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	existing, err := services.GetDeploymentSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load deployment settings",
		})
	}

	loopbackCIDR, err := requireCIDR(input.LoopbackCIDR, "loopback_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	hubToHubCIDR, err := requireCIDR(input.HubToHubCIDR, "hub_to_hub_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	hub1WorkerCIDR, err := requireCIDR(input.Hub1WorkerCIDR, "hub1_worker_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	hub2WorkerCIDR, err := requireCIDR(input.Hub2WorkerCIDR, "hub2_worker_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	hub3WorkerCIDR, err := requireCIDR(input.Hub3WorkerCIDR, "hub3_worker_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	podCIDR, err := requireCIDR(input.KubernetesPodCIDR, "kubernetes_pod_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	serviceCIDR, err := requireCIDR(input.KubernetesServiceCIDR, "kubernetes_service_cidr")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if input.OSPFArea <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_area must be > 0"})
	}
	if input.OSPFHelloInterval <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_hello_interval must be > 0"})
	}
	if input.OSPFDeadInterval <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_dead_interval must be > 0"})
	}
	if input.OSPFHubToHubCost <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_hub_to_hub_cost must be > 0"})
	}
	if input.OSPFHubToWorkerCost <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_hub_to_worker_cost must be > 0"})
	}
	if input.OSPFWorkerToHubCost <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ospf_worker_to_hub_cost must be > 0"})
	}

	requiresRebuild := loopbackCIDR != strings.TrimSpace(existing.LoopbackCIDR) ||
		hubToHubCIDR != strings.TrimSpace(existing.HubToHubCIDR) ||
		hub1WorkerCIDR != strings.TrimSpace(existing.Hub1WorkerCIDR) ||
		hub2WorkerCIDR != strings.TrimSpace(existing.Hub2WorkerCIDR) ||
		hub3WorkerCIDR != strings.TrimSpace(existing.Hub3WorkerCIDR)

	rebuildRequested := input.Rebuild

	if requiresRebuild && !rebuildRequested {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":            "Networking rebuild required for CIDR changes",
			"requires_rebuild": true,
		})
	}

	settings := models.DeploymentSettings{
		LoopbackCIDR:          loopbackCIDR,
		HubToHubCIDR:          hubToHubCIDR,
		Hub1WorkerCIDR:        hub1WorkerCIDR,
		Hub2WorkerCIDR:        hub2WorkerCIDR,
		Hub3WorkerCIDR:        hub3WorkerCIDR,
		KubernetesPodCIDR:     podCIDR,
		KubernetesServiceCIDR: serviceCIDR,
		OSPFArea:              input.OSPFArea,
		OSPFHelloInterval:     input.OSPFHelloInterval,
		OSPFDeadInterval:      input.OSPFDeadInterval,
		OSPFHubToHubCost:      input.OSPFHubToHubCost,
		OSPFHubToWorkerCost:   input.OSPFHubToWorkerCost,
		OSPFWorkerToHubCost:   input.OSPFWorkerToHubCost,
	}

	updated, err := services.UpdateDeploymentSettings(settings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update deployment settings",
		})
	}

	if rebuildRequested {
		if err := services.RebuildNetworking(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to rebuild networking",
			})
		}
	}

	return c.JSON(updated)
}

func requireCIDR(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if _, _, err := net.ParseCIDR(trimmed); err != nil {
		return "", fmt.Errorf("%s must be a valid CIDR", field)
	}
	return trimmed, nil
}
