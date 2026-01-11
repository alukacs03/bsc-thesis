package services

import (
	"errors"
	"gluon-api/config"
	"gluon-api/database"
	"gluon-api/models"

	"gorm.io/gorm"
)

func LoadDeploymentSettings() error {
	settings, err := ensureDeploymentSettings()
	if err != nil {
		return err
	}
	applyDeploymentSettings(settings)
	return nil
}

func GetDeploymentSettings() (models.DeploymentSettings, error) {
	settings, err := ensureDeploymentSettings()
	if err != nil {
		return models.DeploymentSettings{}, err
	}
	return settings, nil
}

func UpdateDeploymentSettings(input models.DeploymentSettings) (models.DeploymentSettings, error) {
	settings, err := ensureDeploymentSettings()
	if err != nil {
		return models.DeploymentSettings{}, err
	}

	settings.LoopbackCIDR = input.LoopbackCIDR
	settings.HubToHubCIDR = input.HubToHubCIDR
	settings.Hub1WorkerCIDR = input.Hub1WorkerCIDR
	settings.Hub2WorkerCIDR = input.Hub2WorkerCIDR
	settings.Hub3WorkerCIDR = input.Hub3WorkerCIDR
	settings.KubernetesPodCIDR = input.KubernetesPodCIDR
	settings.KubernetesServiceCIDR = input.KubernetesServiceCIDR
	settings.OSPFArea = input.OSPFArea
	settings.OSPFHelloInterval = input.OSPFHelloInterval
	settings.OSPFDeadInterval = input.OSPFDeadInterval
	settings.OSPFHubToHubCost = input.OSPFHubToHubCost
	settings.OSPFHubToWorkerCost = input.OSPFHubToWorkerCost
	settings.OSPFWorkerToHubCost = input.OSPFWorkerToHubCost

	if err := database.DB.Save(&settings).Error; err != nil {
		return models.DeploymentSettings{}, err
	}

	applyDeploymentSettings(settings)
	return settings, nil
}

func ensureDeploymentSettings() (models.DeploymentSettings, error) {
	var settings models.DeploymentSettings
	if err := database.DB.First(&settings).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cfg := config.Current()
			settings = models.DeploymentSettings{
				LoopbackCIDR:          cfg.LoopbackCIDR,
				HubToHubCIDR:          cfg.HubToHubCIDR,
				Hub1WorkerCIDR:        cfg.Hub1WorkerCIDR,
				Hub2WorkerCIDR:        cfg.Hub2WorkerCIDR,
				Hub3WorkerCIDR:        cfg.Hub3WorkerCIDR,
				KubernetesPodCIDR:     cfg.KubernetesPodCIDR,
				KubernetesServiceCIDR: cfg.KubernetesServiceCIDR,
				OSPFArea:              cfg.OSPFArea,
				OSPFHelloInterval:     cfg.OSPFHelloInterval,
				OSPFDeadInterval:      cfg.OSPFDeadInterval,
				OSPFHubToHubCost:      cfg.OSPFHubToHubCost,
				OSPFHubToWorkerCost:   cfg.OSPFHubToWorkerCost,
				OSPFWorkerToHubCost:   cfg.OSPFWorkerToHubCost,
			}
			if err := database.DB.Create(&settings).Error; err != nil {
				return models.DeploymentSettings{}, err
			}
			return settings, nil
		}
		return models.DeploymentSettings{}, err
	}
	return settings, nil
}

func applyDeploymentSettings(settings models.DeploymentSettings) {
	config.ApplyOverrides(config.Overrides{
		LoopbackCIDR:          settings.LoopbackCIDR,
		HubToHubCIDR:          settings.HubToHubCIDR,
		Hub1WorkerCIDR:        settings.Hub1WorkerCIDR,
		Hub2WorkerCIDR:        settings.Hub2WorkerCIDR,
		Hub3WorkerCIDR:        settings.Hub3WorkerCIDR,
		KubernetesPodCIDR:     settings.KubernetesPodCIDR,
		KubernetesServiceCIDR: settings.KubernetesServiceCIDR,
		OSPFArea:              settings.OSPFArea,
		OSPFHelloInterval:     settings.OSPFHelloInterval,
		OSPFDeadInterval:      settings.OSPFDeadInterval,
		OSPFHubToHubCost:      settings.OSPFHubToHubCost,
		OSPFHubToWorkerCost:   settings.OSPFHubToWorkerCost,
		OSPFWorkerToHubCost:   settings.OSPFWorkerToHubCost,
	})
}
