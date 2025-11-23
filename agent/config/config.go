package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	APIURL      string `json:"api_url"`
	APIKey      string `json:"api_key"`
	NodeID      string `json:"node_id"`
	RequestID   string `json:"request_id"`
	Hostname    string `json:"hostname"`
	Provider    string `json:"provider"`
	OS          string `json:"os"`
	DesiredRole string `json:"desired_role"`
}

func Load(path string) (*Config, error) {
	// create defaults
	// before everything starts NodeID, APIKey, RequestID are surely empty
	// the rest can be set, OR we can gather the info by default --> we are running debian (12 or 13)
	// at first, try to load from agent.conf
	var osystem string
	cmd := exec.Command("lsb_release", "-si")
	output, err := cmd.Output()
	if err != nil {
		osystem = "unknown"
	} else {
		distro := strings.TrimSpace(string(output))
		cmd2 := exec.Command("lsb_release", "-sr")
		output2, err2 := cmd2.Output()
		if err2 != nil {
			osystem = distro
		} else {
			release := strings.TrimSpace(string(output2))
			osystem = distro + " " + release
		}
	}
	var hostname string
	cmd3 := exec.Command("hostname", "-f")
	output3, err3 := cmd3.Output()
	if err3 != nil {
		hostname = "unknown"
	} else {
		hostname = strings.TrimSpace(string(output3))
	}
	cfg := &Config{
		Provider:    "unset",
		OS:          osystem,
		Hostname:    hostname,
		DesiredRole: "worker",
	}
	file, err := os.Open(path)
	if err != nil {
		return cfg, nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) IsEnrolled() bool {
	return c.NodeID != "" && c.APIKey != ""
}

func (c *Config) HasPendingEnrollment() bool {
	return c.RequestID != "" && c.APIKey == ""
}
