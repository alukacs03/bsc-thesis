package controllers

import (
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"net/netip"

	"github.com/gofiber/fiber/v2"
)

func AddIPPool(c *fiber.Ctx) error {
	var pool models.IPPool
	if err := c.BodyParser(&pool); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if _, err := netip.ParsePrefix(pool.CIDR); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid CIDR",
		})
	}

	if err := database.DB.Create(&pool).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create IP pool",
		})
	}

	logger.Audit(c, "Created IP pool", nil, "create", "IPPool", pool)
	return c.Status(fiber.StatusCreated).JSON(pool)
}

func ListIPPools(c *fiber.Ctx) error {
	var pools []models.IPPool
	if err := database.DB.Find(&pools).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve IP pools",
		})
	}

	return c.JSON(pools)
}

func DeleteIPPool(c *fiber.Ctx) error {
	id := c.Params("id")
	result := database.DB.Delete(&models.IPPool{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete IP pool",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "IP pool not found",
		})
	}

	logger.Audit(c, "Deleted IP pool", nil, "delete", "IPPool", id)
	return c.SendStatus(fiber.StatusNoContent)
}

func AllocateIP(c *fiber.Ctx) error {
	var allocation models.IPAllocation
	if err := c.BodyParser(&allocation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ipStr := allocation.IP
	if _, err := netip.ParsePrefix(ipStr); err != nil {
		if _, err := netip.ParseAddr(ipStr); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid IP address",
			})
		}
	}

	if err := database.DB.Create(&allocation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create IP allocation",
		})
	}

	logger.Audit(c, "Allocated IP", nil, "create", "IPAllocation", allocation)
	return c.Status(fiber.StatusCreated).JSON(allocation)
}

func ListIPAllocations(c *fiber.Ctx) error {
	var allocations []models.IPAllocation
	if err := database.DB.Find(&allocations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve IP allocations",
		})
	}

	return c.JSON(allocations)
}

func DeallocateIP(c *fiber.Ctx) error {
	id := c.Params("id")
	result := database.DB.Delete(&models.IPAllocation{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete IP allocation",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "IP allocation not found",
		})
	}

	logger.Audit(c, "Deallocated IP", nil, "delete", "IPAllocation", id)
	return c.SendStatus(fiber.StatusNoContent)
}

func GetIPAllocation(c *fiber.Ctx) error {
	id := c.Params("id")
	var allocation models.IPAllocation
	result := database.DB.First(&allocation, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "IP allocation not found",
		})
	}

	return c.JSON(allocation)
}

func findNextAvailableIP(cidrStr string, allocations []models.IPAllocation) (*string, error) {
	prefix, err := netip.ParsePrefix(cidrStr)
	if err != nil {
		return nil, err
	}

	allocated := make(map[netip.Addr]bool)
	for _, alloc := range allocations {
		ipStr := alloc.IP
		if addr, err := netip.ParseAddr(ipStr); err == nil {
			allocated[addr] = true
		} else if prefix, err := netip.ParsePrefix(ipStr); err == nil {
			allocated[prefix.Addr()] = true
		}
	}

	addr := prefix.Addr().Next()
	for prefix.Contains(addr) {
		if !allocated[addr] {
			ipStr := addr.String()
			return &ipStr, nil
		}
		addr = addr.Next()
	}
	return nil, nil
}

func GetNextAvailableIP(c *fiber.Ctx) error {
	poolID := c.Params("id")
	var pool models.IPPool
	if err := database.DB.First(&pool, poolID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Pool not found",
		})
	}

	if _, err := netip.ParsePrefix(pool.CIDR); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid CIDR",
		})
	}

	var allocations []models.IPAllocation
	database.DB.Where("pool_id = ?", poolID).Find(&allocations)

	ip, err := findNextAvailableIP(pool.CIDR, allocations)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to find next IP: " + err.Error(),
		})
	}
	if ip == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No available IP",
		})
	}
	return c.JSON(fiber.Map{"ip": *ip})
}

func AllocateNextAvailableIP(c *fiber.Ctx) error {
	poolID := c.Params("id")
	var pool models.IPPool
	if err := database.DB.First(&pool, poolID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Pool not found",
		})
	}

	if _, err := netip.ParsePrefix(pool.CIDR); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid CIDR",
		})
	}

	var allocations []models.IPAllocation
	database.DB.Where("pool_id = ?", poolID).Find(&allocations)

	ip, err := findNextAvailableIP(pool.CIDR, allocations)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to find next IP: " + err.Error(),
		})
	}
	if ip == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No available IP",
		})
	}

	newAlloc := models.IPAllocation{
		PoolID:  pool.ID,
		IP:      *ip,
		Purpose: "auto-allocated",
	}
	if err := database.DB.Create(&newAlloc).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create IP allocation",
		})
	}
	logger.Audit(c, "Auto-allocated next available IP", nil, "create", "IPAllocation", newAlloc)
	return c.Status(fiber.StatusCreated).JSON(newAlloc)
}
