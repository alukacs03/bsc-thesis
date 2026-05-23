package actions

import (
	"fmt"
	"strings"

	"gluon-chaosmonkey/internal/ssh"
)

type IptablesBlock struct {
	Interfaces []string
	ExpID      string
	TTL        int
}

func (a *IptablesBlock) comment() string {
	return fmt.Sprintf("chaosctl-%s", a.ExpID)
}

func (a *IptablesBlock) pidFile() string {
	return fmt.Sprintf("/tmp/chaosctl-%s.pid", a.ExpID)
}

func (a *IptablesBlock) revertCmds() []string {
	var cmds []string
	for _, iface := range a.Interfaces {
		cmds = append(cmds,
			fmt.Sprintf("sudo iptables -D INPUT -i %s -j DROP -m comment --comment %q", iface, a.comment()),
			fmt.Sprintf("sudo iptables -D OUTPUT -o %s -j DROP -m comment --comment %q", iface, a.comment()),
		)
	}
	return cmds
}

func (a *IptablesBlock) Plant(client *ssh.Client) error {
	revertChain := strings.Join(a.revertCmds(), " ; ")
	cmd := fmt.Sprintf("sleep %d ; %s", a.TTL, revertChain)
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
	var cmds []string
	for _, iface := range a.Interfaces {
		cmds = append(cmds,
			fmt.Sprintf("sudo iptables -I INPUT -i %s -j DROP -m comment --comment %q", iface, a.comment()),
			fmt.Sprintf("sudo iptables -I OUTPUT -o %s -j DROP -m comment --comment %q", iface, a.comment()),
		)
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
		_ = err
	}

	for _, cmd := range a.revertCmds() {
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
