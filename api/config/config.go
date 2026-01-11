package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Settings struct {
	SecretKey              string
	LoopbackCIDR           string
	HubToHubCIDR           string
	Hub1WorkerCIDR         string
	Hub2WorkerCIDR         string
	Hub3WorkerCIDR         string
	KubernetesPodCIDR      string
	KubernetesServiceCIDR  string
	OSPFArea               int
	OSPFHelloInterval      int
	OSPFDeadInterval       int
	OSPFHubToHubCost       int
	OSPFHubToWorkerCost    int
	OSPFWorkerToHubCost    int
}

type Overrides struct {
	LoopbackCIDR           string
	HubToHubCIDR           string
	Hub1WorkerCIDR         string
	Hub2WorkerCIDR         string
	Hub3WorkerCIDR         string
	KubernetesPodCIDR      string
	KubernetesServiceCIDR  string
	OSPFArea               int
	OSPFHelloInterval      int
	OSPFDeadInterval       int
	OSPFHubToHubCost       int
	OSPFHubToWorkerCost    int
	OSPFWorkerToHubCost    int
}

var (
	current Settings
	mu      sync.RWMutex
)

func Load() error {
	cfg := Settings{
		SecretKey:             strings.TrimSpace(os.Getenv("GLUON_SECRET_KEY")),
		LoopbackCIDR:          envOrDefault("GLUON_LOOPBACK_CIDR", "10.255.0.0/22"),
		HubToHubCIDR:          envOrDefault("GLUON_HUB_TO_HUB_CIDR", "10.255.4.0/24"),
		Hub1WorkerCIDR:        envOrDefault("GLUON_HUB1_WORKER_CIDR", "10.255.8.0/22"),
		Hub2WorkerCIDR:        envOrDefault("GLUON_HUB2_WORKER_CIDR", "10.255.12.0/22"),
		Hub3WorkerCIDR:        envOrDefault("GLUON_HUB3_WORKER_CIDR", "10.255.16.0/22"),
		KubernetesPodCIDR:     envOrDefault("GLUON_K8S_POD_CIDR", "10.244.0.0/16"),
		KubernetesServiceCIDR: envOrDefault("GLUON_K8S_SERVICE_CIDR", "10.96.0.0/16"),
		OSPFArea:              envIntOrDefault("GLUON_OSPF_AREA", 10),
		OSPFHelloInterval:     envIntOrDefault("GLUON_OSPF_HELLO_INTERVAL", 1),
		OSPFDeadInterval:      envIntOrDefault("GLUON_OSPF_DEAD_INTERVAL", 3),
		OSPFHubToHubCost:      envIntOrDefault("GLUON_OSPF_HUB_TO_HUB_COST", 10),
		OSPFHubToWorkerCost:   envIntOrDefault("GLUON_OSPF_HUB_TO_WORKER_COST", 100),
		OSPFWorkerToHubCost:   envIntOrDefault("GLUON_OSPF_WORKER_TO_HUB_COST", 10),
	}

	if cfg.SecretKey == "" {
		cfg.SecretKey = strings.TrimSpace(os.Getenv("SECRET_KEY"))
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("secret key not configured")
	}

	mu.Lock()
	current = cfg
	mu.Unlock()
	return nil
}

func Current() Settings {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func ApplyOverrides(overrides Overrides) {
	mu.Lock()
	cfg := current
	if overrides.LoopbackCIDR != "" {
		cfg.LoopbackCIDR = overrides.LoopbackCIDR
	}
	if overrides.HubToHubCIDR != "" {
		cfg.HubToHubCIDR = overrides.HubToHubCIDR
	}
	if overrides.Hub1WorkerCIDR != "" {
		cfg.Hub1WorkerCIDR = overrides.Hub1WorkerCIDR
	}
	if overrides.Hub2WorkerCIDR != "" {
		cfg.Hub2WorkerCIDR = overrides.Hub2WorkerCIDR
	}
	if overrides.Hub3WorkerCIDR != "" {
		cfg.Hub3WorkerCIDR = overrides.Hub3WorkerCIDR
	}
	if overrides.KubernetesPodCIDR != "" {
		cfg.KubernetesPodCIDR = overrides.KubernetesPodCIDR
	}
	if overrides.KubernetesServiceCIDR != "" {
		cfg.KubernetesServiceCIDR = overrides.KubernetesServiceCIDR
	}
	if overrides.OSPFArea != 0 {
		cfg.OSPFArea = overrides.OSPFArea
	}
	if overrides.OSPFHelloInterval != 0 {
		cfg.OSPFHelloInterval = overrides.OSPFHelloInterval
	}
	if overrides.OSPFDeadInterval != 0 {
		cfg.OSPFDeadInterval = overrides.OSPFDeadInterval
	}
	if overrides.OSPFHubToHubCost != 0 {
		cfg.OSPFHubToHubCost = overrides.OSPFHubToHubCost
	}
	if overrides.OSPFHubToWorkerCost != 0 {
		cfg.OSPFHubToWorkerCost = overrides.OSPFHubToWorkerCost
	}
	if overrides.OSPFWorkerToHubCost != 0 {
		cfg.OSPFWorkerToHubCost = overrides.OSPFWorkerToHubCost
	}
	current = cfg
	mu.Unlock()
}

func envOrDefault(key string, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
