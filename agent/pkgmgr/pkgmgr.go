package pkgmgr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	frrRepoURL            = "https://deb.frrouting.org/frr"
	frrRepoListPath       = "/etc/apt/sources.list.d/frr.list"
	frrKeyringPath        = "/usr/share/keyrings/frrouting.gpg"
	frrKeyURL             = "https://deb.frrouting.org/frr.gpg"
	defaultDebianCodename = "bookworm"
	frrDaemonsPath        = "/etc/frr/daemons"

	k8sAptListPath    = "/etc/apt/sources.list.d/kubernetes.list"
	k8sKeyringPath    = "/etc/apt/keyrings/kubernetes-apt-keyring.gpg"
	k8sKeyURL         = "https://packages.cloud.google.com/apt/doc/apt-key.gpg"
	k8sRepoURL        = "https://apt.kubernetes.io"
	k8sModulesPath    = "/etc/modules-load.d/k8s.conf"
	k8sSysctlPath     = "/etc/sysctl.d/99-kubernetes-cri.conf"
	containerdCfgPath = "/etc/containerd/config.toml"
	cniBinDir         = "/opt/cni/bin"
	cniNetDir         = "/etc/cni/net.d"
)

var aptUpdated bool

func EnsureDependencies(ctx context.Context) error {
	if os.Geteuid() != 0 {
		return errors.New("package installation requires root privileges")
	}

	if err := ensureWireGuard(ctx); err != nil {
		return fmt.Errorf("wireguard dependency: %w", err)
	}

	if err := ensureFRR(ctx); err != nil {
		return fmt.Errorf("frr dependency: %w", err)
	}

	if err := ensureKubernetes(ctx); err != nil {
		return fmt.Errorf("kubernetes dependency: %w", err)
	}

	return nil
}

func ensureWireGuard(ctx context.Context) error {
	if commandExists("wg") && commandExists("wg-quick") {
		return nil
	}

	log.Println("WireGuard not detected, installing wireguard + wireguard-tools...")
	if err := aptUpdate(ctx); err != nil {
		return err
	}

	return aptInstall(ctx, "wireguard", "wireguard-tools")
}

func ensureFRR(ctx context.Context) error {
	if commandExists("vtysh") || pkgInstalled(ctx, "frr") {
		return nil
	}

	log.Println("FRR not detected, installing from FRRouting repository...")

	if err := ensureFRRRepo(ctx); err != nil {
		return err
	}

	if err := aptUpdate(ctx); err != nil {
		return err
	}

	if err := aptInstall(ctx, "frr", "frr-pythontools"); err != nil {
		return err
	}

	if err := configureFRRDaemons(); err != nil {
		return err
	}
	if _, err := runCommand(ctx, "systemctl", "enable", "--now", "frr"); err != nil {
		return err
	}
	if _, err := runCommand(ctx, "systemctl", "restart", "frr"); err != nil {
		return err
	}

	return nil
}

func ensureKubernetes(ctx context.Context) error {
	if commandExists("kubeadm") && commandExists("kubelet") && commandExists("kubectl") {
		log.Println("Kubernetes tools detected (kubeadm/kubelet/kubectl)")

		if err := ensureKernelPrereqs(ctx); err != nil {
			return err
		}
		if err := ensureContainerd(ctx); err != nil {
			return err
		}
		_, _ = runCommand(ctx, "systemctl", "enable", "--now", "kubelet")
		return nil
	}

	log.Println("Kubernetes tools not detected, installing container runtime + kubeadm/kubelet/kubectl...")

	if err := aptUpdate(ctx); err != nil {
		return err
	}

	_ = aptInstall(ctx, "ca-certificates", "curl", "wget", "gnupg")

	if err := ensureKernelPrereqs(ctx); err != nil {
		return err
	}

	if err := ensureContainerd(ctx); err != nil {
		return err
	}

	if err := ensureK8sRepo(ctx); err != nil {
		return err
	}

	if err := aptUpdate(ctx); err != nil {
		return err
	}

	if err := aptInstall(ctx, "kubelet", "kubeadm", "kubectl"); err != nil {
		return err
	}

	_, _ = runCommand(ctx, "apt-mark", "hold", "kubelet", "kubeadm", "kubectl")
	_, _ = runCommand(ctx, "systemctl", "enable", "--now", "kubelet")

	return nil
}

func ensureKernelPrereqs(ctx context.Context) error {
	if err := os.WriteFile(k8sModulesPath, []byte("overlay\nbr_netfilter\n"), 0644); err != nil {
		return fmt.Errorf("write %s: %w", k8sModulesPath, err)
	}
	_, _ = runCommand(ctx, "modprobe", "overlay")
	_, _ = runCommand(ctx, "modprobe", "br_netfilter")
	_, _ = runCommand(ctx, "modprobe", "vxlan")

	sysctlContent := `net.bridge.bridge-nf-call-iptables = 1
	net.bridge.bridge-nf-call-ip6tables = 1
	net.ipv4.ip_forward = 1
	net.ipv4.conf.all.rp_filter = 0
	net.ipv4.conf.default.rp_filter = 0
	`
	if err := os.WriteFile(k8sSysctlPath, []byte(sysctlContent), 0644); err != nil {
		return fmt.Errorf("write %s: %w", k8sSysctlPath, err)
	}
	_, _ = runCommand(ctx, "sysctl", "--system")

	if err := os.MkdirAll(cniBinDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", cniBinDir, err)
	}
	if err := os.MkdirAll(cniNetDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", cniNetDir, err)
	}

	_, _ = runCommand(ctx, "swapoff", "-a")
	if err := disableSwapInFstab(); err != nil {
		return err
	}

	return nil
}

func ensureContainerd(ctx context.Context) error {
	if commandExists("containerd") {
		log.Println("containerd detected")
		_, _ = runCommand(ctx, "systemctl", "enable", "--now", "containerd")
		return configureContainerdSystemdCgroup(ctx)
	}

	if err := aptUpdate(ctx); err != nil {
		return err
	}
	if err := aptInstall(ctx, "containerd"); err != nil {
		return err
	}
	_, _ = runCommand(ctx, "systemctl", "enable", "--now", "containerd")
	return configureContainerdSystemdCgroup(ctx)
}

func configureContainerdSystemdCgroup(ctx context.Context) error {
	if fileExists(containerdCfgPath) {
		data, err := os.ReadFile(containerdCfgPath)
		if err == nil && bytes.Contains(data, []byte("SystemdCgroup = true")) {
			return nil
		}
	}

	if err := os.MkdirAll("/etc/containerd", 0755); err != nil {
		return err
	}

	out, err := runCommand(ctx, "containerd", "config", "default")
	if err != nil {
		return fmt.Errorf("generate containerd config: %w", err)
	}
	cfg := string(out)
	cfg = strings.Replace(cfg, "SystemdCgroup = false", "SystemdCgroup = true", 1)
	if err := os.WriteFile(containerdCfgPath, []byte(cfg), 0644); err != nil {
		return fmt.Errorf("write %s: %w", containerdCfgPath, err)
	}

	_, _ = runCommand(ctx, "systemctl", "restart", "containerd")
	return nil
}

func ensureK8sRepo(ctx context.Context) error {
	if err := os.MkdirAll("/etc/apt/keyrings", 0755); err != nil {
		return err
	}

	keyBytes := commandOutput(ctx, "curl", "-fsSL", k8sKeyURL)
	if len(keyBytes) == 0 {
		keyBytes = commandOutput(ctx, "wget", "-qO-", k8sKeyURL)
	}
	if len(keyBytes) == 0 {
		return fmt.Errorf("failed to download Kubernetes key from %s", k8sKeyURL)
	}

	cmd := exec.CommandContext(ctx, "gpg", "--dearmor")
	cmd.Stdin = bytes.NewReader(keyBytes)
	keyring, err := cmd.Output()
	if err != nil || len(keyring) == 0 {

		keyring = keyBytes
	}

	if err := os.WriteFile(k8sKeyringPath, keyring, 0644); err != nil {
		return fmt.Errorf("write keyring: %w", err)
	}

	repoLine := fmt.Sprintf("deb [signed-by=%s] %s /\n", k8sKeyringPath, k8sRepoURL)
	if err := os.WriteFile(k8sAptListPath, []byte(repoLine), 0644); err != nil {
		return fmt.Errorf("write repo list: %w", err)
	}

	return nil
}

func disableSwapInFstab() error {
	const fstabPath = "/etc/fstab"
	data, err := os.ReadFile(fstabPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	changed := false

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		fields := strings.Fields(trim)
		if len(fields) >= 3 && fields[2] == "swap" {
			lines[i] = "# " + line
			changed = true
		}
	}

	if !changed {
		return nil
	}
	return os.WriteFile(fstabPath, []byte(strings.Join(lines, "\n")), 0644)
}

func ensureFRRRepo(ctx context.Context) error {
	repoExists := fileExists(frrRepoListPath)

	codename := strings.TrimSpace(string(commandOutput(ctx, "lsb_release", "-s", "-c")))
	if codename == "" {
		codename = defaultDebianCodename
	}

	keyBytes := commandOutput(ctx, "curl", "-fsSL", frrKeyURL)
	if len(keyBytes) == 0 {
		keyBytes = commandOutput(ctx, "wget", "-qO-", frrKeyURL)
	}
	if len(keyBytes) == 0 {
		return fmt.Errorf("failed to download FRR key from %s", frrKeyURL)
	}

	if err := os.WriteFile(frrKeyringPath, keyBytes, 0644); err != nil {
		return fmt.Errorf("write keyring: %w", err)
	}

	if !repoExists {
		repoLine := fmt.Sprintf("deb [signed-by=%s] %s %s frr-stable\n", frrKeyringPath, frrRepoURL, codename)
		if err := os.WriteFile(frrRepoListPath, []byte(repoLine), 0644); err != nil {
			return fmt.Errorf("write repo list: %w", err)
		}
	}

	log.Printf("Added FRR APT repo for %s", codename)
	return nil
}

func pkgInstalled(ctx context.Context, pkg string) bool {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Status}", pkg)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "install ok installed")
}

func aptUpdate(ctx context.Context) error {
	if aptUpdated {
		return nil
	}

	if _, err := runCommand(ctx, "apt-get", "update"); err != nil {
		return err
	}
	aptUpdated = true
	return nil
}

func aptInstall(ctx context.Context, pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	_, err := runCommand(ctx, "apt-get", args...)
	return err
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func commandOutput(ctx context.Context, name string, args ...string) []byte {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return nil
	}
	return out
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func configureFRRDaemons() error {
	desired := map[string]string{
		"ospfd": "yes",
	}

	current := map[string]string{}
	var lines []string
	if fileExists(frrDaemonsPath) {
		data, err := os.ReadFile(frrDaemonsPath)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "#") || !strings.Contains(trim, "=") {
				lines = append(lines, line)
				continue
			}
			parts := strings.SplitN(trim, "=", 2)
			key := parts[0]
			val := parts[1]
			current[key] = val
			if desiredVal, ok := desired[key]; ok && val != desiredVal {
				line = fmt.Sprintf("%s=%s", key, desiredVal)
			}
			lines = append(lines, line)
		}
	} else {
		lines = []string{
			"# Autogenerated by gluon-agent",
			"bgpd=no",
			"ospfd=yes",
			"ospf6d=no",
			"ripd=no",
			"ripngd=no",
			"isisd=no",
			"pimd=no",
			"pim6d=no",
			"ldpd=no",
			"nhrpd=no",
			"eigrpd=no",
			"babeld=no",
			"sharpd=no",
			"pbrd=no",
			"bfdd=no",
			"fabricd=no",
			"vrrpd=no",
			"pathd=no",
		}
	}

	for k, v := range desired {
		if cur, ok := current[k]; !ok || cur != v {
			lines = append(lines, fmt.Sprintf("%s=%s", k, v))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(frrDaemonsPath, []byte(content), 0644); err != nil {
		return err
	}
	return nil
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		return buf.Bytes(), fmt.Errorf("%s %s failed: %w - %s", name, strings.Join(args, " "), err, buf.String())
	}

	return buf.Bytes(), nil
}
