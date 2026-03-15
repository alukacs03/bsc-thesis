package results

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ExperimentResult holds the full record of a chaos experiment run.
type ExperimentResult struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Target         string     `json:"target"`
	Interface      string     `json:"interface,omitempty"`
	TTLSeconds     int        `json:"ttl_seconds"`
	StartedAt      time.Time  `json:"started_at"`
	FaultAppliedAt *time.Time `json:"fault_applied_at,omitempty"`
	FaultRevertedAt *time.Time `json:"fault_reverted_at,omitempty"`
	PingTarget     string     `json:"ping_target,omitempty"`
	FirstLossAt    *time.Time `json:"first_loss_at,omitempty"`
	FirstRecoveryAt *time.Time `json:"first_recovery_at,omitempty"`
	DowntimeMs     int64      `json:"downtime_ms"`
	PingSamples    int        `json:"ping_samples"`
	PacketsLost    int        `json:"packets_lost"`
	Status         string     `json:"status"`
}

// ExpID generates an experiment ID from the given start time.
// Format: exp-<YYYYMMDD>-<HHMMSS>
func ExpID(startedAt time.Time) string {
	return fmt.Sprintf("exp-%s", startedAt.UTC().Format("20060102-150405"))
}

// Save writes result as indented JSON to <dir>/<result.ID>.json, creating dir if needed.
func Save(result ExperimentResult, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("results: create dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("results: marshal: %w", err)
	}

	path := filepath.Join(dir, result.ID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("results: write %s: %w", path, err)
	}
	return nil
}

// List reads all *.json files from dir, unmarshals each as ExperimentResult,
// sorts by StartedAt descending, and returns the first n results.
// If dir is missing or empty, returns an empty slice (not an error).
func List(dir string, n int) ([]ExperimentResult, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("results: glob %s: %w", dir, err)
	}
	if len(entries) == 0 {
		return []ExperimentResult{}, nil
	}

	results := make([]ExperimentResult, 0, len(entries))
	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("results: read %s: %w", path, err)
		}
		var r ExperimentResult
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("results: unmarshal %s: %w", path, err)
		}
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StartedAt.After(results[j].StartedAt)
	})

	if n > 0 && n < len(results) {
		results = results[:n]
	}
	return results, nil
}
