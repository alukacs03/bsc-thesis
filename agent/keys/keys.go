package keys

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	KeysDir = "/etc/wireguard/keys"
)

func EnsureKeysDir() error {
	return os.MkdirAll(KeysDir, 0700)
}

func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privCmd := exec.Command("wg", "genkey")
	privOut, err := privCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOut))

	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privateKey)
	pubOut, err := pubCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOut))

	return privateKey, publicKey, nil
}

func EnsureKeys(requiredInterfaces []string) (map[string]string, error) {
	if err := EnsureKeysDir(); err != nil {
		return nil, fmt.Errorf("failed to create keys directory: %w", err)
	}

	pubKeys := make(map[string]string)

	for _, iface := range requiredInterfaces {
		keyPath := filepath.Join(KeysDir, iface+".key")
		pubPath := filepath.Join(KeysDir, iface+".pub")

		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			privKey, pubKey, err := GenerateKeyPair()
			if err != nil {
				return nil, fmt.Errorf("failed to generate keypair for %s: %w", iface, err)
			}

			if err := os.WriteFile(keyPath, []byte(privKey+"\n"), 0600); err != nil {
				return nil, fmt.Errorf("failed to write private key for %s: %w", iface, err)
			}

			if err := os.WriteFile(pubPath, []byte(pubKey+"\n"), 0644); err != nil {
				return nil, fmt.Errorf("failed to write public key for %s: %w", iface, err)
			}

			pubKeys[iface] = pubKey
		} else {
			pubData, err := os.ReadFile(pubPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read public key for %s: %w", iface, err)
			}
			pubKeys[iface] = strings.TrimSpace(string(pubData))
		}
	}

	return pubKeys, nil
}

func GetPrivateKey(iface string) (string, error) {
	keyPath := filepath.Join(KeysDir, iface+".key")
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key for %s: %w", iface, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func GetPublicKey(iface string) (string, error) {
	pubPath := filepath.Join(KeysDir, iface+".pub")
	data, err := os.ReadFile(pubPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key for %s: %w", iface, err)
	}
	return strings.TrimSpace(string(data)), nil
}
