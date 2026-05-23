package applier

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const overlayForwardComment = "gluon-overlay-forward"

// A KUBE-FORWARD chain INVALID-nak jelölné az ECMP miatt aszimmetrikus útvonalon
// érkező overlay válaszcsomagokat, ezért egy explicit ACCEPT szabályt szúrunk be elé.
func EnsureOverlayForwardRule(overlayCIDR string) error {
	if overlayCIDR == "" {
		overlayCIDR = "10.255.0.0/22"
	}

	checkCmd := exec.Command("iptables", "-C", "FORWARD",
		"-s", overlayCIDR, "-d", overlayCIDR,
		"-j", "ACCEPT",
		"-m", "comment", "--comment", overlayForwardComment)

	if err := checkCmd.Run(); err == nil {
		return nil
	}

	insertCmd := exec.Command("iptables", "-I", "FORWARD", "1",
		"-s", overlayCIDR, "-d", overlayCIDR,
		"-j", "ACCEPT",
		"-m", "comment", "--comment", overlayForwardComment)

	output, err := insertCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to insert overlay forward rule: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	log.Printf("Inserted overlay forward iptables rule for %s", overlayCIDR)
	return nil
}

func RemoveOverlayForwardRule(overlayCIDR string) error {
	if overlayCIDR == "" {
		overlayCIDR = "10.255.0.0/22"
	}

	deleteCmd := exec.Command("iptables", "-D", "FORWARD",
		"-s", overlayCIDR, "-d", overlayCIDR,
		"-j", "ACCEPT",
		"-m", "comment", "--comment", overlayForwardComment)

	if output, err := deleteCmd.CombinedOutput(); err != nil {
		out := strings.ToLower(string(output))
		if strings.Contains(out, "no such file") || strings.Contains(out, "bad rule") || strings.Contains(out, "does a matching rule exist") {
			return nil
		}
		return fmt.Errorf("failed to remove overlay forward rule: %w", err)
	}

	return nil
}
