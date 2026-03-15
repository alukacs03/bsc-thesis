package actions

import (
	"fmt"
	"strings"

	"gluon-chaosmonkey/internal/ssh"
)

// IptablesBlock represents an iptables-based network fault action that drops
// all traffic on a WireGuard interface for a given experiment.
type IptablesBlock struct {
	// Interface is the full WireGuard interface name, e.g. "wg-worker1".
	Interface string
	// ExpID is the experiment identifier used in the iptables comment tag.
	ExpID string
	// TTL is the self-revert delay in seconds planted before applying the fault.
	TTL int
}

// comment returns the iptables comment tag for this experiment.
func (a *IptablesBlock) comment() string {
	return fmt.Sprintf("chaosctl-%s", a.ExpID)
}

// pidFile returns the path to the nohup PID file for this experiment.
func (a *IptablesBlock) pidFile() string {
	return fmt.Sprintf("/tmp/chaosctl-%s.pid", a.ExpID)
}

// Plant runs a nohup self-revert script on the target via SSH before the fault
// is applied. After TTL seconds the remote process will delete the iptables rules
// automatically. The nohup PID is saved to /tmp/chaosctl-<ExpID>.pid.
func (a *IptablesBlock) Plant(client *ssh.Client) error {
	cmd := fmt.Sprintf(
		"sleep %d && iptables -D INPUT -i %s -j DROP -m comment --comment %q && iptables -D OUTPUT -o %s -j DROP -m comment --comment %q",
		a.TTL,
		a.Interface,
		a.comment(),
		a.Interface,
		a.comment(),
	)
	fullCmd := fmt.Sprintf(
		"nohup bash -c %q >/dev/null 2>&1 & echo $! > %s",
		cmd,
		a.pidFile(),
	)
	if err := client.RunBackground(fullCmd); err != nil {
		return fmt.Errorf("plant self-revert for %s: %w", a.ExpID, err)
	}
	return nil
}

// Apply inserts INPUT and OUTPUT DROP rules for the WireGuard interface.
func (a *IptablesBlock) Apply(client *ssh.Client) error {
	cmds := []string{
		fmt.Sprintf("iptables -I INPUT -i %s -j DROP -m comment --comment %q", a.Interface, a.comment()),
		fmt.Sprintf("iptables -I OUTPUT -o %s -j DROP -m comment --comment %q", a.Interface, a.comment()),
	}
	for _, cmd := range cmds {
		_, _, err := client.Run(cmd)
		if err != nil {
			return fmt.Errorf("apply iptables rule (%s): %w", cmd, err)
		}
	}
	return nil
}

// Revert deletes the INPUT and OUTPUT DROP rules inserted by Apply.
// It also kills the nohup self-revert process if it is still running.
// "Rule not found" and "No such process" errors are treated as success
// to make the operation idempotent.
func (a *IptablesBlock) Revert(client *ssh.Client) error {
	// Kill the nohup self-revert process if it is still alive.
	killCmd := fmt.Sprintf(
		"pid=$(cat %s 2>/dev/null) && kill $pid 2>/dev/null || true",
		a.pidFile(),
	)
	if _, _, err := client.Run(killCmd); err != nil {
		// Non-fatal: the process may have already exited.
		_ = err
	}

	// Delete the iptables rules; treat "not found" as success.
	cmds := []string{
		fmt.Sprintf("iptables -D INPUT -i %s -j DROP -m comment --comment %q", a.Interface, a.comment()),
		fmt.Sprintf("iptables -D OUTPUT -o %s -j DROP -m comment --comment %q", a.Interface, a.comment()),
	}
	for _, cmd := range cmds {
		_, stderr, err := client.Run(cmd)
		if err != nil && !isNotFoundError(stderr) {
			return fmt.Errorf("revert iptables rule (%s): %w", cmd, err)
		}
	}
	return nil
}

// isNotFoundError returns true when stderr indicates the iptables rule does
// not exist (exit code 1 from iptables -D when no matching rule is found).
func isNotFoundError(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "no chain/target/match by that name") ||
		strings.Contains(lower, "bad rule") ||
		strings.Contains(lower, "does a matching rule exist") ||
		strings.Contains(lower, "resource temporarily unavailable") ||
		// iptables -D returns this when the rule is not present
		strings.Contains(lower, "no such process") ||
		strings.Contains(lower, "rule not found") ||
		strings.Contains(stderr, "iptables: No such file or directory") ||
		// iptables exits 1 with an empty-ish stderr on missing rule in some versions
		stderr == ""
}
