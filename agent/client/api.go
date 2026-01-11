package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

const (
	AgentVersion = "0.0.1"
)

var ErrInvalidEnrollmentSecret = errors.New("invalid enrollment secret")

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		UserAgent: fmt.Sprintf("gluon-agent/%s (%s; %s)", AgentVersion, runtime.GOOS, runtime.GOARCH),
	}
}

func (c *Client) RequestEnrollment(hostname, provider, os, desiredRole string) (uint, string, error) {
	payload := map[string]string{
		"hostname":     hostname,
		"provider":     provider,
		"os":           os,
		"desired_role": desiredRole,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/enroll", bytes.NewBuffer(body))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("enrollment failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		RequestID        uint   `json:"request_id"`
		Message          string `json:"message"`
		EnrollmentSecret string `json:"enrollment_secret"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.RequestID, result.EnrollmentSecret, nil
}

func (c *Client) CheckEnrollmentStatus(requestID uint, enrollmentSecret string) (status string, nodeID uint, apiKey string, err error) {
	payload := map[string]interface{}{
		"request_id":        requestID,
		"enrollment_secret": enrollmentSecret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/enroll/status", bytes.NewBuffer(body))
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", 0, "", ErrInvalidEnrollmentSecret
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", 0, "", fmt.Errorf("status check failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		RequestID uint   `json:"request_id"`
		Status    string `json:"status"`
		NodeID    uint   `json:"node_id,omitempty"`
		APIKey    string `json:"api_key,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Status, result.NodeID, result.APIKey, nil
}

type NetworkInfo struct {
	NodeID             uint     `json:"node_id"`
	Role               string   `json:"role"`
	RequiredInterfaces []string `json:"required_interfaces"`
}

func (c *Client) GetNetworkInfo(apiKey string) (*NetworkInfo, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/agent/network/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get network info failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result NetworkInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) UploadPublicKeys(apiKey string, keys map[string]string) error {
	payload := map[string]interface{}{
		"keys": keys,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/network/keys", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload keys failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

type ConfigBundle struct {
	Version              int               `json:"version"`
	Hash                 string            `json:"hash"`
	WireGuardConfigs     map[string]string `json:"wireguard_configs"`
	NetworkInterfaceFile string            `json:"network_interface_file"`
	FRRConfigFile        string            `json:"frr_config_file"`
	SSHAuthorizedKeys    []SSHAuthorizedKey `json:"ssh_authorized_keys"`
}

type SSHAuthorizedKey struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

func (c *Client) GetConfig(apiKey string) (*ConfigBundle, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/agent/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get config failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result ConfigBundle
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) ReportConfigApplied(apiKey string, version int, hash string) error {
	payload := map[string]interface{}{
		"version": version,
		"hash":    hash,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/config/applied", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report config applied failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) ReportCommandResults(apiKey string, results []CommandResult) error {
	payload := map[string]any{
		"results": results,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/commands/report", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report command results failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

type KubernetesTask struct {
	Action               string `json:"action"`
	ControlPlaneEndpoint string `json:"control_plane_endpoint,omitempty"`
	PodCIDR              string `json:"pod_cidr,omitempty"`
	ServiceCIDR          string `json:"service_cidr,omitempty"`
	KubernetesVersion    string `json:"kubernetes_version,omitempty"`
	JoinCommand          string `json:"join_command,omitempty"`
	Note                 string `json:"note,omitempty"`
}

func (c *Client) GetKubernetesTask(apiKey string) (*KubernetesTask, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/agent/kubernetes/task", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get kubernetes task failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result KubernetesTask
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

type KubernetesReport struct {
	State string `json:"state"`

	Message string `json:"message,omitempty"`

	ControlPlaneEndpoint string `json:"control_plane_endpoint,omitempty"`
	PodCIDR              string `json:"pod_cidr,omitempty"`
	ServiceCIDR          string `json:"service_cidr,omitempty"`
	KubernetesVersion    string `json:"kubernetes_version,omitempty"`

	WorkerJoinCommand       string `json:"worker_join_command,omitempty"`
	ControlPlaneJoinCommand string `json:"control_plane_join_command,omitempty"`
	JoinCommandExpiresAt    string `json:"join_command_expires_at,omitempty"` 
}

func (c *Client) ReportKubernetes(apiKey string, report KubernetesReport) error {
	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/kubernetes/report", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report kubernetes failed: %s - %s", resp.Status, string(bodyBytes))
	}
	return nil
}
