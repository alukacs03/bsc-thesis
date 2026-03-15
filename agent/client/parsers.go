package client

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type ospfNeighborRaw struct {
	RouterID  string
	State     string
	Interface string
	Priority  *uint64
}

func parseHumanDuration(s string) (time.Duration, bool) {
	parts := strings.Split(s, ",")
	var total time.Duration
	found := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) < 2 {
			continue
		}
		n, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil || n < 0 {
			continue
		}
		unit := strings.ToLower(fields[1])
		switch {
		case strings.HasPrefix(unit, "second"):
			total += time.Duration(n) * time.Second
			found = true
		case strings.HasPrefix(unit, "minute"):
			total += time.Duration(n) * time.Minute
			found = true
		case strings.HasPrefix(unit, "hour"):
			total += time.Duration(n) * time.Hour
			found = true
		case strings.HasPrefix(unit, "day"):
			total += time.Duration(n) * 24 * time.Hour
			found = true
		}
	}
	return total, found
}

func parseSizeBytes(s string) (uint64, bool) {
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return 0, false
	}
	val, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || val < 0 {
		return 0, false
	}
	unit := strings.ToLower(fields[1])
	mult := float64(1)
	switch unit {
	case "b":
		mult = 1
	case "kb":
		mult = 1e3
	case "mb":
		mult = 1e6
	case "gb":
		mult = 1e9
	case "tb":
		mult = 1e12
	case "kib":
		mult = 1024
	case "mib":
		mult = 1024 * 1024
	case "gib":
		mult = 1024 * 1024 * 1024
	case "tib":
		mult = 1024 * 1024 * 1024 * 1024
	default:
		return 0, false
	}
	return uint64(val * mult), true
}

func parseOSPFNeighborsJSON(b []byte) []ospfNeighborRaw {
	var root any
	if err := json.Unmarshal(b, &root); err != nil {
		return nil
	}
	m, ok := root.(map[string]any)
	if !ok {
		return nil
	}

	var neighbors any
	if v, ok := m["neighbors"]; ok {
		neighbors = v
	} else {
		neighbors = m
	}

	out := make([]ospfNeighborRaw, 0)

	switch t := neighbors.(type) {
	case map[string]any:
		for routerID, v := range t {
			switch vv := v.(type) {
			case []any:
				for _, item := range vv {
					if row, ok := item.(map[string]any); ok {
						iface := getStringAny(row, "ifaceName", "ifName", "ifname", "interfaceName", "interface")
						if iface == "" {
							continue
						}
						state := normalizeOSPFState(getStringAny(row, "state", "nbrState", "neighborState", "nbr_state"))
						priority := getUintAny(row, "priority", "pri", "routerPriority", "nbrPriority")
						out = append(out, ospfNeighborRaw{RouterID: routerID, State: state, Interface: iface, Priority: priority})
					}
				}
			case map[string]any:
				iface := getStringAny(vv, "ifaceName", "ifName", "ifname", "interfaceName", "interface")
				if iface == "" {
					continue
				}
				state := normalizeOSPFState(getStringAny(vv, "state", "nbrState", "neighborState", "nbr_state"))
				priority := getUintAny(vv, "priority", "pri", "routerPriority", "nbrPriority")
				out = append(out, ospfNeighborRaw{RouterID: routerID, State: state, Interface: iface, Priority: priority})
			}
		}
	case []any:
		for _, item := range t {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			routerID := getStringAny(row, "neighborId", "routerId", "router_id", "id")
			if routerID == "" {
				continue
			}
			iface := getStringAny(row, "ifaceName", "ifName", "ifname", "interfaceName", "interface")
			if iface == "" {
				continue
			}
			state := normalizeOSPFState(getStringAny(row, "state", "nbrState", "neighborState", "nbr_state"))
			priority := getUintAny(row, "priority", "pri", "routerPriority", "nbrPriority")
			out = append(out, ospfNeighborRaw{RouterID: routerID, State: state, Interface: iface, Priority: priority})
		}
	default:
		return nil
	}

	return out
}

func parseOSPFNeighborsText(b []byte) []ospfNeighborRaw {
	lines := strings.Split(string(b), "\n")
	out := make([]ospfNeighborRaw, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Neighbor ID") || strings.HasPrefix(line, "NeighborID") {
			continue
		}
		if strings.Contains(strings.ToLower(line), "no ospf neighbors") {
			return []ospfNeighborRaw{}
		}
		fields := strings.Fields(line)

		if len(fields) < 6 {
			continue
		}
		routerID := fields[0]
		pri, err := strconv.ParseUint(fields[1], 10, 64)
		var priPtr *uint64
		if err == nil {
			priPtr = &pri
		}
		state := fields[2]
		ifaceField := fields[len(fields)-1]
		out = append(out, ospfNeighborRaw{RouterID: routerID, State: normalizeOSPFState(state), Interface: ifaceField, Priority: priPtr})
	}
	return out
}

func normalizeOSPFState(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if base, _, ok := strings.Cut(s, "/"); ok && base != "" {
		s = base
	}

	l := strings.ToLower(s)
	switch l {
	case "full":
		return "Full"
	case "down":
		return "Down"
	case "init":
		return "Init"
	case "2-way", "2way", "two-way":
		return "2-Way"
	case "exstart":
		return "ExStart"
	case "exchange":
		return "Exchange"
	case "loading":
		return "Loading"
	}

	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func diskUsagePercentFromBytes(totalBytes *uint64, usedBytes *uint64) *float64 {
	if totalBytes == nil || usedBytes == nil || *totalBytes == 0 {
		return nil
	}
	usage := (float64(*usedBytes) / float64(*totalBytes)) * 100
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}
	return &usage
}

func parseMeminfoValueKB(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	n, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func getStringAny(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func getUintAny(m map[string]any, keys ...string) *uint64 {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}
		switch t := v.(type) {
		case float64:
			if t < 0 {
				continue
			}
			u := uint64(t)
			return &u
		case int:
			if t < 0 {
				continue
			}
			u := uint64(t)
			return &u
		case uint64:
			u := t
			return &u
		case string:
			trimmed := strings.TrimSpace(t)
			trimmed = strings.TrimSuffix(trimmed, "s")
			n, err := strconv.ParseUint(trimmed, 10, 64)
			if err != nil {
				continue
			}
			return &n
		}
	}
	return nil
}

func normalizeMillisToSeconds(v *uint64) *uint64 {
	if v == nil {
		return nil
	}
	val := (*v + 500) / 1000
	return &val
}
