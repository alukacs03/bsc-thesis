package services

import (
	"fmt"
	"gluon-api/config"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"net/netip"
	"sort"

	"gorm.io/gorm"
)

const MaxHubs = 3

func EnsureDefaultPools() error {
	cfg := config.Current()
	pools := []struct {
		Purpose   models.IPPoolPurpose
		CIDR      string
		HubNumber *int
		Kind      models.IPPoolKind
	}{
		{models.IPPoolPurposeLoopback, cfg.LoopbackCIDR, nil, models.IPPoolKindWireGuard},
		{models.IPPoolPurposeHubToHub, cfg.HubToHubCIDR, nil, models.IPPoolKindWireGuard},
		{models.IPPoolPurposeHub1Worker, cfg.Hub1WorkerCIDR, intPtr(1), models.IPPoolKindWireGuard},
		{models.IPPoolPurposeHub2Worker, cfg.Hub2WorkerCIDR, intPtr(2), models.IPPoolKindWireGuard},
		{models.IPPoolPurposeHub3Worker, cfg.Hub3WorkerCIDR, intPtr(3), models.IPPoolKindWireGuard},
		{models.IPPoolPurposeKubernetesServices, cfg.KubernetesServiceCIDR, nil, models.IPPoolKindKubernetes},
	}

	for _, p := range pools {
		var existing models.IPPool
		err := database.DB.Where("purpose = ?", p.Purpose).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			pool := models.IPPool{
				Kind:      p.Kind,
				Purpose:   p.Purpose,
				CIDR:      p.CIDR,
				HubNumber: p.HubNumber,
			}
			if err := database.DB.Create(&pool).Error; err != nil {
				return fmt.Errorf("failed to create pool %s: %w", p.Purpose, err)
			}
			logger.Info("Created default IP pool", "purpose", p.Purpose, "cidr", p.CIDR)
		} else if err != nil {
			return fmt.Errorf("failed to check pool %s: %w", p.Purpose, err)
		}
	}

	return nil
}

func SetupNodeNetworking(node *models.Node) error {
	if err := EnsureDefaultPools(); err != nil {
		return fmt.Errorf("failed to ensure default pools: %w", err)
	}

	if node.Role == models.NodeRoleHub {
		if _, err := ensureHubNumber(node); err != nil {
			return fmt.Errorf("failed to assign hub number: %w", err)
		}
	}

	loopbackIP, err := allocateLoopbackIP(node)
	if err != nil {
		return fmt.Errorf("failed to allocate loopback IP: %w", err)
	}
	logger.Info("Allocated loopback IP", "node_id", node.ID, "ip", loopbackIP)

	if node.Role == models.NodeRoleWorker {
		if err := setupWorkerLinks(node); err != nil {
			return fmt.Errorf("failed to setup worker links: %w", err)
		}
	} else if node.Role == models.NodeRoleHub {
		if err := setupHubLinks(node); err != nil {
			return fmt.Errorf("failed to setup hub links: %w", err)
		}
	}

	return nil
}

func AssignHubNumbers() error {
	var hubs []models.Node
	if err := database.DB.Where("role = ?", models.NodeRoleHub).Order("id asc").Find(&hubs).Error; err != nil {
		return err
	}
	for i := range hubs {
		if _, err := ensureHubNumber(&hubs[i]); err != nil {
			return err
		}
	}
	return nil
}

func allocateLoopbackIP(node *models.Node) (string, error) {
	var pool models.IPPool
	if err := database.DB.Where("purpose = ?", models.IPPoolPurposeLoopback).First(&pool).Error; err != nil {
		return "", fmt.Errorf("loopback pool not found: %w", err)
	}

	var existing models.IPAllocation
	if err := database.DB.Where("pool_id = ? AND node_id = ? AND purpose = ?", pool.ID, node.ID, "loopback").First(&existing).Error; err == nil {
		return existing.IP, nil
	}

	var allocations []models.IPAllocation
	database.DB.Where("pool_id = ?", pool.ID).Find(&allocations)

	ip, err := findNextAvailableIP(pool.CIDR, allocations)
	if err != nil {
		return "", err
	}
	if ip == nil {
		return "", fmt.Errorf("loopback pool exhausted")
	}

	allocation := models.IPAllocation{
		PoolID:  pool.ID,
		NodeID:  &node.ID,
		IP:      *ip + "/32",
		Purpose: "loopback",
	}
	if err := database.DB.Create(&allocation).Error; err != nil {
		return "", err
	}

	return *ip, nil
}

func setupWorkerLinks(worker *models.Node) error {
	var hubs []models.Node
	if err := database.DB.Where("role = ?", models.NodeRoleHub).Find(&hubs).Error; err != nil {
		return err
	}

	if len(hubs) == 0 {
		logger.Warn("No hubs found, skipping worker link setup", "worker_id", worker.ID)
		return nil
	}

	for i := range hubs {
		if hubs[i].HubNumber == 0 {
			if num, err := ensureHubNumber(&hubs[i]); err == nil {
				hubs[i].HubNumber = num
			} else {
				return err
			}
		}
	}
	sort.Slice(hubs, func(i, j int) bool {
		return hubs[i].HubNumber < hubs[j].HubNumber
	})
	for _, hub := range hubs {
		if hub.HubNumber == 0 {
			continue
		}
		if err := createLink(&hub, worker); err != nil {
			return fmt.Errorf("failed to create link to hub %d: %w", hub.ID, err)
		}
	}

	return nil
}

func setupHubLinks(hub *models.Node) error {
	var existingHubs []models.Node
	if err := database.DB.Where("role = ? AND id != ?", models.NodeRoleHub, hub.ID).Find(&existingHubs).Error; err != nil {
		return err
	}

	hubNumber := hub.HubNumber
	if hubNumber == 0 {
		if num, err := ensureHubNumber(hub); err != nil {
			return err
		} else {
			hubNumber = num
		}
	}

	for _, otherHub := range existingHubs {
		if otherHub.HubNumber == 0 {
			if num, err := ensureHubNumber(&otherHub); err == nil {
				otherHub.HubNumber = num
			} else {
				return err
			}
		}
		if err := createHubToHubLink(hub, &otherHub); err != nil {
			return fmt.Errorf("failed to create hub-to-hub link: %w", err)
		}
	}

	var workers []models.Node
	if err := database.DB.Where("role = ?", models.NodeRoleWorker).Find(&workers).Error; err != nil {
		return err
	}

	for _, worker := range workers {
		if err := createLink(hub, &worker); err != nil {
			return fmt.Errorf("failed to create link to worker %d: %w", worker.ID, err)
		}
	}

	return nil
}

func createLink(hub *models.Node, worker *models.Node) error {
	hubNumber := hub.HubNumber
	if hubNumber < 1 || hubNumber > MaxHubs {
		return fmt.Errorf("invalid hub number %d for hub %d", hubNumber, hub.ID)
	}

	if _, err := allocateLoopbackIP(hub); err != nil {
		return fmt.Errorf("failed to ensure hub loopback IP: %w", err)
	}
	if _, err := allocateLoopbackIP(worker); err != nil {
		return fmt.Errorf("failed to ensure worker loopback IP: %w", err)
	}

	hubListenPort := hubWorkerListenPort(hubNumber, worker.ID)
	workerListenPort := 51820 + hubNumber - 1

	var purpose models.IPPoolPurpose
	switch hubNumber {
	case 1:
		purpose = models.IPPoolPurposeHub1Worker
	case 2:
		purpose = models.IPPoolPurposeHub2Worker
	case 3:
		purpose = models.IPPoolPurposeHub3Worker
	default:
		return fmt.Errorf("unsupported hub number %d", hubNumber)
	}

	var existingLink models.LinkAllocation
	if err := database.DB.Where("(node_a_id = ? AND node_b_id = ?) OR (node_a_id = ? AND node_b_id = ?)",
		hub.ID, worker.ID, worker.ID, hub.ID).First(&existingLink).Error; err == nil {
		logger.Info("Link already exists", "hub_id", hub.ID, "worker_id", worker.ID)

		hubLoopback, err := GetNodeLoopbackIP(hub.ID)
		if err != nil {
			return fmt.Errorf("failed to get hub loopback IP: %w", err)
		}
		workerLoopback, err := GetNodeLoopbackIP(worker.ID)
		if err != nil {
			return fmt.Errorf("failed to get worker loopback IP: %w", err)
		}

		hubIfaceName := fmt.Sprintf("wg-%s", worker.Hostname)
		workerIfaceName := fmt.Sprintf("wg-hub%d", hubNumber)

		var hubIface models.WireGuardInterface
		if err := database.DB.Where("node_id = ? AND name = ?", hub.ID, hubIfaceName).First(&hubIface).Error; err == nil {
			desired := fmt.Sprintf("%s/32, %s, 224.0.0.5/32", workerLoopback, existingLink.Subnet)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", hubIface.ID).Update("allowed_ips", desired)
			database.DB.Model(&hubIface).Update("listen_port", hubListenPort)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", hubIface.ID).
				Update("endpoint", fmt.Sprintf("%s:%d", worker.PublicIP, workerListenPort))
		}

		var workerIface models.WireGuardInterface
		if err := database.DB.Where("node_id = ? AND name = ?", worker.ID, workerIfaceName).First(&workerIface).Error; err == nil {
			desired := fmt.Sprintf("%s/32, %s, %s, 224.0.0.5/32", hubLoopback, existingLink.Subnet, config.Current().LoopbackCIDR)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", workerIface.ID).Update("allowed_ips", desired)
			database.DB.Model(&workerIface).Update("listen_port", workerListenPort)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", workerIface.ID).
				Update("endpoint", fmt.Sprintf("%s:%d", hub.PublicIP, hubListenPort))
		}

		return nil
	}

	var pool models.IPPool
	if err := database.DB.Where("purpose = ?", purpose).First(&pool).Error; err != nil {
		return fmt.Errorf("pool not found for purpose %s: %w", purpose, err)
	}

	subnet, hubIP, workerIP, err := allocateSubnet31(pool)
	if err != nil {
		return err
	}

	link := models.LinkAllocation{
		PoolID:  pool.ID,
		NodeAID: hub.ID,
		NodeBID: worker.ID,
		Subnet:  subnet,
		NodeAIP: hubIP,
		NodeBIP: workerIP,
	}
	if err := database.DB.Create(&link).Error; err != nil {
		return err
	}

	hubInterfaceName := fmt.Sprintf("wg-%s", worker.Hostname)

	hubInterface := models.WireGuardInterface{
		NodeID:     hub.ID,
		Name:       hubInterfaceName,
		Address:    hubIP + "/31",
		ListenPort: hubListenPort,
		Status:     models.InterfaceStatusDown,
	}
	if err := database.DB.Create(&hubInterface).Error; err != nil {
		return err
	}

	workerInterfaceName := fmt.Sprintf("wg-hub%d", hubNumber)

	workerInterface := models.WireGuardInterface{
		NodeID:     worker.ID,
		Name:       workerInterfaceName,
		Address:    workerIP + "/31",
		ListenPort: workerListenPort,
		Status:     models.InterfaceStatusDown,
	}
	if err := database.DB.Create(&workerInterface).Error; err != nil {
		return err
	}

	hubLoopback, err := GetNodeLoopbackIP(hub.ID)
	if err != nil {
		return fmt.Errorf("failed to get hub loopback IP: %w", err)
	}
	workerLoopback, err := GetNodeLoopbackIP(worker.ID)
	if err != nil {
		return fmt.Errorf("failed to get worker loopback IP: %w", err)
	}

	hubPeer := models.NodePeer{
		InterfaceID:         hubInterface.ID,
		PeerNodeID:          worker.ID,
		Endpoint:            fmt.Sprintf("%s:%d", worker.PublicIP, workerListenPort),
		AllowedIPs:          fmt.Sprintf("%s/32, %s, 224.0.0.5/32", workerLoopback, subnet),
		PersistentKeepAlive: 25,
		Status:              models.PeerStatusActive,
	}
	if err := database.DB.Create(&hubPeer).Error; err != nil {
		return err
	}

	workerPeer := models.NodePeer{
		InterfaceID:         workerInterface.ID,
		PeerNodeID:          hub.ID,
		Endpoint:            fmt.Sprintf("%s:%d", hub.PublicIP, hubListenPort),
		AllowedIPs:          fmt.Sprintf("%s/32, %s, %s, 224.0.0.5/32", hubLoopback, subnet, config.Current().LoopbackCIDR),
		PersistentKeepAlive: 0,
		Status:              models.PeerStatusActive,
	}
	if err := database.DB.Create(&workerPeer).Error; err != nil {
		return err
	}

	logger.Info("Created hub-worker link", "hub_id", hub.ID, "worker_id", worker.ID, "subnet", subnet)
	return nil
}

func createHubToHubLink(hubA *models.Node, hubB *models.Node) error {
	if _, err := allocateLoopbackIP(hubA); err != nil {
		return fmt.Errorf("failed to ensure hub loopback IP: %w", err)
	}
	if _, err := allocateLoopbackIP(hubB); err != nil {
		return fmt.Errorf("failed to ensure hub loopback IP: %w", err)
	}

	var existingLink models.LinkAllocation
	if err := database.DB.Where("(node_a_id = ? AND node_b_id = ?) OR (node_a_id = ? AND node_b_id = ?)",
		hubA.ID, hubB.ID, hubB.ID, hubA.ID).First(&existingLink).Error; err == nil {
		logger.Info("Hub-to-hub link already exists", "hub_a_id", hubA.ID, "hub_b_id", hubB.ID)
		if hubA.HubNumber == 0 || hubB.HubNumber == 0 {
			return fmt.Errorf("missing hub number for hub-to-hub link")
		}
		hubAPort := hubToHubListenPort(hubA.HubNumber, hubB.HubNumber)
		hubBPort := hubToHubListenPort(hubB.HubNumber, hubA.HubNumber)
		allowed := fmt.Sprintf("%s, %s, 224.0.0.5/32", existingLink.Subnet, config.Current().LoopbackCIDR)

		hubAInterfaceName := fmt.Sprintf("wg-%s", hubB.Hostname)
		var hubAInterface models.WireGuardInterface
		if err := database.DB.Where("node_id = ? AND name = ?", hubA.ID, hubAInterfaceName).First(&hubAInterface).Error; err == nil {
			database.DB.Model(&hubAInterface).Update("listen_port", hubAPort)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", hubAInterface.ID).
				Updates(map[string]any{
					"endpoint":    fmt.Sprintf("%s:%d", hubB.PublicIP, hubBPort),
					"allowed_ips": allowed,
				})
		}

		hubBInterfaceName := fmt.Sprintf("wg-%s", hubA.Hostname)
		var hubBInterface models.WireGuardInterface
		if err := database.DB.Where("node_id = ? AND name = ?", hubB.ID, hubBInterfaceName).First(&hubBInterface).Error; err == nil {
			database.DB.Model(&hubBInterface).Update("listen_port", hubBPort)
			database.DB.Model(&models.NodePeer{}).Where("interface_id = ?", hubBInterface.ID).
				Updates(map[string]any{
					"endpoint":    fmt.Sprintf("%s:%d", hubA.PublicIP, hubAPort),
					"allowed_ips": allowed,
				})
		}

		return nil
	}

	var pool models.IPPool
	if err := database.DB.Where("purpose = ?", models.IPPoolPurposeHubToHub).First(&pool).Error; err != nil {
		return fmt.Errorf("hub-to-hub pool not found: %w", err)
	}

	subnet, hubAIP, hubBIP, err := allocateSubnet31(pool)
	if err != nil {
		return err
	}

	link := models.LinkAllocation{
		PoolID:  pool.ID,
		NodeAID: hubA.ID,
		NodeBID: hubB.ID,
		Subnet:  subnet,
		NodeAIP: hubAIP,
		NodeBIP: hubBIP,
	}
	if err := database.DB.Create(&link).Error; err != nil {
		return err
	}

	if hubA.HubNumber == 0 || hubB.HubNumber == 0 {
		return fmt.Errorf("missing hub number for hub-to-hub link")
	}
	hubAPort := hubToHubListenPort(hubA.HubNumber, hubB.HubNumber)
	hubBPort := hubToHubListenPort(hubB.HubNumber, hubA.HubNumber)

	hubAInterfaceName := fmt.Sprintf("wg-%s", hubB.Hostname)
	hubAInterface := models.WireGuardInterface{
		NodeID:     hubA.ID,
		Name:       hubAInterfaceName,
		Address:    hubAIP + "/31",
		ListenPort: hubAPort,
		Status:     models.InterfaceStatusDown,
	}
	if err := database.DB.Create(&hubAInterface).Error; err != nil {
		return err
	}

	hubBInterfaceName := fmt.Sprintf("wg-%s", hubA.Hostname)
	hubBInterface := models.WireGuardInterface{
		NodeID:     hubB.ID,
		Name:       hubBInterfaceName,
		Address:    hubBIP + "/31",
		ListenPort: hubBPort,
		Status:     models.InterfaceStatusDown,
	}
	if err := database.DB.Create(&hubBInterface).Error; err != nil {
		return err
	}

	hubAPeer := models.NodePeer{
		InterfaceID:         hubAInterface.ID,
		PeerNodeID:          hubB.ID,
		Endpoint:            fmt.Sprintf("%s:%d", hubB.PublicIP, hubBPort),
		AllowedIPs:          fmt.Sprintf("%s, %s, 224.0.0.5/32", subnet, config.Current().LoopbackCIDR),
		PersistentKeepAlive: 0,
		Status:              models.PeerStatusActive,
	}
	if err := database.DB.Create(&hubAPeer).Error; err != nil {
		return err
	}

	hubBPeer := models.NodePeer{
		InterfaceID:         hubBInterface.ID,
		PeerNodeID:          hubA.ID,
		Endpoint:            fmt.Sprintf("%s:%d", hubA.PublicIP, hubAPort),
		AllowedIPs:          fmt.Sprintf("%s, %s, 224.0.0.5/32", subnet, config.Current().LoopbackCIDR),
		PersistentKeepAlive: 0,
		Status:              models.PeerStatusActive,
	}
	if err := database.DB.Create(&hubBPeer).Error; err != nil {
		return err
	}

	logger.Info("Created hub-to-hub link", "hub_a_id", hubA.ID, "hub_b_id", hubB.ID, "subnet", subnet)
	return nil
}

func allocateSubnet31(pool models.IPPool) (string, string, string, error) {
	var existingLinks []models.LinkAllocation
	database.DB.Where("pool_id = ?", pool.ID).Find(&existingLinks)

	allocated := make(map[string]bool)
	for _, link := range existingLinks {
		allocated[link.Subnet] = true
	}

	prefix, err := netip.ParsePrefix(pool.CIDR)
	if err != nil {
		return "", "", "", err
	}

	addr := prefix.Addr()
	for prefix.Contains(addr) {
		subnet := fmt.Sprintf("%s/31", addr.String())

		if !allocated[subnet] {
			lowerIP := addr.String()
			higherIP := addr.Next().String()
			return subnet, lowerIP, higherIP, nil
		}

		addr = addr.Next().Next()
	}

	return "", "", "", fmt.Errorf("pool %s exhausted", pool.CIDR)
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

func GetNodeLoopbackIP(nodeID uint) (string, error) {
	var allocation models.IPAllocation
	if err := database.DB.Where("node_id = ? AND purpose = ?", nodeID, "loopback").First(&allocation).Error; err != nil {
		return "", err
	}
	ip := allocation.IP
	if len(ip) > 3 && ip[len(ip)-3:] == "/32" {
		ip = ip[:len(ip)-3]
	}
	return ip, nil
}

func intPtr(i int) *int {
	return &i
}

func hubWorkerListenPort(hubNumber int, workerID uint) int {
	base := 52000 + (hubNumber-1)*1000
	return base + int(workerID)
}

func hubToHubListenPort(localHubNumber int, remoteHubNumber int) int {
	return 51820 + localHubNumber*10 + remoteHubNumber
}

func ensureHubNumber(node *models.Node) (int, error) {
	if node.Role != models.NodeRoleHub {
		return node.HubNumber, nil
	}

	var hubs []models.Node
	if err := database.DB.Where("role = ?", models.NodeRoleHub).Order("id asc").Find(&hubs).Error; err != nil {
		return 0, err
	}

	used := map[int]bool{}
	for _, h := range hubs {
		if h.HubNumber >= 1 && h.HubNumber <= MaxHubs {
			used[h.HubNumber] = true
		}
	}

	next := 1
	for i := range hubs {
		if hubs[i].HubNumber >= 1 && hubs[i].HubNumber <= MaxHubs {
			continue
		}
		for used[next] && next <= MaxHubs {
			next++
		}
		if next > MaxHubs {
			return 0, fmt.Errorf("max hubs reached")
		}
		if err := database.DB.Model(&models.Node{}).Where("id = ?", hubs[i].ID).Update("hub_number", next).Error; err != nil {
			return 0, err
		}
		hubs[i].HubNumber = next
		used[next] = true
		next++
	}

	for _, h := range hubs {
		if h.ID == node.ID {
			node.HubNumber = h.HubNumber
			return h.HubNumber, nil
		}
	}

	return node.HubNumber, nil
}
