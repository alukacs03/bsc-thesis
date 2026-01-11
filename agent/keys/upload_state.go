package keys

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const UploadStatePath = "/var/lib/gluon/wg-keys-state.json"

type UploadState struct {
	PublicKeys map[string]string `json:"public_keys"`
}

func LoadUploadState() (*UploadState, error) {
	b, err := os.ReadFile(UploadStatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &UploadState{PublicKeys: map[string]string{}}, nil
		}
		return nil, err
	}

	var s UploadState
	if err := json.Unmarshal(b, &s); err != nil {
		return &UploadState{PublicKeys: map[string]string{}}, nil
	}
	if s.PublicKeys == nil {
		s.PublicKeys = map[string]string{}
	}
	return &s, nil
}

func SaveUploadState(state *UploadState) error {
	dir := filepath.Dir(UploadStatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if state.PublicKeys == nil {
		state.PublicKeys = map[string]string{}
	}
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(UploadStatePath, b, 0644)
}

func EqualPublicKeys(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		if bv, ok := b[k]; !ok || bv != av {
			return false
		}
	}
	return true
}

