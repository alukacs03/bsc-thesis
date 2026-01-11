package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gluon-api/database"
	"gluon-api/models"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/ssh"
)

var linuxUsernameRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]*[$]?$`)

func ListNodeSSHKeys(c *fiber.Ctx) error {
	id := c.Params("id")

	var node models.Node
	if err := database.DB.Select("id").First(&node, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var keys []models.NodeSSHAuthorizedKey
	if err := database.DB.Where("node_id = ?", node.ID).Order("username asc, id asc").Find(&keys).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve SSH keys"})
	}

	return c.JSON(keys)
}

func CreateNodeSSHKey(c *fiber.Ctx) error {
	id := c.Params("id")

	nodeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil || nodeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node id"})
	}

	var node models.Node
	if err := database.DB.Select("id").First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var input struct {
		Username  string `json:"username"`
		PublicKey string `json:"public_key"`
		Comment   string `json:"comment"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	username := strings.TrimSpace(input.Username)
	if username == "" || !linuxUsernameRe.MatchString(username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid username"})
	}

	line := normalizeAuthorizedKeyLine(input.PublicKey, input.Comment)
	if line == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid public key"})
	}

	
	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid SSH public key format"})
	}

	var existing models.NodeSSHAuthorizedKey
	if err := database.DB.
		Where("node_id = ? AND username = ? AND public_key = ?", node.ID, username, line).
		First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "SSH key already exists"})
	}

	record := models.NodeSSHAuthorizedKey{
		NodeID:    node.ID,
		Username:  username,
		PublicKey: line,
		Comment:   strings.TrimSpace(input.Comment),
	}

	if err := database.DB.Create(&record).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store SSH key"})
	}

	return c.Status(fiber.StatusCreated).JSON(record)
}

func DeleteNodeSSHKey(c *fiber.Ctx) error {
	id := c.Params("id")
	keyID := c.Params("keyId")

	nodeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil || nodeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node id"})
	}
	kid, err := strconv.ParseUint(keyID, 10, 64)
	if err != nil || kid == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid key id"})
	}

	var record models.NodeSSHAuthorizedKey
	if err := database.DB.Where("node_id = ? AND id = ?", uint(nodeID), uint(kid)).First(&record).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SSH key not found"})
	}

	if err := database.DB.Delete(&record).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete SSH key"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GenerateNodeSSHKey(c *fiber.Ctx) error {
	id := c.Params("id")

	nodeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil || nodeID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node id"})
	}

	var node models.Node
	if err := database.DB.Select("id").First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Node not found"})
	}

	var input struct {
		Username string `json:"username"`
		Comment  string `json:"comment"`
		Bits     int    `json:"bits"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	username := strings.TrimSpace(input.Username)
	if username == "" || !linuxUsernameRe.MatchString(username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid username"})
	}

	bits := input.Bits
	if bits == 0 {
		bits = 4096
	}
	if bits < 2048 || bits > 8192 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bits must be between 2048 and 8192"})
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate keypair"})
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode public key"})
	}

	comment := strings.TrimSpace(input.Comment)
	pubLine := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))
	if comment != "" {
		pubLine = fmt.Sprintf("%s %s", pubLine, comment)
	}

	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})

	record := models.NodeSSHAuthorizedKey{
		NodeID:    node.ID,
		Username:  username,
		PublicKey: pubLine,
		Comment:   comment,
	}
	if err := database.DB.Create(&record).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store SSH key"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":              record.ID,
		"node_id":         record.NodeID,
		"username":        record.Username,
		"public_key":      record.PublicKey,
		"private_key_pem": string(privPEM),
	})
}

func normalizeAuthorizedKeyLine(publicKey string, comment string) string {
	line := strings.TrimSpace(publicKey)
	line = strings.ReplaceAll(line, "\r\n", "\n")
	line = strings.ReplaceAll(line, "\n", " ")
	line = strings.Join(strings.Fields(line), " ")
	if line == "" {
		return ""
	}

	
	if strings.TrimSpace(comment) != "" {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			line = fmt.Sprintf("%s %s", line, strings.TrimSpace(comment))
		}
	}
	return line
}
