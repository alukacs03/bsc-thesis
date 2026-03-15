package actions

import (
	"fmt"
	"strings"

	"gluon-chaosmonkey/internal/ssh"
)

type IptablesBlock struct {
	Interface string
	ExpID     string
	TTL       int
}

func (a *IptablesBlock) comment() string {
	return fmt.Sprintf("chaosctl-%s", a.ExpID)
}

func (a *IptablesBlock) pidFile() string {
	return fmt.Sprintf("/tmp/chaosctl-%s.pid", a.ExpID)
}

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

func (a *IptablesBlock) Revert(client *ssh.Client) error {
	killCmd := fmt.Sprintf(
		"pid=$(cat %s 2>/dev/null) && kill $pid 2>/dev/null || true",
		a.pidFile(),
	)
	if _, _, err := client.Run(killCmd); err != nil {
		// Non-fatal: the process may have already exited.
		_ = err
	}

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

func isNotFoundError(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "no chain/target/match by that name") ||
		strings.Contains(lower, "bad rule") ||
		strings.Contains(lower, "does a matching rule exist") ||
		strings.Contains(lower, "resource temporarily unavailable") ||
		strings.Contains(lower, "no such process") ||
		strings.Contains(lower, "rule not found") ||
		strings.Contains(stderr, "iptables: No such file or directory") ||
		stderr == ""
}
