package services

import (
	"gluon-api/database"
	"gluon-api/models"

	"gorm.io/gorm"
)

func RebuildNetworking() error {
	purposes := []models.IPPoolPurpose{
		models.IPPoolPurposeLoopback,
		models.IPPoolPurposeHubToHub,
		models.IPPoolPurposeHub1Worker,
		models.IPPoolPurposeHub2Worker,
		models.IPPoolPurposeHub3Worker,
	}

	if err := database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.LinkAllocation{}).Error; err != nil {
		return err
	}
	if err := database.DB.Where("purpose = ?", "loopback").Delete(&models.IPAllocation{}).Error; err != nil {
		return err
	}

	var pools []models.IPPool
	if err := database.DB.Where("purpose IN ?", purposes).Find(&pools).Error; err != nil {
		return err
	}

	poolIDs := make([]uint, 0, len(pools))
	for _, pool := range pools {
		poolIDs = append(poolIDs, pool.ID)
	}

	if len(poolIDs) > 0 {
		if err := database.DB.Where("pool_id IN ?", poolIDs).Delete(&models.LinkAllocation{}).Error; err != nil {
			return err
		}
		if err := database.DB.Where("pool_id IN ?", poolIDs).Delete(&models.IPAllocation{}).Error; err != nil {
			return err
		}
		if err := database.DB.Where("id IN ?", poolIDs).Delete(&models.IPPool{}).Error; err != nil {
			return err
		}
	}

	if err := database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.NodePeer{}).Error; err != nil {
		return err
	}
	if err := database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.WireGuardInterface{}).Error; err != nil {
		return err
	}

	if err := EnsureDefaultPools(); err != nil {
		return err
	}

	var nodes []models.Node
	if err := database.DB.Order("id asc").Find(&nodes).Error; err != nil {
		return err
	}
	for i := range nodes {
		if err := SetupNodeNetworking(&nodes[i]); err != nil {
			return err
		}
	}

	return nil
}
