package main

import (
	"context"
	"errors"
	"gluon-agent/applier"
	"gluon-agent/client"
	"gluon-agent/config"
	"gluon-agent/keys"
	"gluon-agent/kubernetes"
	"gluon-agent/pkgmgr"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	log.Println("gluon-agent v0 starting up...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	configPath := getConfigPath()
	log.Printf("Using config file: %s", configPath)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.APIURL == "" {
		cfg.APIURL = getEnvOrDefault("GLUON_API_URL", "http://localhost:3000")
		log.Printf("API URL not in config, using: %s", cfg.APIURL)
	}

	// Set default CA cert path if not configured
	if cfg.CACertPath == "" {
		cfg.CACertPath = getEnvOrDefault("GLUON_CA_CERT_PATH", "/etc/gluon/ca.crt")
	}

	if err := pkgmgr.EnsureDependencies(ctx); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	log.Printf("System info: hostname=%s, os=%s, provider=%s", cfg.Hostname, cfg.OS, cfg.Provider)

	// Bootstrap TLS: fetch CA certificate if needed and using HTTPS
	apiClient := createAPIClient(cfg, configPath)

	// Set up decommission handler
	client.DecommissionHandler = func() {
		log.Println("Executing decommission cleanup...")
		if err := applier.Cleanup(); err != nil {
			log.Printf("Cleanup error: %v", err)
		}
		if err := applier.ClearCredentials(configPath); err != nil {
			log.Printf("Failed to clear credentials: %v", err)
		}
	}

	if cfg.IsEnrolled() {
		log.Printf("Agent already enrolled (node_id=%s)", cfg.NodeID)
	} else if cfg.HasPendingEnrollment() {
		requestID, err := strconv.Atoi(cfg.RequestID)
		if err != nil {
			log.Fatalf("Invalid request_id in config: %v", err)
		}
		if cfg.EnrollmentSecret == "" {
			log.Fatalf("Enrollment secret missing for pending request_id=%d", requestID)
		}

		log.Printf("Resuming enrollment polling for request_id=%d", requestID)
		log.Println("Waiting for admin approval...")

		if err := pollForApproval(ctx, apiClient, uint(requestID), cfg, configPath); err != nil {
			log.Fatalf("Enrollment failed: %v", err)
		}

		log.Printf("Enrollment successful! Node ID: %s", cfg.NodeID)
	} else {
		log.Println("Agent not enrolled, starting enrollment process...")

		requestID, enrollmentSecret, err := apiClient.RequestEnrollment(
			cfg.Hostname,
			cfg.Provider,
			cfg.OS,
			cfg.DesiredRole,
		)
		if err != nil {
			log.Fatalf("Enrollment request failed: %v", err)
		}

		cfg.RequestID = strconv.Itoa(int(requestID))
		cfg.EnrollmentSecret = enrollmentSecret
		if err := cfg.Save(configPath); err != nil {
			log.Printf("Warning: Failed to save request_id to config: %v", err)
		}

		log.Printf("Enrollment requested, request_id=%d", requestID)
		log.Println("Waiting for admin approval...")

		if err := pollForApproval(ctx, apiClient, requestID, cfg, configPath); err != nil {
			log.Fatalf("Enrollment failed: %v", err)
		}

		log.Printf("Enrollment successful! Node ID: %s", cfg.NodeID)
	}

	heartbeatSecondsStr := getEnvOrDefault("GLUON_HEARTBEAT_INTERVAL_SECONDS", "30")
	heartbeatSeconds, err := strconv.Atoi(heartbeatSecondsStr)
	if err != nil || heartbeatSeconds <= 0 {
		heartbeatSeconds = 30
	}

	log.Printf("Starting heartbeat loop (%ds interval)...", heartbeatSeconds)
	heartbeatTicker := time.NewTicker(time.Duration(heartbeatSeconds) * time.Second)
	defer heartbeatTicker.Stop()

	go func() {
		if err := apiClient.Heartbeat(cfg.APIKey, cfg.DesiredRole); err != nil {
			log.Printf("Initial heartbeat failed: %v", err)
		} else {
			log.Println("Initial heartbeat sent successfully")
		}

		for {
			select {
			case <-ctx.Done():
				log.Println("Heartbeat goroutine exiting...")
				return
			case <-heartbeatTicker.C:
				if err := apiClient.Heartbeat(cfg.APIKey, cfg.DesiredRole); err != nil {
					log.Printf("Heartbeat failed: %v", err)
				} else {
					log.Println("Heartbeat sent")
				}
			}
		}
	}()

	log.Println("Starting config sync loop (60s interval)...")
	configTicker := time.NewTicker(60 * time.Second)
	defer configTicker.Stop()

	go func() {
		syncConfig(ctx, apiClient, cfg.APIKey)

		for {
			select {
			case <-ctx.Done():
				log.Println("Config sync goroutine exiting...")
				return
			case <-configTicker.C:
				syncConfig(ctx, apiClient, cfg.APIKey)
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, exiting...")

}

func pollForApproval(ctx context.Context, apiClient *client.Client, requestID uint, cfg *config.Config, configPath string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	if cfg.EnrollmentSecret == "" {
		return errors.New("enrollment secret not set in config")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			status, nodeID, apiKey, err := apiClient.CheckEnrollmentStatus(requestID, cfg.EnrollmentSecret)
			if err != nil {
				if errors.Is(err, client.ErrInvalidEnrollmentSecret) {
					return err
				}
				log.Printf("Status check failed: %v", err)
				continue
			}

			log.Printf("Enrollment status: %s", status)

			switch status {
			case "accepted":
				if apiKey != "" && nodeID > 0 {
					cfg.APIKey = apiKey
					cfg.NodeID = strconv.Itoa(int(nodeID))
					if err := cfg.Save(configPath); err != nil {
						return err
					}
					log.Printf("API key received and saved! Node ID: %d", nodeID)
					return nil
				}
				log.Println("Already enrolled (API key previously received)")
				return nil

			case "rejected":
				return nil

			case "pending":
				continue

			default:
				log.Printf("Unknown status: %s", status)
			}
		}
	}
}

func getConfigPath() string {
	if path := os.Getenv("GLUON_CONFIG"); path != "" {
		return path
	}

	if stat, err := os.Stat("/etc/gluon"); err == nil && stat.IsDir() {
		return "/etc/gluon/agent.conf"
	}

	return "./agent.conf"
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func createAPIClient(cfg *config.Config, configPath string) *client.Client {
	if strings.HasPrefix(cfg.APIURL, "http://") {
		httpsURL := "https://" + strings.TrimPrefix(cfg.APIURL, "http://")
		if trySwitchToHTTPS(cfg, configPath, httpsURL) {
			cfg.APIURL = httpsURL
		}
	}

	isHTTPS := strings.HasPrefix(cfg.APIURL, "https://")

	if !isHTTPS {
		log.Println("Using HTTP (no TLS)")
		return client.New(cfg.APIURL)
	}

	// Using HTTPS - need CA certificate
	if cfg.TLSSkipVerify {
		log.Println("WARNING: TLS verification disabled (development mode)")
		return client.NewWithOptions(cfg.APIURL, client.ClientOptions{
			TLSSkipVerify: true,
		})
	}

	if _, err := os.Stat(cfg.CACertPath); os.IsNotExist(err) {
		log.Printf("CA certificate not found at %s, fetching from API...", cfg.CACertPath)

		dir := cfg.CACertPath[:len(cfg.CACertPath)-len("/ca.crt")]
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Warning: failed to create directory %s: %v", dir, err)
			}
		}

		if err := client.FetchCACertificate(cfg.APIURL, cfg.CACertPath); err != nil {
			log.Printf("Failed to fetch CA certificate: %v", err)
			log.Println("Falling back to InsecureSkipVerify")
			return client.NewWithOptions(cfg.APIURL, client.ClientOptions{
				TLSSkipVerify: true,
			})
		}

		log.Printf("CA certificate saved to %s", cfg.CACertPath)

		if err := cfg.Save(configPath); err != nil {
			log.Printf("Warning: failed to save config: %v", err)
		}
	}

	log.Printf("Using HTTPS with CA certificate: %s", cfg.CACertPath)
	return client.NewWithOptions(cfg.APIURL, client.ClientOptions{
		CACertPath: cfg.CACertPath,
	})
}

func trySwitchToHTTPS(cfg *config.Config, configPath, httpsURL string) bool {
	dir := cfg.CACertPath[:len(cfg.CACertPath)-len("/ca.crt")]
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: failed to create directory %s: %v", dir, err)
		}
	}
	if err := client.FetchCACertificate(httpsURL, cfg.CACertPath); err != nil {
		log.Printf("HTTPS check failed: %v", err)
		return false
	}
	cfg.APIURL = httpsURL
	if err := cfg.Save(configPath); err != nil {
		log.Printf("Warning: failed to save config: %v", err)
	}
	return true
}

func syncConfig(ctx context.Context, apiClient *client.Client, apiKey string) {
	log.Println("Syncing configuration...")
	// Kubernetes bootstrap/join (single cluster) is driven by the API task endpoint,
	// and should run even when the network/config bundle is unchanged.
	defer kubernetes.Sync(ctx, apiClient, apiKey)

	networkInfo, err := apiClient.GetNetworkInfo(apiKey)
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
	} else if len(networkInfo.RequiredInterfaces) == 0 {
		log.Println("No network interfaces configured yet; skipping WireGuard key upload")
	} else {
		log.Printf("Required interfaces: %v", networkInfo.RequiredInterfaces)

		pubKeys, err := keys.EnsureKeys(networkInfo.RequiredInterfaces)
		if err != nil {
			log.Printf("Failed to ensure keys: %v", err)
		} else {
			state, _ := keys.LoadUploadState()
			if state != nil && keys.EqualPublicKeys(state.PublicKeys, pubKeys) {
				log.Printf("WireGuard public keys unchanged (%d); skipping upload", len(pubKeys))
			} else if err := apiClient.UploadPublicKeys(apiKey, pubKeys); err != nil {
				log.Printf("Failed to upload public keys: %v", err)
			} else {
				log.Printf("Uploaded %d WireGuard public keys", len(pubKeys))
				_ = keys.SaveUploadState(&keys.UploadState{PublicKeys: pubKeys})
			}
		}
	}

	configBundle, err := apiClient.GetConfig(apiKey)
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return
	}

	state, err := applier.LoadState()
	if err != nil {
		log.Printf("Failed to load state: %v", err)
		state = &applier.ConfigState{Version: 0}
	}

	if !applier.NeedsUpdate(configBundle, state) {
		log.Printf("Config is up to date (version %d)", state.Version)
		// Even if the bundle is unchanged, ensure interfaces are up (e.g., after reboot),
		// otherwise kubelet can fall back to the LAN IP and control-plane traffic may break.
		if networkInfo != nil && len(networkInfo.RequiredInterfaces) > 0 {
			applier.EnsureInterfacesUp(networkInfo.RequiredInterfaces)
		} else {
			applier.EnsureInterfacesUp(nil)
		}
		return
	}

	log.Printf("Config update needed: current=%d, new=%d", state.Version, configBundle.Version)

	if err := applier.ApplyConfig(configBundle); err != nil {
		log.Printf("Failed to apply config: %v", err)
		return
	}

	if err := apiClient.ReportConfigApplied(apiKey, configBundle.Version, configBundle.Hash); err != nil {
		log.Printf("Failed to report config applied: %v", err)
	}

	log.Println("Config sync completed successfully")
}
