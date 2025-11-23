package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

const (
	AgentVersion = "0.0.1"
)

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

// RequestEnrollment sends enrollment request to API
// POST /api/agent/enroll
func (c *Client) RequestEnrollment(hostname, provider, os, desiredRole string) (uint, error) {
	payload := map[string]string{
		"hostname":     hostname,
		"provider":     provider,
		"os":           os,
		"desired_role": desiredRole,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/enroll", bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("enrollment failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		RequestID uint   `json:"request_id"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.RequestID, nil
}

// CheckEnrollmentStatus polls for enrollment approval
// POST /api/agent/enroll/status
func (c *Client) CheckEnrollmentStatus(requestID uint) (status string, nodeID uint, apiKey string, err error) {
	payload := map[string]uint{
		"request_id": requestID,
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

// Heartbeat sends heartbeat to API
// POST /api/agent/heartbeat
// Requires Authorization header with API key
func (c *Client) Heartbeat(apiKey string) error {
	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: API key may be revoked")
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}
