package applier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gluon-agent/client"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	SSHStateFilePath  = "/var/lib/gluon/ssh-state.json"
	sshManagedBegin   = "# BEGIN GLUON MANAGED KEYS"
	sshManagedEnd     = "# END GLUON MANAGED KEYS"
)

type sshState struct {
	ManagedUsers map[string][]string `json:"managed_users"`
}

func applySSHAuthorizedKeys(keys []client.SSHAuthorizedKey) error {
	state, _ := loadSSHState()
	prevUsers := make(map[string]bool)
	for u := range state.ManagedUsers {
		prevUsers[u] = true
	}

	byUser := make(map[string][]string)
	for _, k := range keys {
		user := strings.TrimSpace(k.Username)
		line := strings.TrimSpace(k.PublicKey)
		if user == "" || line == "" {
			continue
		}
		byUser[user] = append(byUser[user], line)
	}

	
	for username, lines := range byUser {
		if err := ensureUserExists(username); err != nil {
			return err
		}
		homeDir, err := userHomeDir(username)
		if err != nil {
			return err
		}
		if err := reconcileAuthorizedKeys(username, homeDir, lines); err != nil {
			return err
		}
		delete(prevUsers, username)
	}

	
	for username := range prevUsers {
		if err := exec.Command("id", "-u", username).Run(); err != nil {
			
			continue
		}
		homeDir, err := userHomeDir(username)
		if err != nil {
			return err
		}
		if err := reconcileAuthorizedKeys(username, homeDir, nil); err != nil {
			return err
		}
	}

	
	next := make(map[string][]string)
	for username, lines := range byUser {
		clean := make([]string, 0, len(lines))
		seen := make(map[string]bool)
		for _, l := range lines {
			ll := strings.TrimSpace(l)
			if ll == "" || seen[ll] {
				continue
			}
			seen[ll] = true
			clean = append(clean, ll)
		}
		if len(clean) > 0 {
			next[username] = clean
		}
	}
	_ = saveSSHState(&sshState{ManagedUsers: next})

	return nil
}

func ensureUserExists(username string) error {
	if err := exec.Command("id", "-u", username).Run(); err == nil {
		return nil
	}
	
	return runCommand("useradd", "-m", "-s", "/bin/bash", username)
}

func userHomeDir(username string) (string, error) {
	if username == "root" {
		return "/root", nil
	}

	out, err := exec.Command("getent", "passwd", username).Output()
	if err != nil {
		return filepath.Join("/home", username), nil
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ":")
	if len(parts) >= 6 && parts[5] != "" {
		return parts[5], nil
	}
	return filepath.Join("/home", username), nil
}

func reconcileAuthorizedKeys(username string, homeDir string, publicKeys []string) error {
	sshDir := filepath.Join(homeDir, ".ssh")
	authKeysPath := filepath.Join(sshDir, "authorized_keys")

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create %s: %w", sshDir, err)
	}
	_ = os.Chmod(sshDir, 0700)

	f, err := os.OpenFile(authKeysPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", authKeysPath, err)
	}
	defer f.Close()
	_ = os.Chmod(authKeysPath, 0600)

	existingBytes, _ := os.ReadFile(authKeysPath)
	existingText := string(existingBytes)

	base := stripManagedBlock(existingText)

	desiredKeys := normalizeKeyLines(publicKeys)
	managed := renderManagedBlock(desiredKeys)

	baseTrim := strings.TrimRight(base, "\n")
	nextText := baseTrim
	if managed != "" {
		if baseTrim != "" {
			nextText = baseTrim + "\n\n" + managed
		} else {
			nextText = managed
		}
	}
	if !strings.HasSuffix(nextText, "\n") {
		nextText += "\n"
	 }

	if nextText != existingText {
		if err := os.WriteFile(authKeysPath, []byte(nextText), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", authKeysPath, err)
		}
	}

	
	_ = runCommand("chown", "-R", fmt.Sprintf("%s:%s", username, username), sshDir)
	return nil
}

func stripManagedBlock(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	inManaged := false
	for _, line := range lines {
		switch strings.TrimSpace(line) {
		case sshManagedBegin:
			inManaged = true
			continue
		case sshManagedEnd:
			inManaged = false
			continue
		}
		if inManaged {
			continue
		}
		out = append(out, line)
	}
	
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func renderManagedBlock(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString(sshManagedBegin)
	buf.WriteString("\n")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("\n")
	}
	buf.WriteString(sshManagedEnd)
	buf.WriteString("\n")
	return buf.String()
}

func normalizeKeyLines(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := make(map[string]bool)
	for _, k := range keys {
		line := strings.TrimSpace(k)
		line = strings.ReplaceAll(line, "\r\n", "\n")
		line = strings.ReplaceAll(line, "\n", " ")
		line = strings.Join(strings.Fields(line), " ")
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, line)
	}
	return out
}

func loadSSHState() (*sshState, error) {
	b, err := os.ReadFile(SSHStateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &sshState{ManagedUsers: map[string][]string{}}, nil
		}
		return nil, err
	}
	var s sshState
	if err := json.Unmarshal(b, &s); err != nil {
		return &sshState{ManagedUsers: map[string][]string{}}, nil
	}
	if s.ManagedUsers == nil {
		s.ManagedUsers = map[string][]string{}
	}
	return &s, nil
}

func saveSSHState(s *sshState) error {
	dir := filepath.Dir(SSHStateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if s.ManagedUsers == nil {
		s.ManagedUsers = map[string][]string{}
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(SSHStateFilePath, b, 0644)
}
