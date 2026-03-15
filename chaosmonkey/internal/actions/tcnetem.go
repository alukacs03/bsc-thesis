package actions

import (
	"fmt"
	"strings"

	"gluon-chaosmonkey/internal/ssh"
)

// TcNetem represents a tc-netem fault action that adds network latency
// (and optionally packet loss) on a WireGuard interface for a given experiment.
type TcNetem struct {
	// Interface is the full WireGuard interface name, e.g. "wg-worker1".
	Interface string
	// LatencyMs is the emulated one-way delay in milliseconds.
	LatencyMs int
	// LossPct is the emulated packet loss percentage; 0 means no loss.
	LossPct float64
	// ExpID is the experiment identifier used for the PID file name.
	ExpID string
	// TTL is the self-revert delay in seconds planted before applying the fault.
	TTL int
}

// pidFile returns the path to the nohup PID file for this experiment.
func (a *TcNetem) pidFile() string {
	return fmt.Sprintf("/tmp/chaosctl-%s.pid", a.ExpID)
}

// Plant runs a nohup self-revert script on the target via SSH before the fault
// is applied. After TTL seconds the remote process deletes the netem qdisc
// automatically. The nohup PID is saved to /tmp/chaosctl-<ExpID>.pid.
func (a *TcNetem) Plant(client *ssh.Client) error {
	revertCmd := fmt.Sprintf("sleep %d && tc qdisc del dev %s root", a.TTL, a.Interface)
	fullCmd := fmt.Sprintf(
		"nohup bash -c %q >/dev/null 2>&1 & echo $! > %s",
		revertCmd,
		a.pidFile(),
	)
	if err := client.RunBackground(fullCmd); err != nil {
		return fmt.Errorf("plant self-revert for %s: %w", a.ExpID, err)
	}
	return nil
}

// Apply adds a netem qdisc on the WireGuard interface with the configured
// latency and optional packet loss.
func (a *TcNetem) Apply(client *ssh.Client) error {
	cmd := fmt.Sprintf("tc qdisc add dev %s root netem delay %dms", a.Interface, a.LatencyMs)
	if a.LossPct > 0 {
		cmd += fmt.Sprintf(" loss %g%%", a.LossPct)
	}
	_, _, err := client.Run(cmd)
	if err != nil {
		return fmt.Errorf("apply tc-netem on %s: %w", a.Interface, err)
	}
	return nil
}

// Revert kills the nohup self-revert process if still running and removes the
// netem qdisc. "No qdisc" / exit-code-2 errors are treated as success to make
// the operation idempotent.
func (a *TcNetem) Revert(client *ssh.Client) error {
	// Kill the nohup self-revert process if it is still alive.
	killCmd := fmt.Sprintf(
		"pid=$(cat %s 2>/dev/null) && kill $pid 2>/dev/null || true",
		a.pidFile(),
	)
	if _, _, err := client.Run(killCmd); err != nil {
		// Non-fatal: the process may have already exited.
		_ = err
	}

	// Delete the netem qdisc; treat "no qdisc" as success.
	delCmd := fmt.Sprintf("tc qdisc del dev %s root", a.Interface)
	_, stderr, err := client.Run(delCmd)
	if err != nil && !isNoQdiscError(stderr) {
		return fmt.Errorf("revert tc-netem on %s: %w", a.Interface, err)
	}
	return nil
}

// isNoQdiscError returns true when stderr indicates there is no qdisc to
// delete — tc exits with code 2 in that case.
func isNoQdiscError(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "no such file or directory") ||
		strings.Contains(lower, "cannot delete qdisc with handle of zero") ||
		strings.Contains(lower, "no qdisc") ||
		stderr == ""
}
