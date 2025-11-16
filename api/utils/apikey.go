package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

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
