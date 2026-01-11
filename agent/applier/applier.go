package applier

import (
	"encoding/json"
	"fmt"
	"gluon-agent/client"
	"gluon-agent/keys"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	WireGuardDir         = "/etc/wireguard"
	NetworkInterfacesDir = "/etc/network/interfaces.d"
	FRRConfigPath        = "/etc/frr/frr.conf"
	StateFilePath        = "/var/lib/gluon/config-state.json"
)

type ConfigState struct {
	Version int    `json:"version"`
	Hash    string `json:"hash"`
}

func LoadState() (*ConfigState, error) {
	data, err := os.ReadFile(StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigState{Version: 0}, nil
		}
		return nil, err
	}

	var state ConfigState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveState(state *ConfigState) error {
	dir := filepath.Dir(StateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(StateFilePath, data, 0644)
}

func ApplyConfig(bundle *client.ConfigBundle) error {
	log.Printf("Applying config version %d...", bundle.Version)

	networkTouched := false
	if len(bundle.WireGuardConfigs) > 0 {
		if err := applyWireGuardConfigs(bundle.WireGuardConfigs); err != nil {
			return fmt.Errorf("failed to apply WireGuard configs: %w", err)
		}
		networkTouched = true
	}

	if strings.TrimSpace(bundle.NetworkInterfaceFile) != "" {
		if err := applyNetworkInterfaces(bundle.NetworkInterfaceFile); err != nil {
			return fmt.Errorf("failed to apply network interfaces: %w", err)
		}
		networkTouched = true
	}

	if strings.TrimSpace(bundle.FRRConfigFile) != "" {
		if err := applyFRRConfig(bundle.FRRConfigFile); err != nil {
			return fmt.Errorf("failed to apply FRR config: %w", err)
		}
		networkTouched = true
	}

	if err := applySSHAuthorizedKeys(bundle.SSHAuthorizedKeys); err != nil {
		return fmt.Errorf("failed to apply SSH keys: %w", err)
	}

	if networkTouched {
		if err := bringUpInterfaces(); err != nil {
			return fmt.Errorf("failed to bring up interfaces: %w", err)
		}

		if err := reloadFRR(); err != nil {
			return fmt.Errorf("failed to reload FRR: %w", err)
		}
	}

	state := &ConfigState{
		Version: bundle.Version,
		Hash:    bundle.Hash,
	}
	if err := SaveState(state); err != nil {
		log.Printf("Warning: failed to save state: %v", err)
	}

	log.Printf("Config version %d applied successfully", bundle.Version)
	return nil
}




func EnsureInterfacesUp(requiredInterfaces []string) {
	ifaces := normalizeInterfaceList(requiredInterfaces)
	if len(ifaces) == 0 {
		return
	}

	
	ensureIfUp("dummy")

	for _, iface := range ifaces {
		if strings.TrimSpace(iface) == "" {
			continue
		}
		ensureIfUp(iface)
	}
}

func normalizeInterfaceList(requiredInterfaces []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(requiredInterfaces))

	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}

	for _, n := range requiredInterfaces {
		add(n)
	}

	
	if len(out) == 0 {
		files, _ := filepath.Glob(filepath.Join(WireGuardDir, "wg-*.conf"))
		for _, f := range files {
			add(strings.TrimSuffix(filepath.Base(f), ".conf"))
		}
	}

	return out
}

func ensureIfUp(iface string) {
	up, err := isLinkUp(iface)
	if err == nil && up {
		return
	}

	if err := runCommand("ifup", iface); err != nil {
		
		log.Printf("Warning: failed to bring up %s: %v", iface, err)
		return
	}
	log.Printf("Brought up interface: %s", iface)
}

func isLinkUp(iface string) (bool, error) {
	out, err := exec.Command("ip", "-o", "link", "show", "dev", iface).CombinedOutput()
	if err != nil {
		return false, err
	}
	
	re := regexp.MustCompile(`<[^>]*\bUP\b[^>]*>`)
	return re.Match(out), nil
}

func applyWireGuardConfigs(configs map[string]string) error {
	if err := os.MkdirAll(WireGuardDir, 0700); err != nil {
		return err
	}

	for ifaceName, configContent := range configs {
		privateKey, err := keys.GetPrivateKey(ifaceName)
		if err != nil {
			return fmt.Errorf("failed to get private key for %s: %w", ifaceName, err)
		}

		finalConfig := strings.Replace(configContent, "PrivateKey = PRIVATE_KEY_PLACEHOLDER", "PrivateKey = "+privateKey, 1)

		configPath := filepath.Join(WireGuardDir, ifaceName+".conf")
		if err := os.WriteFile(configPath, []byte(finalConfig), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", configPath, err)
		}
		log.Printf("Wrote WireGuard config: %s", configPath)
	}

	return nil
}

func applyNetworkInterfaces(content string) error {
	if err := os.MkdirAll(NetworkInterfacesDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(NetworkInterfacesDir, "gluon")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return err
	}
	log.Printf("Wrote network interfaces config: %s", configPath)
	return nil
}

func applyFRRConfig(content string) error {
	if err := os.WriteFile(FRRConfigPath, []byte(content), 0640); err != nil {
		return err
	}
	log.Printf("Wrote FRR config: %s", FRRConfigPath)
	return nil
}

func bringUpInterfaces() error {
	log.Println("Bringing down existing interfaces...")
	exec.Command("ifdown", "--force", "dummy").Run()

	files, _ := filepath.Glob(filepath.Join(WireGuardDir, "wg-*.conf"))
	for _, f := range files {
		ifaceName := strings.TrimSuffix(filepath.Base(f), ".conf")
		exec.Command("ifdown", "--force", ifaceName).Run()
		exec.Command("ip", "link", "delete", ifaceName).Run()
	}

	exec.Command("ip", "link", "delete", "dummy").Run()

	log.Println("Bringing up interfaces...")

	if err := runCommand("ifup", "dummy"); err != nil {
		return fmt.Errorf("failed to bring up dummy: %w", err)
	}

	for _, f := range files {
		ifaceName := strings.TrimSuffix(filepath.Base(f), ".conf")
		if err := runCommand("ifup", ifaceName); err != nil {
			return fmt.Errorf("failed to bring up %s: %w", ifaceName, err)
		}
	}

	return nil
}

func reloadFRR() error {
	log.Println("Reloading FRR...")
	return runCommand("systemctl", "reload", "frr")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func NeedsUpdate(bundle *client.ConfigBundle, state *ConfigState) bool {
	if bundle.Version > state.Version {
		return true
	}
	if bundle.Hash != state.Hash {
		return true
	}
	return false
}
