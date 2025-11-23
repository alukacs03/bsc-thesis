package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"gluon-api/database"
	"gluon-api/logger"
	"gluon-api/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GenerateAPIKey() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	plainKey := "glx_" + hex.EncodeToString(randomBytes)

	return plainKey, nil
}

func HashAPIKey(plainKey string) (bcryptHash string, hashIndex string, err error) {
	sha := sha256.Sum256([]byte(plainKey))
	hashIndex = hex.EncodeToString(sha[:8])

	hashed, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	bcryptHash = string(hashed)

	return bcryptHash, hashIndex, nil
}

func GetNodeIDFromAPIKeyHashIndex(hashIndex string) (uint, error) {
	var candidates []models.APIKey
	if err := database.DB.Where("hash_index = ? AND (expires_at IS NULL OR expires_at > ?) AND (revoked_at IS NULL)", hashIndex, time.Now()).Find(&candidates).Error; err != nil {
		logger.Error("Database error while fetching API keys: ", err)
	}
	if len(candidates) == 0 {
		return 0, nil // No matching API key found
	}
	return candidates[0].NodeID, nil
}
