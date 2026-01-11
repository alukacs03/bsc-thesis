//go:build linux
// +build linux

package client

import (
	"context"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type heartbeatPayload struct {
	AgentVersion string   `json:"agent_version"`
	DesiredRole  string   `json:"desired_role,omitempty"`
	CPUUsage     *float64 `json:"cpu_usage"`
	MemoryUsage  *float64 `json:"memory_usage"`
	DiskUsage    *float64 `json:"disk_usage"`
	DiskTotalBytes *uint64 `json:"disk_total_bytes"`
	DiskUsedBytes  *uint64 `json:"disk_used_bytes"`
	Logs         []string `json:"logs"`
	UptimeSeconds *uint64 `json:"uptime_seconds"`
	WireGuardPeers []wireGuardPeerSnapshot `json:"wireguard_peers"`
	OSPFNeighbors  []ospfNeighborSnapshot  `json:"ospf_neighbors"`
	SystemUsers   []string `json:"system_users"`
	SystemServices []systemServiceSnapshot `json:"system_services"`
}

type systemServiceSnapshot struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ActiveState   string `json:"active_state"`
	SubState      string `json:"sub_state"`
	UnitFileState string `json:"unit_file_state"`
}

type wireGuardPeerSnapshot struct {
	Interface           string `json:"interface"`
	PeerPublicKey       string `json:"peer_public_key"`
	Endpoint            string `json:"endpoint"`
	AllowedIPs          string `json:"allowed_ips"`
	LatestHandshakeUnix int64  `json:"latest_handshake_unix"`
	RxBytes             uint64 `json:"rx_bytes"`
	TxBytes             uint64 `json:"tx_bytes"`
}

type ospfNeighborSnapshot struct {
	RouterID              string  `json:"router_id"`
	Area                  string  `json:"area"`
	State                 string  `json:"state"`
	Interface             string  `json:"interface"`
	HelloIntervalSeconds  *uint64 `json:"hello_interval_seconds"`
	DeadIntervalSeconds   *uint64 `json:"dead_interval_seconds"`
	Cost                  *uint64 `json:"cost"`
	Priority              *uint64 `json:"priority"`
}

func (c *Client) Heartbeat(apiKey string, desiredRole string) error {
	diskTotalBytes, diskUsedBytes := diskBytes()
	payload := heartbeatPayload{
		AgentVersion: AgentVersion,
		DesiredRole:  strings.TrimSpace(desiredRole),
		CPUUsage:     cpuUsagePercent(),
		MemoryUsage:  memUsagePercent(),
		DiskUsage:    diskUsagePercentFromBytes(diskTotalBytes, diskUsedBytes),
		DiskTotalBytes: diskTotalBytes,
		DiskUsedBytes:  diskUsedBytes,
		Logs:         readAgentLogsLastTwoMinutes(),
		UptimeSeconds: uptimeSeconds(),
		WireGuardPeers: readWireGuardPeers(),
		OSPFNeighbors:  readOSPFNeighbors(),
		SystemUsers:    readSystemUsers(),
		SystemServices: readSystemServices(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/agent/heartbeat", bytes.NewBuffer(body))
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

	
	var respPayload struct {
		Commands []struct {
			ID      uint            `json:"id"`
			Kind    string          `json:"kind"`
			Payload json.RawMessage `json:"payload"`
		} `json:"commands"`
	}
	if b, _ := io.ReadAll(resp.Body); len(b) > 0 {
		_ = json.Unmarshal(b, &respPayload)
	}
	if len(respPayload.Commands) > 0 {
		results := executeCommands(respPayload.Commands)
		if len(results) > 0 {
			_ = c.ReportCommandResults(apiKey, results)
		}
	}

	return nil
}

func readAgentLogsLastTwoMinutes() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	
	
	cmd := exec.CommandContext(ctx, "journalctl", "-xeu", "gluon-agent", "--since", "2 minutes ago", "--no-pager")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return []string{}
	}
	return strings.Split(text, "\n")
}

func readWireGuardPeers() []wireGuardPeerSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wg", "show", "all", "dump")
	out, err := cmd.Output()
	if err != nil {
		return readWireGuardPeersText()
	}

	splitDumpFields := func(line string) []string {
		line = strings.TrimRight(line, "\r\n")
		if strings.Contains(line, "\t") {
			parts := strings.Split(line, "\t")
			for len(parts) > 0 && parts[len(parts)-1] == "" {
				parts = parts[:len(parts)-1]
			}
			return parts
		}
		return strings.Fields(line)
	}

	peers := make([]wireGuardPeerSnapshot, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := splitDumpFields(line)
		
		if len(fields) == 5 {
			continue
		}
		
		
		if len(fields) < 9 {
			continue
		}

		iface := fields[0]
		peerPub := fields[1]
		endpoint := fields[3]
		handshakeUnixStr := fields[len(fields)-4]
		rxStr := fields[len(fields)-3]
		txStr := fields[len(fields)-2]

		allowedIPs := fields[4]
		if len(fields) > 9 {
			allowedIPs = strings.Join(fields[4:len(fields)-4], " ")
		}

		handshakeUnix, err := strconv.ParseInt(handshakeUnixStr, 10, 64)
		if err != nil {
			handshakeUnix = 0
		}
		rx, err := strconv.ParseUint(rxStr, 10, 64)
		if err != nil {
			rx = 0
		}
		tx, err := strconv.ParseUint(txStr, 10, 64)
		if err != nil {
			tx = 0
		}

		peers = append(peers, wireGuardPeerSnapshot{
			Interface:           iface,
			PeerPublicKey:       peerPub,
			Endpoint:            endpoint,
			AllowedIPs:          allowedIPs,
			LatestHandshakeUnix: handshakeUnix,
			RxBytes:             rx,
			TxBytes:             tx,
		})
	}

	if len(peers) == 0 {
		return readWireGuardPeersText()
	}
	return peers
}

func readWireGuardPeersText() []wireGuardPeerSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wg", "show")
	out, err := cmd.Output()
	if err != nil {
		return []wireGuardPeerSnapshot{}
	}

	now := time.Now()
	var currentInterface string
	var currentPeer *wireGuardPeerSnapshot
	peers := make([]wireGuardPeerSnapshot, 0)

	flushPeer := func() {
		if currentPeer == nil || currentPeer.PeerPublicKey == "" || currentInterface == "" {
			currentPeer = nil
			return
		}
		currentPeer.Interface = currentInterface
		peers = append(peers, *currentPeer)
		currentPeer = nil
	}

	lines := strings.Split(string(out), "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "interface:") {
			flushPeer()
			currentInterface = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
			continue
		}
		if strings.HasPrefix(line, "peer:") {
			flushPeer()
			currentPeer = &wireGuardPeerSnapshot{
				PeerPublicKey: strings.TrimSpace(strings.TrimPrefix(line, "peer:")),
			}
			continue
		}
		if currentPeer == nil {
			continue
		}

		if strings.HasPrefix(line, "endpoint:") {
			currentPeer.Endpoint = strings.TrimSpace(strings.TrimPrefix(line, "endpoint:"))
			continue
		}
		if strings.HasPrefix(line, "allowed ips:") {
			currentPeer.AllowedIPs = strings.TrimSpace(strings.TrimPrefix(line, "allowed ips:"))
			continue
		}
		if strings.HasPrefix(line, "latest handshake:") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "latest handshake:"))
			if strings.EqualFold(v, "never") {
				currentPeer.LatestHandshakeUnix = 0
			} else if strings.HasSuffix(v, " ago") {
				ageStr := strings.TrimSuffix(v, " ago")
				if d, ok := parseHumanDuration(ageStr); ok {
					currentPeer.LatestHandshakeUnix = now.Add(-d).Unix()
				}
			}
			continue
		}
		if strings.HasPrefix(line, "transfer:") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "transfer:"))
			parts := strings.Split(v, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if strings.HasSuffix(p, " received") {
					size := strings.TrimSpace(strings.TrimSuffix(p, " received"))
					if n, ok := parseSizeBytes(size); ok {
						currentPeer.RxBytes = n
					}
				} else if strings.HasSuffix(p, " sent") {
					size := strings.TrimSpace(strings.TrimSuffix(p, " sent"))
					if n, ok := parseSizeBytes(size); ok {
						currentPeer.TxBytes = n
					}
				}
			}
			continue
		}
	}

	flushPeer()
	return peers
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

type ospfInterfaceMeta struct {
	area                string
	helloIntervalSeconds *uint64
	deadIntervalSeconds  *uint64
	cost                *uint64
}

func readOSPFNeighbors() []ospfNeighborSnapshot {
	ifaceMeta := readOSPFInterfaceMeta()
	rawNeighbors := readOSPFNeighborsRaw()
	out := make([]ospfNeighborSnapshot, 0, len(rawNeighbors))
	for _, n := range rawNeighbors {
		meta, ok := ifaceMeta[n.Interface]
		if !ok {
			if base, _, ok2 := strings.Cut(n.Interface, ":"); ok2 && base != "" {
				meta = ifaceMeta[base]
			}
		}
		out = append(out, ospfNeighborSnapshot{
			RouterID:             n.RouterID,
			Area:                 meta.area,
			State:                n.State,
			Interface:            n.Interface,
			HelloIntervalSeconds: meta.helloIntervalSeconds,
			DeadIntervalSeconds:  meta.deadIntervalSeconds,
			Cost:                 meta.cost,
			Priority:             n.Priority,
		})
	}
	return out
}

type ospfNeighborRaw struct {
	RouterID  string
	State     string
	Interface string
	Priority  *uint64
}

func readOSPFNeighborsRaw() []ospfNeighborRaw {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vtysh", "-c", "show ip ospf neighbor json")
	out, err := cmd.Output()
	if err == nil {
		raw := parseOSPFNeighborsJSON(out)
		if raw != nil {
			return raw
		}
	}

	
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	cmd2 := exec.CommandContext(ctx2, "vtysh", "-c", "show ip ospf neighbor")
	out2, err2 := cmd2.Output()
	if err2 != nil {
		return []ospfNeighborRaw{}
	}
	return parseOSPFNeighborsText(out2)
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
						priority := getUintAny(row, "priority", "pri", "routerPriority")
						out = append(out, ospfNeighborRaw{RouterID: routerID, State: state, Interface: iface, Priority: priority})
					}
				}
			case map[string]any:
				iface := getStringAny(vv, "ifaceName", "ifName", "ifname", "interfaceName", "interface")
				if iface == "" {
					continue
				}
				state := normalizeOSPFState(getStringAny(vv, "state", "nbrState", "neighborState", "nbr_state"))
				priority := getUintAny(vv, "priority", "pri", "routerPriority")
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
			priority := getUintAny(row, "priority", "pri", "routerPriority")
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

func readOSPFInterfaceMeta() map[string]ospfInterfaceMeta {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vtysh", "-c", "show ip ospf interface json")
	out, err := cmd.Output()
	if err != nil {
		return map[string]ospfInterfaceMeta{}
	}

	var root any
	if err := json.Unmarshal(out, &root); err != nil {
		return map[string]ospfInterfaceMeta{}
	}
	m, ok := root.(map[string]any)
	if !ok {
		return map[string]ospfInterfaceMeta{}
	}

	ifacesAny, ok := m["interfaces"]
	if !ok {
		
		ifacesAny = m["interface"]
	}
	outMeta := make(map[string]ospfInterfaceMeta)

	addMeta := func(name string, row map[string]any) {
		if name == "" {
			name = getStringAny(row, "ifName", "ifname", "ifaceName", "interfaceName", "interface", "name")
		}
		if name == "" {
			return
		}

		area := getStringAny(row, "areaId", "area", "area_id")
		hello := getUintAny(row, "helloInterval", "hello_interval", "helloIntervalSeconds")
		dead := getUintAny(row, "deadInterval", "dead_interval", "deadIntervalSeconds")
		cost := getUintAny(row, "cost")

		outMeta[name] = ospfInterfaceMeta{
			area:                 area,
			helloIntervalSeconds: hello,
			deadIntervalSeconds:  dead,
			cost:                 cost,
		}

		
		if base, _, ok := strings.Cut(name, ":"); ok && base != "" {
			if _, exists := outMeta[base]; !exists {
				outMeta[base] = outMeta[name]
			}
		}
	}

	switch t := ifacesAny.(type) {
	case map[string]any:
		for name, v := range t {
			row, ok := v.(map[string]any)
			if !ok {
				continue
			}
			addMeta(name, row)
		}
	case []any:
		for _, item := range t {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			addMeta("", row)
		}
	default:
		return map[string]ospfInterfaceMeta{}
	}

	return outMeta
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

func uptimeSeconds() *uint64 {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return nil
	}
	fields := strings.Fields(string(b))
	if len(fields) < 1 {
		return nil
	}
	secsFloat, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || secsFloat < 0 {
		return nil
	}
	secs := uint64(secsFloat)
	return &secs
}

func diskBytes() (totalBytes *uint64, usedBytes *uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return nil, nil
	}
	if stat.Blocks == 0 {
		return nil, nil
	}

	blockSize := uint64(stat.Bsize)
	total := uint64(stat.Blocks) * blockSize
	used := uint64(stat.Blocks-stat.Bfree) * blockSize
	return &total, &used
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

func cpuUsagePercent() *float64 {
	total, idle, ok := readProcStat()
	if !ok {
		return nil
	}

	cpuStatMu.Lock()
	defer cpuStatMu.Unlock()

	
	if lastCPUStat.ok {
		deltaTotal := total - lastCPUStat.total
		deltaIdle := idle - lastCPUStat.idle
		lastCPUStat.total = total
		lastCPUStat.idle = idle
		if deltaTotal == 0 {
			return nil
		}

		usage := (float64(deltaTotal-deltaIdle) / float64(deltaTotal)) * 100
		if usage < 0 {
			usage = 0
		} else if usage > 100 {
			usage = 100
		}
		return &usage
	}

	
	lastCPUStat.total = total
	lastCPUStat.idle = idle
	lastCPUStat.ok = true

	time.Sleep(250 * time.Millisecond)
	total2, idle2, ok2 := readProcStat()
	if !ok2 {
		return nil
	}
	lastCPUStat.total = total2
	lastCPUStat.idle = idle2

	deltaTotal := total2 - total
	deltaIdle := idle2 - idle
	if deltaTotal == 0 {
		return nil
	}

	usage := (float64(deltaTotal-deltaIdle) / float64(deltaTotal)) * 100
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}
	return &usage
}

func readProcStat() (total uint64, idle uint64, ok bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return 0, 0, false
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, false
	}

	var nums []uint64
	for _, f := range fields[1:] {
		n, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		nums = append(nums, n)
	}

	for _, n := range nums {
		total += n
	}

	
	idle = nums[3]
	if len(nums) >= 5 {
		idle += nums[4]
	}

	return total, idle, true
}

var cpuStatMu sync.Mutex
var lastCPUStat struct {
	total uint64
	idle  uint64
	ok    bool
}

func memUsagePercent() *float64 {
	totalKB, availableKB, ok := readMemInfo()
	if !ok || totalKB == 0 {
		return nil
	}

	usedKB := totalKB - availableKB
	usage := (float64(usedKB) / float64(totalKB)) * 100
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}
	return &usage
}

func readMemInfo() (totalKB uint64, availableKB uint64, ok bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			totalKB = parseMeminfoValueKB(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			availableKB = parseMeminfoValueKB(line)
		}
		if totalKB > 0 && availableKB > 0 {
			return totalKB, availableKB, true
		}
	}
	return 0, 0, false
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

func diskUsagePercent() *float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return nil
	}
	if stat.Blocks == 0 {
		return nil
	}

	used := stat.Blocks - stat.Bfree
	usage := (float64(used) / float64(stat.Blocks)) * 100
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}
	return &usage
}

func readSystemUsers() []string {
	b, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return []string{}
	}
	lines := strings.Split(string(b), "\n")
	users := make([]string, 0, 8)
	seen := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		username := parts[0]
		if username == "" {
			continue
		}
		uid, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}
		if username != "root" && uid < 1000 {
			continue
		}
		if seen[username] {
			continue
		}
		seen[username] = true
		users = append(users, username)
	}
	
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			if users[j] < users[i] {
				users[i], users[j] = users[j], users[i]
			}
		}
	}
	return users
}

func readSystemServices() []systemServiceSnapshot {
	servicesEnv := strings.TrimSpace(os.Getenv("GLUON_MONITORED_SERVICES"))
	names := []string{"gluon-agent.service", "frr.service"}
	if servicesEnv != "" {
		parts := strings.Split(servicesEnv, ",")
		names = make([]string, 0, len(parts))
		for _, p := range parts {
			v := strings.TrimSpace(p)
			if v != "" {
				names = append(names, v)
			}
		}
	}
	if len(names) == 0 {
		return []systemServiceSnapshot{}
	}

	readOne := func(name string) []systemctlShowBlock {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(
			ctx,
			"systemctl",
			"show",
			name,
			"-p", "Id",
			"-p", "Description",
			"-p", "ActiveState",
			"-p", "SubState",
			"-p", "UnitFileState",
			"--no-page",
		)
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		return splitSystemctlShowBlocks(string(out))
	}

	blocks := []systemctlShowBlock{}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		args := []string{"show"}
		args = append(args, names...)
		args = append(args,
			"-p", "Id",
			"-p", "Description",
			"-p", "ActiveState",
			"-p", "SubState",
			"-p", "UnitFileState",
			"--no-page",
		)
		cmd := exec.CommandContext(ctx, "systemctl", args...)
		if out, err := cmd.Output(); err == nil {
			blocks = splitSystemctlShowBlocks(string(out))
		}
	}

	if len(blocks) == 0 {
		for _, name := range names {
			blocks = append(blocks, readOne(name)...)
		}
	}

	outList := make([]systemServiceSnapshot, 0, len(blocks))
	for _, b := range blocks {
		if b.Name == "" {
			continue
		}
		outList = append(outList, systemServiceSnapshot{
			Name:          b.Name,
			Description:   b.Description,
			ActiveState:   b.ActiveState,
			SubState:      b.SubState,
			UnitFileState: b.UnitFileState,
		})
	}

	
	for i := 0; i < len(outList); i++ {
		for j := i + 1; j < len(outList); j++ {
			if outList[j].Name < outList[i].Name {
				outList[i], outList[j] = outList[j], outList[i]
			}
		}
	}
	return outList
}

type systemctlShowBlock struct {
	Name          string
	Description   string
	ActiveState   string
	SubState      string
	UnitFileState string
}

func splitSystemctlShowBlocks(out string) []systemctlShowBlock {
	lines := strings.Split(out, "\n")
	blocks := make([]systemctlShowBlock, 0)
	var cur systemctlShowBlock
	flush := func() {
		if cur.Name != "" || cur.Description != "" {
			blocks = append(blocks, cur)
		}
		cur = systemctlShowBlock{}
	}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "Id":
			cur.Name = v
		case "Description":
			cur.Description = v
		case "ActiveState":
			cur.ActiveState = v
		case "SubState":
			cur.SubState = v
		case "UnitFileState":
			cur.UnitFileState = v
		}
	}
	flush()
	return blocks
}
