package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	caValidityYears     = 10
	serverValidityYears = 1
	keySize             = 4096
)

// EnsureCA loads existing CA or generates a new one if it doesn't exist.
// Returns the CA certificate and private key.
func EnsureCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Try to load existing CA
	if cert, key, err := LoadCA(certPath, keyPath); err == nil {
		return cert, key, nil
	}

	// Generate new CA
	return GenerateCA(certPath, keyPath)
}

// LoadCA loads an existing CA certificate and private key from files.
func LoadCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA key: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA key: %w", err)
	}

	return cert, key, nil
}

// GenerateCA creates a new self-signed CA certificate and saves it to files.
func GenerateCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Create CA certificate template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Gluon"},
			OrganizationalUnit: []string{"Mesh Network"},
			CommonName:         "Gluon Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(caValidityYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse the certificate back to get the x509.Certificate struct
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse generated CA certificate: %w", err)
	}

	// Write certificate to file
	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return nil, nil, fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Write key to file (with restricted permissions)
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return nil, nil, fmt.Errorf("failed to write CA key: %w", err)
	}

	return cert, key, nil
}

// GenerateServerCert creates a new server certificate signed by the CA.
// The hosts parameter can include both DNS names and IP addresses.
func GenerateServerCert(ca *x509.Certificate, caKey *rsa.PrivateKey, hosts []string) (certPEM, keyPEM []byte, err error) {
	// Generate server private key
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Separate DNS names and IP addresses
	var dnsNames []string
	var ipAddresses []net.IP
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, host)
		}
	}

	// Create server certificate template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Gluon"},
			OrganizationalUnit: []string{"Mesh Network"},
			CommonName:         "Gluon API Server",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(serverValidityYears, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
	}

	// Sign with CA
	certDER, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode key to PEM
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return certPEM, keyPEM, nil
}

// SaveServerCert writes the server certificate and key to files.
func SaveServerCert(certPEM, keyPEM []byte, certPath, keyPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to write server certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write server key: %w", err)
	}

	return nil
}

// GetCAFingerprint returns the SHA256 fingerprint of the CA certificate.
// This can be used for trust-on-first-use verification.
func GetCAFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%X", cert.Raw[:32])
}
