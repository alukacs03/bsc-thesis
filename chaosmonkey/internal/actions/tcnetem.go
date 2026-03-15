package actions

import (
	"fmt"
	"strings"

	"gluon-chaosmonkey/internal/ssh"
)

type TcNetem struct {
	Interface string
	LatencyMs int
	LossPct   float64
	ExpID     string
	TTL       int
}

func (a *TcNetem) pidFile() string {
	return fmt.Sprintf("/tmp/chaosctl-%s.pid", a.ExpID)
}

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

func (a *TcNetem) Revert(client *ssh.Client) error {
	killCmd := fmt.Sprintf(
		"pid=$(cat %s 2>/dev/null) && kill $pid 2>/dev/null || true",
		a.pidFile(),
	)
	if _, _, err := client.Run(killCmd); err != nil {
		// Non-fatal: the process may have already exited.
		_ = err
	}

	delCmd := fmt.Sprintf("tc qdisc del dev %s root", a.Interface)
	_, stderr, err := client.Run(delCmd)
	if err != nil && !isNoQdiscError(stderr) {
		return fmt.Errorf("revert tc-netem on %s: %w", a.Interface, err)
	}
	return nil
}

func isNoQdiscError(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "no such file or directory") ||
		strings.Contains(lower, "cannot delete qdisc with handle of zero") ||
		strings.Contains(lower, "no qdisc") ||
		stderr == ""
}
