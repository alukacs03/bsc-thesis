package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	assert.NoError(t, err)

	// Starts with glx_ prefix
	assert.True(t, len(key) >= 4 && key[:4] == "glx_", "key should start with glx_ prefix")

	// Total length: "glx_" (4) + 64 hex chars (32 bytes) = 68
	assert.Equal(t, 68, len(key), "key should be 68 characters total")

	// After prefix, only hex characters
	hexPart := key[4:]
	matched, err := regexp.MatchString("^[0-9a-f]+$", hexPart)
	assert.NoError(t, err)
	assert.True(t, matched, "characters after prefix should be lowercase hex")

	// Two consecutive calls produce different keys
	key2, err := GenerateAPIKey()
	assert.NoError(t, err)
	assert.NotEqual(t, key, key2, "two generated keys should be different")
}

func TestHashAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	assert.NoError(t, err)

	bcryptHash, hashIndex, err := HashAPIKey(key)
	assert.NoError(t, err)

	// bcrypt hash is verifiable against the plain key
	err = bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(key))
	assert.NoError(t, err, "bcrypt hash should be verifiable with the plain key")

	// hash index is 16 hex chars (first 8 bytes of SHA256 encoded as hex)
	assert.Equal(t, 16, len(hashIndex), "hash index should be 16 characters")
	matched, err := regexp.MatchString("^[0-9a-f]{16}$", hashIndex)
	assert.NoError(t, err)
	assert.True(t, matched, "hash index should be 16 lowercase hex characters")
}

func TestGenerateEnrollmentSecret(t *testing.T) {
	secret, err := GenerateEnrollmentSecret()
	assert.NoError(t, err)

	// Starts with es_ prefix
	assert.True(t, len(secret) >= 3 && secret[:3] == "es_", "secret should start with es_ prefix")

	// Total length: "es_" (3) + 64 hex chars (32 bytes) = 67
	assert.Equal(t, 67, len(secret), "secret should be 67 characters total")

	// After prefix, only hex characters
	hexPart := secret[3:]
	matched, err := regexp.MatchString("^[0-9a-f]+$", hexPart)
	assert.NoError(t, err)
	assert.True(t, matched, "characters after prefix should be lowercase hex")
}

func TestHashEnrollmentSecret(t *testing.T) {
	secret, err := GenerateEnrollmentSecret()
	assert.NoError(t, err)

	bcryptHash, hashIndex, err := HashEnrollmentSecret(secret)
	assert.NoError(t, err)

	// bcrypt hash is verifiable against the plain secret
	err = bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(secret))
	assert.NoError(t, err, "bcrypt hash should be verifiable with the plain secret")

	// hash index is 16 hex chars (first 8 bytes of SHA256 encoded as hex)
	assert.Equal(t, 16, len(hashIndex), "hash index should be 16 characters")
	matched, err := regexp.MatchString("^[0-9a-f]{16}$", hashIndex)
	assert.NoError(t, err)
	assert.True(t, matched, "hash index should be 16 lowercase hex characters")
}
