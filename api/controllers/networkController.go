package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gluon-api/database"
	"gluon-api/generators"
	"gluon-api/logger"
	"gluon-api/models"
	"gluon-api/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetNetworkInfo(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	var interfaces []models.WireGuardInterface
	if err := database.DB.Where("node_id = ?", nodeID).Find(&interfaces).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get interfaces",
		})
	}

	requiredInterfaces := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		requiredInterfaces = append(requiredInterfaces, iface.Name)
	}

	return c.JSON(fiber.Map{
		"node_id":             nodeID,
		"role":                node.Role,
		"hub_number":          node.HubNumber,
		"required_interfaces": requiredInterfaces,
	})
}

func UploadPublicKeys(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	type keysInput struct {
		Keys map[string]string `json:"keys"`
	}

	var input keysInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	updated := 0
	for ifaceName, publicKey := range input.Keys {
		var iface models.WireGuardInterface
		if err := database.DB.Where("node_id = ? AND name = ?", nodeID, ifaceName).First(&iface).Error; err != nil {
			logger.Warn("Interface not found for key upload", "node_id", nodeID, "interface", ifaceName)
			continue
		}

		publicKey = stringsTrim(publicKey)
		if publicKey == "" || publicKey == iface.PublicKey {
			continue
		}

		if err := database.DB.Model(&models.WireGuardInterface{}).
			Where("id = ?", iface.ID).
			UpdateColumn("public_key", publicKey).Error; err != nil {
			logger.Error("Failed to save public key", "error", err, "node_id", nodeID, "interface", ifaceName)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save public key",
			})
		}
		updated++

		endpoint := fmt.Sprintf("%s:%d", node.PublicIP, iface.ListenPort)
		if err := database.DB.Model(&models.NodePeer{}).
			Where("peer_node_id = ? AND endpoint = ?", nodeID, endpoint).
			Update("peer_public_key", publicKey).Error; err != nil {
			logger.Error("Failed to update peer records with public key", "error", err, "node_id", nodeID)
		}
	}

	if updated > 0 {
		logger.Info("Public keys uploaded", "node_id", nodeID, "count", updated)
	}
	return c.JSON(fiber.Map{
		"message": "Public keys saved",
		"count":   updated,
	})
}

func GetConfig(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	var node models.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	var existingConfig models.NodeConfig
	hasExistingConfig := database.DB.Where("node_id = ?", nodeID).First(&existingConfig).Error == nil

	configBundle, err := generateConfigBundle(&node)
	if err != nil {
		logger.Error("Failed to generate config bundle", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate configuration",
		})
	}

	hash := calculateConfigHash(configBundle)

	version := 1
	if hasExistingConfig {
		if existingConfig.Hash == hash {
			sshKeysJSON := existingConfig.SSHAuthorizedKeys
			if stringsTrim(sshKeysJSON) == "" {
				sshKeysJSON = "[]"
			}
			return c.JSON(fiber.Map{
				"version":                 existingConfig.Version,
				"hash":                    existingConfig.Hash,
				"wireguard_configs":       json.RawMessage(existingConfig.WireGuardConfigs),
				"network_interface_file":  existingConfig.NetworkInterfaceConfig,
				"frr_config_file":         existingConfig.FRRConfig,
				"ssh_authorized_keys":     json.RawMessage(sshKeysJSON),
			})
		}
		version = existingConfig.Version + 1
	}

	wgConfigsJSON, _ := json.Marshal(configBundle.WireGuardConfigs)
	sshKeysJSON, _ := json.Marshal(configBundle.SSHAuthorizedKeys)

	newConfig := models.NodeConfig{
		NodeID:                 nodeID,
		Version:                version,
		WireGuardConfigs:       string(wgConfigsJSON),
		NetworkInterfaceConfig: configBundle.NetworkInterfaceFile,
		FRRConfig:              configBundle.FRRConfigFile,
		SSHAuthorizedKeys:      string(sshKeysJSON),
		Hash:                   hash,
		GeneratedAt:            time.Now(),
	}

	if hasExistingConfig {
		newConfig.ID = existingConfig.ID
		if err := database.DB.Save(&newConfig).Error; err != nil {
			logger.Error("Failed to save config", "error", err, "node_id", nodeID)
		}
	} else {
		if err := database.DB.Create(&newConfig).Error; err != nil {
			logger.Error("Failed to create config", "error", err, "node_id", nodeID)
		}
	}

	return c.JSON(fiber.Map{
		"version":                version,
		"hash":                   hash,
		"wireguard_configs":      configBundle.WireGuardConfigs,
		"network_interface_file": configBundle.NetworkInterfaceFile,
		"frr_config_file":        configBundle.FRRConfigFile,
		"ssh_authorized_keys":    configBundle.SSHAuthorizedKeys,
	})
}

func ReportConfigApplied(c *fiber.Ctx) error {
	nodeID := c.Locals("node_id").(uint)

	type appliedInput struct {
		Version int    `json:"version"`
		Hash    string `json:"hash"`
	}

	var input appliedInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	var config models.NodeConfig
	if err := database.DB.Where("node_id = ? AND version = ?", nodeID, input.Version).First(&config).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Config version not found",
		})
	}

	now := time.Now()
	config.AppliedAt = &now
	if err := database.DB.Save(&config).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update config status",
		})
	}

	logger.Info("Config applied by agent", "node_id", nodeID, "version", input.Version)
	return c.JSON(fiber.Map{
		"message": "Config status updated",
	})
}

type configBundle struct {
	WireGuardConfigs     map[string]string
	NetworkInterfaceFile string
	FRRConfigFile        string
	SSHAuthorizedKeys    []sshAuthorizedKey
}

type sshAuthorizedKey struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

func generateConfigBundle(node *models.Node) (*configBundle, error) {
	loopbackIP, err := services.GetNodeLoopbackIP(node.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get loopback IP: %w", err)
	}

	var interfaces []models.WireGuardInterface
	if err := database.DB.Where("node_id = ?", node.ID).Find(&interfaces).Error; err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	wgConfigs := make(map[string]string)
	networkInterfaces := make([]generators.NetworkInterface, 0)
	frrInterfaceNames := make([]string, 0)
	hubLinkInterfaces := make(map[string]bool)
	hubLinkPeerLoopbacks := make(map[string][]string)

	for _, iface := range interfaces {
		var peers []models.NodePeer
		if err := database.DB.Where("interface_id = ?", iface.ID).Preload("PeerNode").Find(&peers).Error; err != nil {
			return nil, fmt.Errorf("failed to get peers for interface %s: %w", iface.Name, err)
		}

		wgPeers := make([]generators.WireGuardPeer, 0, len(peers))
		for _, peer := range peers {
			if peer.PeerPublicKey == "" {
				continue
			}

			if peer.PeerNode.Role == models.NodeRoleHub {
				hubLinkInterfaces[iface.Name] = true
				if peerLB, err := services.GetNodeLoopbackIP(peer.PeerNode.ID); err == nil && stringsTrim(peerLB) != "" {
					hubLinkPeerLoopbacks[iface.Name] = append(hubLinkPeerLoopbacks[iface.Name], stringsTrim(peerLB))
				}
			}
			wgPeers = append(wgPeers, generators.WireGuardPeer{
				PublicKey:           peer.PeerPublicKey,
				Endpoint:            peer.Endpoint,
				AllowedIPs:          splitAllowedIPs(peer.AllowedIPs),
				PersistentKeepalive: peer.PersistentKeepAlive,
			})
		}

		wgConfig := generators.GenerateWireGuardConfig(iface.ListenPort, "", wgPeers)
		wgConfigs[iface.Name] = wgConfig

		postUp := []string{}
		preDown := []string{}
		if node.Role == models.NodeRoleHub && hubLinkInterfaces[iface.Name] {
			
			
			for _, peerLB := range hubLinkPeerLoopbacks[iface.Name] {
				postUp = append(postUp, fmt.Sprintf("/sbin/ip route replace %s/32 dev %s src %s", peerLB, iface.Name, loopbackIP))
				preDown = append(preDown, fmt.Sprintf("/sbin/ip route del %s/32 dev %s src %s || true", peerLB, iface.Name, loopbackIP))
			}
		}

		networkInterfaces = append(networkInterfaces, generators.NetworkInterface{
			Name:          iface.Name,
			Address:       iface.Address,
			WireGuardConf: fmt.Sprintf("/etc/wireguard/%s.conf", iface.Name),
			PostUpCommands: postUp,
			PreDownCommands: preDown,
		})

		frrInterfaceNames = append(frrInterfaceNames, iface.Name)
	}

	networkInterfaceFile := generators.GenerateNetworkInterfacesConfig(loopbackIP+"/32", networkInterfaces)

	var frrConfig string
	if node.Role == models.NodeRoleHub {
		var hubToHubInterfaces []string
		var workerInterfaces []string
		for _, ifaceName := range frrInterfaceNames {
			if hubLinkInterfaces[ifaceName] {
				hubToHubInterfaces = append(hubToHubInterfaces, ifaceName)
			} else {
				workerInterfaces = append(workerInterfaces, ifaceName)
			}
		}
		frrConfig = generators.GenerateFRRConfigForHub(node.Hostname, loopbackIP, hubToHubInterfaces, workerInterfaces)
	} else {
		frrConfig = generators.GenerateFRRConfigForWorker(node.Hostname, loopbackIP, frrInterfaceNames)
	}

	return &configBundle{
		WireGuardConfigs:     wgConfigs,
		NetworkInterfaceFile: networkInterfaceFile,
		FRRConfigFile:        frrConfig,
		SSHAuthorizedKeys:    loadSSHAuthorizedKeys(node.ID),
	}, nil
}

func calculateConfigHash(bundle *configBundle) string {
	h := sha256.New()

	wgJSON, _ := json.Marshal(bundle.WireGuardConfigs)
	h.Write(wgJSON)
	h.Write([]byte(bundle.NetworkInterfaceFile))
	h.Write([]byte(bundle.FRRConfigFile))
	sshJSON, _ := json.Marshal(bundle.SSHAuthorizedKeys)
	h.Write(sshJSON)

	return hex.EncodeToString(h.Sum(nil))
}

func loadSSHAuthorizedKeys(nodeID uint) []sshAuthorizedKey {
	var keys []models.NodeSSHAuthorizedKey
	if err := database.DB.Select("username", "public_key").Where("node_id = ?", nodeID).Find(&keys).Error; err != nil {
		return []sshAuthorizedKey{}
	}
	out := make([]sshAuthorizedKey, 0, len(keys))
	for _, k := range keys {
		u := stringsTrim(k.Username)
		p := stringsTrim(k.PublicKey)
		if u == "" || p == "" {
			continue
		}
		out = append(out, sshAuthorizedKey{Username: u, PublicKey: p})
	}
	sortSSHKeys(out)
	return out
}

func sortSSHKeys(keys []sshAuthorizedKey) {
	
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j].Username < keys[i].Username || (keys[j].Username == keys[i].Username && keys[j].PublicKey < keys[i].PublicKey) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

func splitAllowedIPs(allowedIPs string) []string {
	if allowedIPs == "" {
		return nil
	}
	result := make([]string, 0)
	for _, ip := range splitAndTrim(allowedIPs, ",") {
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}

func splitAndTrim(s string, sep string) []string {
	parts := make([]string, 0)
	for _, part := range stringsSplit(s, sep) {
		trimmed := stringsTrim(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func stringsSplit(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func stringsTrim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
