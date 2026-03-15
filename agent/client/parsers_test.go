package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
		ok    bool
	}{
		{
			name:  "day hours seconds",
			input: "1 day, 2 hours, 30 seconds",
			want:  time.Duration(93630) * time.Second,
			ok:    true,
		},
		{
			name:  "minutes only",
			input: "5 minutes",
			want:  5 * time.Minute,
			ok:    true,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
			ok:    false,
		},
		{
			name:  "garbage input",
			input: "not a duration at all",
			want:  0,
			ok:    false,
		},
		{
			name:  "single second",
			input: "1 second",
			want:  1 * time.Second,
			ok:    true,
		},
		{
			name:  "zero hours",
			input: "0 hours",
			want:  0,
			ok:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseHumanDuration(tt.input)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseSizeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint64
		ok    bool
	}{
		{
			name:  "megabytes",
			input: "5 MB",
			want:  5000000,
			ok:    true,
		},
		{
			name:  "gibibytes",
			input: "1 GiB",
			want:  1073741824,
			ok:    true,
		},
		{
			name:  "zero bytes",
			input: "0 B",
			want:  0,
			ok:    true,
		},
		{
			name:  "kibibytes",
			input: "100 KiB",
			want:  102400,
			ok:    true,
		},
		{
			name:  "garbage input",
			input: "lots of space",
			want:  0,
			ok:    false,
		},
		{
			name:  "no unit",
			input: "1234",
			want:  0,
			ok:    false,
		},
		{
			name:  "unknown unit",
			input: "5 PB",
			want:  0,
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseSizeBytes(tt.input)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNormalizeOSPFState(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "full with DR", input: "Full/Dr", want: "Full"},
		{name: "full with BDR", input: "Full/BDR", want: "Full"},
		{name: "init", input: "Init", want: "Init"},
		{name: "two way", input: "2-Way", want: "2-Way"},
		{name: "down", input: "Down", want: "Down"},
		{name: "empty", input: "", want: ""},
		{name: "exstart", input: "ExStart", want: "ExStart"},
		{name: "exchange", input: "Exchange", want: "Exchange"},
		{name: "loading", input: "Loading", want: "Loading"},
		{name: "full with DROther", input: "Full/DROther", want: "Full"},
		{name: "2way alt", input: "2way", want: "2-Way"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeOSPFState(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseOSPFNeighborsJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		check     func(t *testing.T, result []ospfNeighborRaw)
	}{
		{
			name: "FRR format with router ID keys",
			input: `{
				"10.255.0.1": [
					{
						"nbrState": "Full/DR",
						"ifaceName": "wg-hub2"
					}
				],
				"10.255.0.3": [
					{
						"nbrState": "Init",
						"ifaceName": "wg-spoke1",
						"priority": 0
					}
				]
			}`,
			wantCount: 2,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				byRouter := map[string]ospfNeighborRaw{}
				for _, r := range result {
					byRouter[r.RouterID] = r
				}
				r1 := byRouter["10.255.0.1"]
				assert.Equal(t, "Full", r1.State)
				assert.Equal(t, "wg-hub2", r1.Interface)

				r2 := byRouter["10.255.0.3"]
				assert.Equal(t, "Init", r2.State)
				assert.Equal(t, "wg-spoke1", r2.Interface)
				assert.NotNil(t, r2.Priority)
				assert.Equal(t, uint64(0), *r2.Priority)
			},
		},
		{
			name: "single neighbor object (not array)",
			input: `{
				"10.255.0.5": {
					"nbrState": "Full/BDR",
					"ifaceName": "wg0"
				}
			}`,
			wantCount: 1,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Equal(t, "10.255.0.5", result[0].RouterID)
				assert.Equal(t, "Full", result[0].State)
				assert.Equal(t, "wg0", result[0].Interface)
			},
		},
		{
			name: "array format with routerId",
			input: `{
				"neighbors": [
					{
						"neighborId": "10.255.0.1",
						"state": "Full/DR",
						"ifaceName": "wg-hub2",
						"priority": 1
					}
				]
			}`,
			wantCount: 1,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Equal(t, "10.255.0.1", result[0].RouterID)
				assert.Equal(t, "Full", result[0].State)
				assert.Equal(t, "wg-hub2", result[0].Interface)
				assert.NotNil(t, result[0].Priority)
				assert.Equal(t, uint64(1), *result[0].Priority)
			},
		},
		{
			name:      "invalid JSON",
			input:     `not json at all`,
			wantCount: 0,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Nil(t, result)
			},
		},
		{
			name:      "empty object",
			input:     `{}`,
			wantCount: 0,
			check:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOSPFNeighborsJSON([]byte(tt.input))
			if tt.check != nil {
				tt.check(t, result)
			} else {
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestParseOSPFNeighborsText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		check     func(t *testing.T, result []ospfNeighborRaw)
	}{
		{
			name: "typical vtysh output",
			input: "Neighbor ID     Pri   State           Up Time         Dead Time       Address         Interface\n" +
				"10.255.0.1        1   Full/DR         1d02h           00:00:35        10.0.12.1       wg-hub2:10.0.12.2\n" +
				"10.255.0.3        0   2-Way           00:05:12        00:00:38        10.0.13.1       wg-spoke1:10.0.13.2\n",
			wantCount: 2,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Equal(t, "10.255.0.1", result[0].RouterID)
				assert.Equal(t, "Full", result[0].State)
				assert.Equal(t, "wg-hub2:10.0.12.2", result[0].Interface)
				assert.NotNil(t, result[0].Priority)
				assert.Equal(t, uint64(1), *result[0].Priority)

				assert.Equal(t, "10.255.0.3", result[1].RouterID)
				assert.Equal(t, "2-Way", result[1].State)
				assert.Equal(t, "wg-spoke1:10.0.13.2", result[1].Interface)
				assert.NotNil(t, result[1].Priority)
				assert.Equal(t, uint64(0), *result[1].Priority)
			},
		},
		{
			name:      "no ospf neighbors message",
			input:     "  No OSPF neighbors found  ",
			wantCount: 0,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.NotNil(t, result)
				assert.Empty(t, result)
			},
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Empty(t, result)
			},
		},
		{
			name:      "header only",
			input:     "Neighbor ID     Pri   State           Up Time         Dead Time       Address         Interface\n",
			wantCount: 0,
			check: func(t *testing.T, result []ospfNeighborRaw) {
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOSPFNeighborsText([]byte(tt.input))
			if tt.check != nil {
				tt.check(t, result)
			} else {
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestDiskUsagePercentFromBytes(t *testing.T) {
	u64 := func(v uint64) *uint64 { return &v }

	tests := []struct {
		name  string
		total *uint64
		used  *uint64
		want  *float64
	}{
		{
			name:  "normal 50 percent",
			total: u64(1000),
			used:  u64(500),
			want:  float64Ptr(50.0),
		},
		{
			name:  "nil total",
			total: nil,
			used:  u64(500),
			want:  nil,
		},
		{
			name:  "zero total",
			total: u64(0),
			used:  u64(500),
			want:  nil,
		},
		{
			name:  "nil used",
			total: u64(1000),
			used:  nil,
			want:  nil,
		},
		{
			name:  "both nil",
			total: nil,
			used:  nil,
			want:  nil,
		},
		{
			name:  "full disk",
			total: u64(1000),
			used:  u64(1000),
			want:  float64Ptr(100.0),
		},
		{
			name:  "empty disk",
			total: u64(1000),
			used:  u64(0),
			want:  float64Ptr(0.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diskUsagePercentFromBytes(tt.total, tt.used)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.InDelta(t, *tt.want, *got, 0.01)
			}
		})
	}
}

func float64Ptr(v float64) *float64 { return &v }

func TestParseMeminfoValueKB(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint64
	}{
		{
			name:  "typical meminfo line",
			input: "MemTotal:       16384 kB",
			want:  16384,
		},
		{
			name:  "single field",
			input: "MemTotal:",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "non-numeric value",
			input: "MemTotal:       abc kB",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMeminfoValueKB(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetStringAny(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		keys []string
		want string
	}{
		{
			name: "key exists",
			m:    map[string]any{"name": "test", "id": "123"},
			keys: []string{"name"},
			want: "test",
		},
		{
			name: "missing key",
			m:    map[string]any{"name": "test"},
			keys: []string{"missing"},
			want: "",
		},
		{
			name: "multiple fallback keys first found",
			m:    map[string]any{"ifaceName": "wg0", "ifName": "eth0"},
			keys: []string{"ifaceName", "ifName"},
			want: "wg0",
		},
		{
			name: "multiple fallback keys second found",
			m:    map[string]any{"ifName": "eth0"},
			keys: []string{"ifaceName", "ifName"},
			want: "eth0",
		},
		{
			name: "value is not a string",
			m:    map[string]any{"count": 42},
			keys: []string{"count"},
			want: "",
		},
		{
			name: "empty map",
			m:    map[string]any{},
			keys: []string{"anything"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringAny(tt.m, tt.keys...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetUintAny(t *testing.T) {
	u64 := func(v uint64) *uint64 { return &v }

	tests := []struct {
		name string
		m    map[string]any
		keys []string
		want *uint64
	}{
		{
			name: "float64 value",
			m:    map[string]any{"priority": float64(5)},
			keys: []string{"priority"},
			want: u64(5),
		},
		{
			name: "int value",
			m:    map[string]any{"priority": int(3)},
			keys: []string{"priority"},
			want: u64(3),
		},
		{
			name: "string value",
			m:    map[string]any{"priority": "7"},
			keys: []string{"priority"},
			want: u64(7),
		},
		{
			name: "missing key",
			m:    map[string]any{"other": float64(1)},
			keys: []string{"priority"},
			want: nil,
		},
		{
			name: "negative float64",
			m:    map[string]any{"priority": float64(-1)},
			keys: []string{"priority"},
			want: nil,
		},
		{
			name: "string with s suffix",
			m:    map[string]any{"timeout": "10s"},
			keys: []string{"timeout"},
			want: u64(10),
		},
		{
			name: "uint64 value",
			m:    map[string]any{"val": uint64(42)},
			keys: []string{"val"},
			want: u64(42),
		},
		{
			name: "fallback to second key",
			m:    map[string]any{"pri": float64(2)},
			keys: []string{"priority", "pri"},
			want: u64(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUintAny(tt.m, tt.keys...)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestNormalizeMillisToSeconds(t *testing.T) {
	u64 := func(v uint64) *uint64 { return &v }

	tests := []struct {
		name  string
		input *uint64
		want  *uint64
	}{
		{
			name:  "5000ms to 5s",
			input: u64(5000),
			want:  u64(5),
		},
		{
			name:  "999ms rounds to 1s",
			input: u64(999),
			want:  u64(1),
		},
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "0ms to 0s",
			input: u64(0),
			want:  u64(0),
		},
		{
			name:  "1500ms rounds to 2s",
			input: u64(1500),
			want:  u64(2),
		},
		{
			name:  "499ms rounds to 0s",
			input: u64(499),
			want:  u64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMillisToSeconds(tt.input)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}
