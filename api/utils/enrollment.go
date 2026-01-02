package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func GenerateEnrollmentSecret() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return "es_" + hex.EncodeToString(randomBytes), nil
}

func HashEnrollmentSecret(plainSecret string) (bcryptHash string, hashIndex string, err error) {
	sha := sha256.Sum256([]byte(plainSecret))
	hashIndex = hex.EncodeToString(sha[:8])

	hashed, err := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return string(hashed), hashIndex, nil
}
