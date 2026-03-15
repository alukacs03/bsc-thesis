package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"gluon-chaosmonkey/internal/actions"
	"gluon-chaosmonkey/internal/config"
	"gluon-chaosmonkey/internal/measure"
	"gluon-chaosmonkey/internal/results"
	"gluon-chaosmonkey/internal/ssh"
)

var configFlag string

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "chaosctl",
	Short: "chaosctl — inject and measure network faults on lab nodes",
	Long: `chaosctl injects network faults (iptables DROP, tc-netem) on remote nodes
over SSH, measures availability with ICMP probes, and saves structured
experiment results to disk.`,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a fault injection experiment",
}

var (
	iptablesTarget    string
	iptablesInterface string
	iptablesTTL       int
	iptablesPing      string
)

var (
	tcNetemTarget    string
	tcNetemInterface string
	tcNetemLatency   int
	tcNetemLoss      float64
	tcNetemTTL       int
	tcNetemPing      string
)

var tcNetemCmd = &cobra.Command{
	Use:   "tc-netem",
	Short: "Apply tc-netem latency/loss fault on a node interface",
	Long: `tc-netem connects to the target node via SSH, plants a self-revert
script, adds a netem qdisc on the specified interface for TTL seconds,
optionally measures ICMP availability, then reverts the qdisc and saves
a structured experiment result.`,
	RunE: runTcNetem,
}

var iptablesBlockCmd = &cobra.Command{
	Use:   "iptables-block",
	Short: "Apply iptables DROP fault on a node interface",
	Long: `iptables-block connects to the target node via SSH, plants a self-revert
script, drops all traffic on the specified interface for TTL seconds,
optionally measures ICMP availability, then reverts the rule and saves
a structured experiment result.`,
	RunE: runIptablesBlock,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "path to chaosmonkey.yaml (default: auto-detect)")

	iptablesBlockCmd.Flags().StringVar(&iptablesTarget, "target", "", "node name as defined in config (required)")
	iptablesBlockCmd.Flags().StringVar(&iptablesInterface, "interface", "", "WireGuard interface to block, e.g. wg-worker1 (required)")
	iptablesBlockCmd.Flags().IntVar(&iptablesTTL, "ttl", 60, "fault duration in seconds (self-revert after TTL)")
	iptablesBlockCmd.Flags().StringVar(&iptablesPing, "ping", "", "IP to ICMP-probe during the fault (default: node loopback)")

	_ = iptablesBlockCmd.MarkFlagRequired("target")
	_ = iptablesBlockCmd.MarkFlagRequired("interface")

	tcNetemCmd.Flags().StringVar(&tcNetemTarget, "target", "", "node name as defined in config (required)")
	tcNetemCmd.Flags().StringVar(&tcNetemInterface, "interface", "", "WireGuard interface to apply netem on, e.g. wg-worker1 (required)")
	tcNetemCmd.Flags().IntVar(&tcNetemLatency, "latency", 0, "emulated one-way delay in milliseconds (required, must be > 0)")
	tcNetemCmd.Flags().Float64Var(&tcNetemLoss, "loss", 0, "emulated packet loss percentage (optional, default 0)")
	tcNetemCmd.Flags().IntVar(&tcNetemTTL, "ttl", 60, "fault duration in seconds (self-revert after TTL)")
	tcNetemCmd.Flags().StringVar(&tcNetemPing, "ping", "", "IP to ICMP-probe during the fault (default: node loopback)")

	_ = tcNetemCmd.MarkFlagRequired("target")
	_ = tcNetemCmd.MarkFlagRequired("interface")
	_ = tcNetemCmd.MarkFlagRequired("latency")

	runCmd.AddCommand(iptablesBlockCmd)
	runCmd.AddCommand(tcNetemCmd)
	rootCmd.AddCommand(runCmd)
}

func runIptablesBlock(cmd *cobra.Command, args []string) error {
	startedAt := time.Now()
	expID := results.ExpID(startedAt)

	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	node, ok := cfg.Nodes[iptablesTarget]
	if !ok {
		return fmt.Errorf("node %q not found in config (known nodes: %v)", iptablesTarget, nodeNames(cfg))
	}

	pingTarget := iptablesPing
	if pingTarget == "" {
		pingTarget = node.Loopback
	}

	fmt.Printf("Experiment: %s\n", expID)
	fmt.Printf("Target:     %s (%s)\n", iptablesTarget, node.Host)
	fmt.Printf("Interface:  %s\n", iptablesInterface)
	fmt.Printf("TTL:        %ds\n", iptablesTTL)
	if pingTarget != "" {
		fmt.Printf("Ping:       %s\n", pingTarget)
	}
	fmt.Println()

	client, err := ssh.NewClient(node, cfg.SSHUser, cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("ssh connect to %s: %w", iptablesTarget, err)
	}
	defer client.Close()

	var measurer *measure.Measurer
	if pingTarget != "" {
		measurer = measure.NewMeasurer(pingTarget)
		measurer.Start()
	}

	action := actions.IptablesBlock{
		Interface: iptablesInterface,
		ExpID:     expID,
		TTL:       iptablesTTL,
	}

	if err := action.Plant(client); err != nil {
		if measurer != nil {
			measurer.Stop()
		}
		return fmt.Errorf("plant self-revert: %w", err)
	}

	faultAppliedAt := time.Now()
	fmt.Printf("Fault applied at %s\n", faultAppliedAt.Format(time.RFC3339))
	if err := action.Apply(client); err != nil {
		_ = action.Revert(client)
		if measurer != nil {
			measurer.Stop()
		}
		res := results.ExperimentResult{
			ID:             expID,
			Type:           "iptables-block",
			Target:         iptablesTarget,
			Interface:      iptablesInterface,
			TTLSeconds:     iptablesTTL,
			StartedAt:      startedAt,
			FaultAppliedAt: &faultAppliedAt,
			PingTarget:     pingTarget,
			Status:         "failed",
		}
		if saveErr := results.Save(res, cfg.ResultsDir); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save result: %v\n", saveErr)
		}
		return fmt.Errorf("apply fault: %w", err)
	}

	fmt.Printf("Waiting %ds for auto-revert...\n", iptablesTTL)
	time.Sleep(time.Duration(iptablesTTL) * time.Second)

	faultRevertedAt := time.Now()
	revertErr := action.Revert(client)

	var measureResult measure.MeasureResult
	if measurer != nil {
		measureResult = measurer.Stop()
	}

	status := "completed"
	if revertErr != nil {
		status = "revert_failed"
		fmt.Fprintf(os.Stderr, "warning: revert failed: %v\n", revertErr)
	}

	var downtimeMs int64
	if measureResult.FirstLossAt != nil && measureResult.FirstRecoveryAt != nil {
		downtimeMs = measureResult.FirstRecoveryAt.Sub(*measureResult.FirstLossAt).Milliseconds()
	}

	res := results.ExperimentResult{
		ID:              expID,
		Type:            "iptables-block",
		Target:          iptablesTarget,
		Interface:       iptablesInterface,
		TTLSeconds:      iptablesTTL,
		StartedAt:       startedAt,
		FaultAppliedAt:  &faultAppliedAt,
		FaultRevertedAt: &faultRevertedAt,
		PingTarget:      pingTarget,
		FirstLossAt:     measureResult.FirstLossAt,
		FirstRecoveryAt: measureResult.FirstRecoveryAt,
		DowntimeMs:      downtimeMs,
		PingSamples:     measureResult.TotalSamples,
		PacketsLost:     measureResult.PacketsLost,
		Status:          status,
	}

	resultPath := filepath.Join(cfg.ResultsDir, expID+".json")
	if err := results.Save(res, cfg.ResultsDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save result: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("Experiment:   %s\n", expID)
	fmt.Printf("Target:       %s (%s)\n", iptablesTarget, iptablesInterface)
	if measurer != nil {
		fmt.Printf("Downtime:     %dms\n", downtimeMs)
		fmt.Printf("Packets lost: %d/%d\n", measureResult.PacketsLost, measureResult.TotalSamples)
	}
	fmt.Printf("Status:       %s\n", status)
	fmt.Printf("Result:       %s\n", resultPath)

	return nil
}

func runTcNetem(cmd *cobra.Command, args []string) error {
	if tcNetemLatency <= 0 {
		return fmt.Errorf("--latency must be > 0 (got %d)", tcNetemLatency)
	}

	startedAt := time.Now()
	expID := results.ExpID(startedAt)

	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	node, ok := cfg.Nodes[tcNetemTarget]
	if !ok {
		return fmt.Errorf("node %q not found in config (known nodes: %v)", tcNetemTarget, nodeNames(cfg))
	}

	pingTarget := tcNetemPing
	if pingTarget == "" {
		pingTarget = node.Loopback
	}

	fmt.Printf("Experiment: %s\n", expID)
	fmt.Printf("Target:     %s (%s)\n", tcNetemTarget, node.Host)
	fmt.Printf("Interface:  %s\n", tcNetemInterface)
	fmt.Printf("Latency:    %dms\n", tcNetemLatency)
	if tcNetemLoss > 0 {
		fmt.Printf("Loss:       %g%%\n", tcNetemLoss)
	}
	fmt.Printf("TTL:        %ds\n", tcNetemTTL)
	if pingTarget != "" {
		fmt.Printf("Ping:       %s\n", pingTarget)
	}
	fmt.Println()

	client, err := ssh.NewClient(node, cfg.SSHUser, cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("ssh connect to %s: %w", tcNetemTarget, err)
	}
	defer client.Close()

	var measurer *measure.Measurer
	if pingTarget != "" {
		measurer = measure.NewMeasurer(pingTarget)
		measurer.Start()
	}

	action := actions.TcNetem{
		Interface: tcNetemInterface,
		LatencyMs: tcNetemLatency,
		LossPct:   tcNetemLoss,
		ExpID:     expID,
		TTL:       tcNetemTTL,
	}

	if err := action.Plant(client); err != nil {
		if measurer != nil {
			measurer.Stop()
		}
		return fmt.Errorf("plant self-revert: %w", err)
	}

	faultAppliedAt := time.Now()
	fmt.Printf("Fault applied at %s\n", faultAppliedAt.Format(time.RFC3339))
	if err := action.Apply(client); err != nil {
		_ = action.Revert(client)
		if measurer != nil {
			measurer.Stop()
		}
		res := results.ExperimentResult{
			ID:             expID,
			Type:           "tc-netem",
			Target:         tcNetemTarget,
			Interface:      tcNetemInterface,
			TTLSeconds:     tcNetemTTL,
			StartedAt:      startedAt,
			FaultAppliedAt: &faultAppliedAt,
			PingTarget:     pingTarget,
			Status:         "failed",
		}
		if saveErr := results.Save(res, cfg.ResultsDir); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save result: %v\n", saveErr)
		}
		return fmt.Errorf("apply fault: %w", err)
	}

	fmt.Printf("Waiting %ds for auto-revert...\n", tcNetemTTL)
	time.Sleep(time.Duration(tcNetemTTL) * time.Second)

	faultRevertedAt := time.Now()
	revertErr := action.Revert(client)

	var measureResult measure.MeasureResult
	if measurer != nil {
		measureResult = measurer.Stop()
	}

	status := "completed"
	if revertErr != nil {
		status = "revert_failed"
		fmt.Fprintf(os.Stderr, "warning: revert failed: %v\n", revertErr)
	}

	var downtimeMs int64
	if measureResult.FirstLossAt != nil && measureResult.FirstRecoveryAt != nil {
		downtimeMs = measureResult.FirstRecoveryAt.Sub(*measureResult.FirstLossAt).Milliseconds()
	}

	res := results.ExperimentResult{
		ID:              expID,
		Type:            "tc-netem",
		Target:          tcNetemTarget,
		Interface:       tcNetemInterface,
		TTLSeconds:      tcNetemTTL,
		StartedAt:       startedAt,
		FaultAppliedAt:  &faultAppliedAt,
		FaultRevertedAt: &faultRevertedAt,
		PingTarget:      pingTarget,
		FirstLossAt:     measureResult.FirstLossAt,
		FirstRecoveryAt: measureResult.FirstRecoveryAt,
		DowntimeMs:      downtimeMs,
		PingSamples:     measureResult.TotalSamples,
		PacketsLost:     measureResult.PacketsLost,
		Status:          status,
	}

	resultPath := filepath.Join(cfg.ResultsDir, expID+".json")
	if err := results.Save(res, cfg.ResultsDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save result: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("Experiment:   %s\n", expID)
	fmt.Printf("Target:       %s (%s)\n", tcNetemTarget, tcNetemInterface)
	if measurer != nil {
		fmt.Printf("Downtime:     %dms\n", downtimeMs)
		fmt.Printf("Packets lost: %d/%d\n", measureResult.PacketsLost, measureResult.TotalSamples)
	}
	fmt.Printf("Status:       %s\n", status)
	fmt.Printf("Result:       %s\n", resultPath)

	return nil
}

var cleanupTarget string

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Force-revert any active fault on a target node",
	RunE:  runCleanup,
}

func init() {
	cleanupCmd.Flags().StringVar(&cleanupTarget, "target", "", "node name as defined in config (required)")
	_ = cleanupCmd.MarkFlagRequired("target")
	rootCmd.AddCommand(cleanupCmd)
}

func runCleanup(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	node, ok := cfg.Nodes[cleanupTarget]
	if !ok {
		return fmt.Errorf("node %q not found in config", cleanupTarget)
	}
	client, err := ssh.NewClient(node, cfg.SSHUser, cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	// Kill any nohup revert PIDs
	stdout, _, _ := client.Run("ls /tmp/chaosctl-*.pid 2>/dev/null")
	pidFiles := strings.Fields(stdout)
	for _, f := range pidFiles {
		pid, _, _ := client.Run(fmt.Sprintf("cat %s 2>/dev/null", f))
		pid = strings.TrimSpace(pid)
		if pid != "" {
			client.Run(fmt.Sprintf("kill %s 2>/dev/null || true", pid))
			fmt.Printf("Killed nohup PID %s (%s)\n", pid, f)
		}
		client.Run(fmt.Sprintf("rm -f %s", f))
	}

	// Flush iptables DROP rules tagged with chaosctl-
	_, _, err = client.Run("iptables-save | grep 'chaosctl-' | sed 's/-A/-D/' | while read r; do iptables $r 2>/dev/null || true; done")
	if err == nil {
		fmt.Println("Flushed iptables chaosctl rules")
	}

	// Delete tc qdiscs on all node interfaces
	for _, iface := range node.Interfaces {
		_, stderr, _ := client.Run(fmt.Sprintf("tc qdisc del dev %s root 2>&1 || true", iface))
		if !strings.Contains(stderr, "No such file") && !strings.Contains(stderr, "Cannot delete") {
			fmt.Printf("Removed tc qdisc on %s\n", iface)
		}
	}

	fmt.Printf("Cleanup complete on %s\n", cleanupTarget)
	return nil
}

var resultsLast int

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "List recent experiment results",
	RunE:  runResults,
}

func init() {
	resultsCmd.Flags().IntVar(&resultsLast, "last", 10, "number of recent results to show")
	rootCmd.AddCommand(resultsCmd)
}

func runResults(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	list, err := results.List(cfg.ResultsDir, resultsLast)
	if err != nil {
		return fmt.Errorf("list results: %w", err)
	}
	if len(list) == 0 {
		fmt.Println("No results found")
		return nil
	}
	fmt.Printf("%-28s %-15s %-10s %-14s %-12s %-12s %s\n",
		"EXP_ID", "TYPE", "TARGET", "INTERFACE", "DOWNTIME_MS", "STATUS", "STARTED_AT")
	fmt.Println(strings.Repeat("-", 110))
	for _, r := range list {
		fmt.Printf("%-28s %-15s %-10s %-14s %-12d %-12s %s\n",
			r.ID, r.Type, r.Target, r.Interface, r.DowntimeMs, r.Status,
			r.StartedAt.UTC().Format("2006-01-02T15:04:05Z"))
	}
	return nil
}

func nodeNames(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Nodes))
	for k := range cfg.Nodes {
		names = append(names, k)
	}
	return names
}
