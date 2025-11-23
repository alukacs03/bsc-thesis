package main

import (
	"context"
	"gluon-agent/client"
	"gluon-agent/config"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	log.Println("gluon-agent v0 starting up...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Determine config path
	configPath := getConfigPath()
	log.Printf("Using config file: %s", configPath)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure API URL is set
	if cfg.APIURL == "" {
		cfg.APIURL = getEnvOrDefault("GLUON_API_URL", "http://localhost:3000")
		log.Printf("API URL not in config, using: %s", cfg.APIURL)
	}

	log.Printf("System info: hostname=%s, os=%s, provider=%s", cfg.Hostname, cfg.OS, cfg.Provider)

	// Create API client
	apiClient := client.New(cfg.APIURL)

	// Check enrollment status
	if cfg.IsEnrolled() {
		log.Printf("Agent already enrolled (node_id=%s)", cfg.NodeID)
	} else if cfg.HasPendingEnrollment() {
		// Resume polling for existing enrollment request
		requestID, err := strconv.Atoi(cfg.RequestID)
		if err != nil {
			log.Fatalf("Invalid request_id in config: %v", err)
		}

		log.Printf("Resuming enrollment polling for request_id=%d", requestID)
		log.Println("Waiting for admin approval...")

		// Poll for approval
		if err := pollForApproval(ctx, apiClient, uint(requestID), cfg, configPath); err != nil {
			log.Fatalf("Enrollment failed: %v", err)
		}

		log.Printf("Enrollment successful! Node ID: %s", cfg.NodeID)
	} else {
		// No enrollment request exists - create new one
		log.Println("Agent not enrolled, starting enrollment process...")

		// Request enrollment using config values
		requestID, err := apiClient.RequestEnrollment(
			cfg.Hostname,
			cfg.Provider,
			cfg.OS,
			cfg.DesiredRole,
		)
		if err != nil {
			log.Fatalf("Enrollment request failed: %v", err)
		}

		cfg.RequestID = strconv.Itoa(int(requestID))
		if err := cfg.Save(configPath); err != nil {
			log.Printf("Warning: Failed to save request_id to config: %v", err)
		}

		log.Printf("Enrollment requested, request_id=%d", requestID)
		log.Println("Waiting for admin approval...")

		// Poll for approval
		if err := pollForApproval(ctx, apiClient, requestID, cfg, configPath); err != nil {
			log.Fatalf("Enrollment failed: %v", err)
		}

		log.Printf("Enrollment successful! Node ID: %s", cfg.NodeID)
	}

	// Start heartbeat loop
	log.Println("Starting heartbeat loop (30s interval)...")
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Send initial heartbeat
	if err := apiClient.Heartbeat(cfg.APIKey); err != nil {
		log.Printf("Initial heartbeat failed: %v", err)
	} else {
		log.Println("Initial heartbeat sent successfully")
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, exiting...")
			return

		case <-heartbeatTicker.C:
			if err := apiClient.Heartbeat(cfg.APIKey); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			} else {
				log.Println("Heartbeat sent")
			}
		}
	}
}

// pollForApproval polls the enrollment status until approved or rejected
func pollForApproval(ctx context.Context, apiClient *client.Client, requestID uint, cfg *config.Config, configPath string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			status, nodeID, apiKey, err := apiClient.CheckEnrollmentStatus(requestID)
			if err != nil {
				log.Printf("Status check failed: %v", err)
				continue
			}

			log.Printf("Enrollment status: %s", status)

			switch status {
			case "accepted":
				if apiKey != "" && nodeID > 0 {
					// First time receiving approval - save API key and node ID
					cfg.APIKey = apiKey
					cfg.NodeID = strconv.Itoa(int(nodeID))
					if err := cfg.Save(configPath); err != nil {
						return err
					}
					log.Printf("API key received and saved! Node ID: %d", nodeID)
					return nil
				}
				// Already enrolled, API key not returned again
				log.Println("Already enrolled (API key previously received)")
				return nil

			case "rejected":
				return nil

			case "pending":
				// Keep polling
				continue

			default:
				log.Printf("Unknown status: %s", status)
			}
		}
	}
}

// getConfigPath determines the config file location
func getConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("GLUON_CONFIG"); path != "" {
		return path
	}

	// Check if /etc/gluon exists (production)
	if stat, err := os.Stat("/etc/gluon"); err == nil && stat.IsDir() {
		return "/etc/gluon/agent.conf"
	}

	// Default to current directory (development)
	return "./agent.conf"
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
