//go:build linux
// +build linux

package kubernetes

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gluon-agent/client"
	"gluon-agent/pkgmgr"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	adminConfPath            = "/etc/kubernetes/admin.conf"
	kubeletConfPath          = "/etc/kubernetes/kubelet.conf"
	controlPlaneManifestPath = "/etc/kubernetes/manifests/kube-apiserver.yaml"
	etcdManifestPath         = "/etc/kubernetes/manifests/etcd.yaml"
	cniNetDir                = "/etc/cni/net.d"
	kubeletKubeadmFlagsPath  = "/var/lib/kubelet/kubeadm-flags.env"
	kubeletSystemdDropInDir  = "/etc/systemd/system/kubelet.service.d"
	kubeletNodeIPDropInPath  = "/etc/systemd/system/kubelet.service.d/20-gluon-node-ip.conf"
)

var flannelPatchMu sync.Mutex
var lastFlannelTolerationPatch time.Time
var lastFlannelInstallAttempt time.Time

var controlPlaneLabelMu sync.Mutex
var lastControlPlaneLabelAttempt time.Time

var controlPlaneRejoinMu sync.Mutex
var lastControlPlaneRejoinAttempt time.Time

var controlPlaneHealthRejoinMu sync.Mutex
var lastControlPlaneHealthRejoinAttempt time.Time

var kubeletNodeIPMu sync.Mutex
var lastKubeletNodeIPAttempt time.Time

var missingJoinMu sync.Mutex
var lastMissingJoinReport time.Time

var apiserverManifestMu sync.Mutex
var lastApiserverManifestPatch time.Time

type initResult struct {
	workerJoinCommand       string
	controlPlaneJoinCommand string
	joinExpiresAt           time.Time
}

func Sync(ctx context.Context, apiClient *client.Client, apiKey string) {
	task, err := apiClient.GetKubernetesTask(apiKey)
	if err != nil {
		log.Printf("Failed to get kubernetes task: %v", err)
		return
	}

	
	maybeEnsureFlannelTolerations(ctx)
	
	maybeKickKubeletForCNI(ctx)
	
	maybeEnsureKubeletNodeIP(ctx)
	
	
	maybeEnsureControlPlaneLabels(ctx)
	
	maybeEnsureApiserverAnonymousAuth(ctx)

	
	
	task = maybeForceControlPlaneWireGuardRejoin(ctx, apiClient, apiKey, task)
	
	task = maybeForceControlPlaneRejoinWhenLocalAPIServerUnhealthy(ctx, apiClient, apiKey, task)

	action := strings.ToLower(strings.TrimSpace(task.Action))
	if action != "" && action != "none" {
		if task.Note != "" {
			log.Printf("Kubernetes task: action=%s note=%s", action, task.Note)
		} else {
			log.Printf("Kubernetes task: action=%s", action)
		}
	}

	retriedNone := false
DECIDE:
	switch action {
	case "", "none":
		
		
		if !retriedNone && !isJoined() {
			retriedNone = true
			if maybeForceRejoinWhenNotJoined(apiClient, apiKey) {
				if refreshed, err := apiClient.GetKubernetesTask(apiKey); err == nil && refreshed != nil {
					task = refreshed
					action = strings.ToLower(strings.TrimSpace(task.Action))
					goto DECIDE
				}
			}
		}
		return
	case "wait":
		return
	case "init":
		log.Println("Kubernetes: ensuring dependencies...")
		if err := pkgmgr.EnsureDependencies(ctx); err != nil {
			log.Printf("Kubernetes init failed (dependencies): %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: fmt.Sprintf("dependency install failed: %v", err)})
			return
		}

		log.Printf("Kubernetes init parameters: endpoint=%q podCIDR=%q serviceCIDR=%q version=%q",
			task.ControlPlaneEndpoint, task.PodCIDR, task.ServiceCIDR, task.KubernetesVersion)

		res, err := initCluster(ctx, task)
		if err != nil {
			log.Printf("Kubernetes init failed: %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
			return
		}

		log.Println("Kubernetes init succeeded; reporting join commands to API")
		maybeEnsureKubeletNodeIP(ctx)

		report := client.KubernetesReport{
			State:                   "cluster_initialized",
			ControlPlaneEndpoint:    task.ControlPlaneEndpoint,
			PodCIDR:                 task.PodCIDR,
			ServiceCIDR:             task.ServiceCIDR,
			KubernetesVersion:       task.KubernetesVersion,
			WorkerJoinCommand:       res.workerJoinCommand,
			ControlPlaneJoinCommand: res.controlPlaneJoinCommand,
			JoinCommandExpiresAt:    res.joinExpiresAt.UTC().Format(time.RFC3339),
		}
		if err := apiClient.ReportKubernetes(apiKey, report); err != nil {
			log.Printf("Failed to report kubernetes init: %v", err)
		}
		return
	case "join_control_plane":
		if isJoined() {
			if isControlPlaneNode() {
				didReset := false
				
				if mismatch, current, desired, err := controlPlaneAdvertiseAddressMismatch(ctx); err == nil && mismatch {
					if isSecondaryControlPlane() && shouldAutoResetOnRoleMismatch() {
						log.Printf("Kubernetes: control-plane advertise address mismatch (%s != %s); resetting to rejoin over WireGuard", current, desired)
						if err := resetKubeadmState(ctx); err != nil {
							log.Printf("Kubernetes reset failed: %v", err)
							_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
							return
						}
						didReset = true
					} else if isSecondaryControlPlane() {
						msg := fmt.Sprintf("control-plane advertise address mismatch (%s != %s); auto-reset disabled", current, desired)
						log.Printf("Kubernetes: %s", msg)
						_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: msg})
						return
					}
				}
				
				if !didReset {
					_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "joined_control_plane"})
					return
				}
			}

			msg := "node already joined as worker; cannot become control-plane without reset"
			if shouldAutoResetOnRoleMismatch() {
				log.Println("Kubernetes: hub joined as worker; auto-reset enabled, resetting kubeadm state...")
				if err := resetKubeadmState(ctx); err != nil {
					log.Printf("Kubernetes reset failed: %v", err)
					_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
					return
				}
			} else {
				log.Printf("Kubernetes: %s (set GLUON_K8S_AUTO_RESET_ON_ROLE_MISMATCH=true to auto-fix)", msg)
				_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{
					State:   "error",
					Message: msg + "; run: kubeadm reset -f && rm -rf /etc/kubernetes /var/lib/etcd /var/lib/kubelet/pki",
				})
				return
			}
		} else if isControlPlaneNode() {
			
			
			
			log.Println("Kubernetes: corrupted state detected (control-plane manifests exist but kubelet.conf missing); resetting to rejoin cleanly")
			if err := resetKubeadmState(ctx); err != nil {
				log.Printf("Kubernetes reset failed: %v", err)
				_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
				return
			}
		}
		log.Println("Kubernetes: ensuring dependencies...")
		if err := pkgmgr.EnsureDependencies(ctx); err != nil {
			log.Printf("Kubernetes join(control-plane) failed (dependencies): %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: fmt.Sprintf("dependency install failed: %v", err)})
			return
		}
		log.Printf("Kubernetes join(control-plane): target=%s", parseJoinTarget(task.JoinCommand))
		if err := joinCluster(ctx, task.JoinCommand); err != nil {
			log.Printf("Kubernetes join(control-plane) failed: %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
			return
		}
		ensureControlPlaneLabels(ctx)
		maybeEnsureKubeletNodeIP(ctx)
		_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "joined_control_plane"})
		return
	case "join_worker":
		if isJoined() {
			
			
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "joined_worker"})
			return
		} else if hasPartialKubeletState() {
			
			
			log.Println("Kubernetes: corrupted state detected (partial kubelet state without kubelet.conf); resetting to rejoin cleanly")
			if err := resetKubeadmState(ctx); err != nil {
				log.Printf("Kubernetes reset failed: %v", err)
				_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
				return
			}
		}
		log.Println("Kubernetes: ensuring dependencies...")
		if err := pkgmgr.EnsureDependencies(ctx); err != nil {
			log.Printf("Kubernetes join(worker) failed (dependencies): %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: fmt.Sprintf("dependency install failed: %v", err)})
			return
		}
		log.Printf("Kubernetes join(worker): target=%s", parseJoinTarget(task.JoinCommand))
		if err := joinCluster(ctx, task.JoinCommand); err != nil {
			log.Printf("Kubernetes join(worker) failed: %v", err)
			_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
			return
		}
		maybeEnsureKubeletNodeIP(ctx)
		_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "joined_worker"})
		return
	default:
		log.Printf("Unknown kubernetes task action: %q", task.Action)
		return
	}
}

func maybeForceRejoinWhenNotJoined(apiClient *client.Client, apiKey string) bool {
	missingJoinMu.Lock()
	defer missingJoinMu.Unlock()
	if time.Since(lastMissingJoinReport) < 5*time.Minute {
		return false
	}
	lastMissingJoinReport = time.Now()

	msg := "local kubelet.conf missing; forcing rejoin task"
	if err := apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: msg}); err != nil {
		log.Printf("Kubernetes: failed to report missing join state: %v", err)
		return false
	}
	log.Printf("Kubernetes: %s", msg)
	return true
}

func maybeForceControlPlaneRejoinWhenLocalAPIServerUnhealthy(ctx context.Context, apiClient *client.Client, apiKey string, task *client.KubernetesTask) *client.KubernetesTask {
	action := strings.ToLower(strings.TrimSpace(task.Action))
	if action != "" && action != "none" && action != "wait" {
		return task
	}

	
	if !isJoined() || !isControlPlaneNode() {
		return task
	}
	
	if !isSecondaryControlPlane() {
		return task
	}

	status, err := localAPIServerLivezStatus(ctx)
	if err == nil && status == http.StatusOK {
		return task
	}

	
	if err == nil && status == http.StatusForbidden {
		if patched, perr := patchApiserverManifestFlag("--anonymous-auth=true"); perr == nil && patched {
			log.Printf("Kubernetes: patched kube-apiserver manifest to allow anonymous /livez (status=%d); restarting kubelet", status)
			_, _ = runLogged(ctx, "systemctl", "restart", "kubelet")
			return task
		}
	}

	controlPlaneHealthRejoinMu.Lock()
	defer controlPlaneHealthRejoinMu.Unlock()
	if time.Since(lastControlPlaneHealthRejoinAttempt) < 30*time.Minute {
		return task
	}
	lastControlPlaneHealthRejoinAttempt = time.Now()

	msg := fmt.Sprintf("control-plane unhealthy (local /livez status=%s); forcing rejoin", formatHTTPStatus(status, err))
	log.Printf("Kubernetes: %s", msg)

	
	if err := apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: msg}); err != nil {
		log.Printf("Kubernetes: failed to report control-plane health error to API: %v", err)
		return task
	}
	newTask, err := apiClient.GetKubernetesTask(apiKey)
	if err != nil {
		log.Printf("Kubernetes: failed to refetch task after health report: %v", err)
		return task
	}
	if strings.ToLower(strings.TrimSpace(newTask.Action)) != "join_control_plane" || strings.TrimSpace(newTask.JoinCommand) == "" {
		log.Printf("Kubernetes: expected join_control_plane task after health report, got %q", strings.TrimSpace(newTask.Action))
		return task
	}

	
	if err := resetKubeadmState(ctx); err != nil {
		log.Printf("Kubernetes reset failed: %v", err)
		_ = apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: err.Error()})
		return task
	}
	return newTask
}

func initCluster(ctx context.Context, task *client.KubernetesTask) (*initResult, error) {
	if isInitialized() {
		log.Println("Kubernetes already initialized; refreshing join commands...")
		if err := ensureRootKubeconfig(); err != nil {
			log.Printf("Warning: failed to set up kubeconfig: %v", err)
		}
		ensureFlannelTolerations(ctx)
		return generateJoinCommands(ctx)
	}

	if err := detectBrokenKubeadmState(); err != nil {
		log.Printf("Kubernetes preflight failed: %v", err)
		return nil, err
	}

	ensureKubeletNodeIPPreJoin(ctx)

	endpoint := strings.TrimSpace(task.ControlPlaneEndpoint)
	if endpoint == "" {
		log.Printf("Kubernetes init missing control_plane_endpoint from task: %+v", *task)
		return nil, fmt.Errorf("missing control_plane_endpoint")
	}
	if !strings.Contains(endpoint, ":") {
		endpoint = endpoint + ":6443"
	}
	advertiseAddr := strings.Split(endpoint, ":")[0]

	podCIDR := strings.TrimSpace(task.PodCIDR)
	if podCIDR == "" {
		podCIDR = "10.244.0.0/16"
	}
	serviceCIDR := strings.TrimSpace(task.ServiceCIDR)
	if serviceCIDR == "" {
		serviceCIDR = "10.96.0.0/16"
	}

	log.Printf("Initializing Kubernetes cluster (endpoint=%s podCIDR=%s serviceCIDR=%s)...", endpoint, podCIDR, serviceCIDR)

	if out, err := runLogged(ctx, "kubeadm", "init",
		"--apiserver-advertise-address", advertiseAddr,
		"--apiserver-cert-extra-sans", advertiseAddr,
		"--control-plane-endpoint", endpoint,
		"--pod-network-cidr", podCIDR,
		"--service-cidr", serviceCIDR,
		"--upload-certs",
		"--skip-token-print",
	); err != nil {
		return nil, fmt.Errorf("kubeadm init failed: %w\n%s", err, truncate(out, 8000))
	}

	if err := ensureRootKubeconfig(); err != nil {
		log.Printf("Warning: failed to set up kubeconfig: %v", err)
	}

	maybeEnsureApiserverAnonymousAuth(ctx)

	
	log.Println("Installing Flannel CNI...")
	if out, err := runKubectlLogged(ctx, "apply", "-f", "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"); err != nil {
		log.Printf("Warning: failed to apply flannel manifest: %v\n%s", err, truncate(out, 4000))
	}
	ensureFlannelTolerations(ctx)

	return generateJoinCommands(ctx)
}

func generateJoinCommands(ctx context.Context) (*initResult, error) {
	
	
	
	const joinTTL = 2 * time.Hour

	joinOut, err := output(ctx, "kubeadm", "token", "create", "--print-join-command", "--ttl", "2h0m0s")
	if err != nil {
		return nil, err
	}
	workerJoin := strings.TrimSpace(string(joinOut))
	if workerJoin == "" {
		return nil, fmt.Errorf("empty worker join command")
	}

	certOut, err := output(ctx, "kubeadm", "init", "phase", "upload-certs", "--upload-certs")
	if err != nil {
		return nil, err
	}
	certKey := extractCertificateKey(string(certOut))
	if certKey == "" {
		return nil, fmt.Errorf("failed to parse certificate key from kubeadm output")
	}

	controlPlaneJoin := strings.TrimSpace(workerJoin + " --control-plane --certificate-key " + certKey)

	return &initResult{
		workerJoinCommand:       workerJoin,
		controlPlaneJoinCommand: controlPlaneJoin,
		joinExpiresAt:           time.Now().Add(joinTTL),
	}, nil
}

func joinCluster(ctx context.Context, joinCommand string) error {
	if err := detectBrokenKubeadmState(); err != nil {
		log.Printf("Kubernetes preflight failed: %v", err)
		return err
	}

	joinCommand = strings.TrimSpace(joinCommand)
	if joinCommand == "" {
		return fmt.Errorf("missing join command")
	}

	ensureKubeletNodeIPPreJoin(ctx)

	log.Printf("Joining Kubernetes cluster...")

	parts := splitShellWords(joinCommand)
	if len(parts) < 2 || parts[0] != "kubeadm" || parts[1] != "join" {
		return fmt.Errorf("unexpected join command (expected kubeadm join ...)")
	}

	
	
	if containsArg(parts, "--control-plane") && !hasFlag(parts, "--apiserver-advertise-address") {
		if advertiseAddr, err := detectWireGuardAdvertiseAddress(ctx); err == nil && advertiseAddr != "" {
			log.Printf("Kubernetes: using WireGuard advertise address %s for control-plane join", advertiseAddr)
			parts = append(parts, "--apiserver-advertise-address", advertiseAddr)
		} else if err != nil {
			log.Printf("Warning: failed to detect WireGuard advertise address; proceeding without override: %v", err)
		}
	}

	out, err := runLogged(ctx, parts[0], parts[1:]...)
	if err != nil {
		return fmt.Errorf("kubeadm join failed: %w\n%s", err, truncate(out, 8000))
	}
	maybeEnsureApiserverAnonymousAuth(ctx)
	return nil
}

func ensureFlannelTolerations(ctx context.Context) {
	
	
	patch := `{"spec":{"template":{"spec":{"tolerations":[{"operator":"Exists"}]}}}}`

	dsList, err := listDaemonSets(ctx)
	if err != nil {
		log.Printf("Warning: failed to list daemonsets for flannel patch: %v", err)
		return
	}

	var targets [][2]string
	for _, ds := range dsList {
		ns, name := ds[0], ds[1]
		ln := strings.ToLower(name)
		if strings.Contains(ln, "flannel") {
			targets = append(targets, [2]string{ns, name})
		}
	}

	if len(targets) == 0 {
		
		if time.Since(lastFlannelInstallAttempt) > 5*time.Minute {
			lastFlannelInstallAttempt = time.Now()
			log.Printf("No flannel daemonset found; attempting to install Flannel CNI...")
			if out, err := runKubectlLogged(ctx, "apply", "-f", "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"); err != nil {
				log.Printf("Warning: failed to apply flannel manifest: %v\n%s", err, truncate(out, 4000))
				return
			}
			
			dsList, err = listDaemonSets(ctx)
			if err != nil {
				log.Printf("Warning: failed to re-list daemonsets after flannel install: %v", err)
				return
			}
			targets = targets[:0]
			for _, ds := range dsList {
				ns, name := ds[0], ds[1]
				ln := strings.ToLower(name)
				if strings.Contains(ln, "flannel") {
					targets = append(targets, [2]string{ns, name})
				}
			}
		}

		if len(targets) == 0 {
			log.Printf("Warning: no flannel daemonset found to patch tolerations")
			return
		}
	}

	for _, t := range targets {
		ns, name := t[0], t[1]
		out, err := runKubectlLogged(ctx, "-n", ns, "patch", "daemonset", name, "--type=merge", "-p", patch)
		if err != nil {
			log.Printf("Warning: failed to patch flannel tolerations for %s/%s: %v\n%s", ns, name, err, truncate(out, 4000))
			continue
		}
		log.Printf("Patched flannel tolerations (%s/%s)", ns, name)
	}
}

func maybeEnsureFlannelTolerations(ctx context.Context) {
	
	if !isInitialized() {
		return
	}

	flannelPatchMu.Lock()
	defer flannelPatchMu.Unlock()
	if time.Since(lastFlannelTolerationPatch) < 2*time.Minute {
		return
	}
	lastFlannelTolerationPatch = time.Now()

	if err := ensureRootKubeconfig(); err != nil {
		log.Printf("Warning: failed to set up kubeconfig: %v", err)
		return
	}
	ensureFlannelTolerations(ctx)
}

func ensureControlPlaneLabels(ctx context.Context) {
	
	
	if !fileExists(adminConfPath) {
		return
	}
	nodeName := resolveSelfNodeName(ctx)
	if nodeName == "" {
		return
	}

	
	_, _ = runKubectlLogged(ctx, "label", "node", nodeName, "node-role.kubernetes.io/control-plane=", "--overwrite")
	_, _ = runKubectlLogged(ctx, "label", "node", nodeName, "node-role.kubernetes.io/master=", "--overwrite")
	_, _ = runKubectlLogged(ctx, "taint", "node", nodeName, "node-role.kubernetes.io/control-plane=:NoSchedule", "--overwrite")
}

func isInitialized() bool {
	_, err := os.Stat(adminConfPath)
	return err == nil
}

func isJoined() bool {
	_, err := os.Stat(kubeletConfPath)
	return err == nil
}

func isControlPlaneNode() bool {
	_, err := os.Stat(controlPlaneManifestPath)
	return err == nil
}




func hasPartialKubeletState() bool {
	
	if isJoined() {
		return false
	}
	
	partialIndicators := []string{
		"/var/lib/kubelet/config.yaml",
		"/var/lib/kubelet/pki",
	}
	for _, p := range partialIndicators {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func ensureRootKubeconfig() error {
	if _, err := os.Stat(adminConfPath); err != nil {
		return err
	}

	dir := "/root/.kube"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	dst := filepath.Join(dir, "config")
	if err := copyFile(adminConfPath, dst, 0600); err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func extractCertificateKey(out string) string {
	re := regexp.MustCompile(`(?m)^[0-9a-f]{64}$`)
	m := re.FindString(out)
	return strings.TrimSpace(m)
}

func runLogged(ctx context.Context, name string, args ...string) (string, error) {
	return runLoggedWithEnv(ctx, nil, name, args...)
}

func runLoggedWithEnv(ctx context.Context, extraEnv []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Env, extraEnv...)
	}
	var buf bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &buf)
	cmd.Stdout = mw
	cmd.Stderr = mw
	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return buf.String(), nil
}

func output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s failed: %w - %s", name, strings.Join(args, " "), err, string(out))
	}
	return out, nil
}

func splitShellWords(s string) []string {
	
	
	return strings.Fields(s)
}

func runKubectlLogged(ctx context.Context, args ...string) (string, error) {
	if !fileExists(adminConfPath) {
		return "", fmt.Errorf("missing kubeconfig %s", adminConfPath)
	}
	kubeArgs := append([]string{"--kubeconfig", adminConfPath}, args...)
	env := []string{
		"KUBECONFIG=" + adminConfPath,
		"HOME=/root",
	}
	return runLoggedWithEnv(ctx, env, "kubectl", kubeArgs...)
}

func listDaemonSets(ctx context.Context) ([][2]string, error) {
	out, err := runKubectlCaptured(ctx, "get", "daemonset", "-A", "-o", "custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name", "--no-headers")
	if err != nil {
		return nil, err
	}

	var result [][2]string
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		result = append(result, [2]string{fields[0], fields[1]})
	}
	return result, nil
}

func runKubectlCaptured(ctx context.Context, args ...string) (string, error) {
	if !fileExists(adminConfPath) {
		return "", fmt.Errorf("missing kubeconfig %s", adminConfPath)
	}
	kubeArgs := append([]string{"--kubeconfig", adminConfPath}, args...)
	env := []string{
		"KUBECONFIG=" + adminConfPath,
		"HOME=/root",
	}
	return runCapturedWithEnv(ctx, env, "kubectl", kubeArgs...)
}

func runCapturedWithEnv(ctx context.Context, extraEnv []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Env, extraEnv...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func parseJoinTarget(joinCommand string) string {
	parts := strings.Fields(strings.TrimSpace(joinCommand))
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "join" && i+1 < len(parts) {
			
			return parts[i+1]
		}
	}
	return "unknown"
}

func maybeForceControlPlaneWireGuardRejoin(ctx context.Context, apiClient *client.Client, apiKey string, task *client.KubernetesTask) *client.KubernetesTask {
	action := strings.ToLower(strings.TrimSpace(task.Action))
	if action != "" && action != "none" && action != "wait" {
		return task
	}

	
	if !isJoined() || !isControlPlaneNode() {
		return task
	}

	
	if !isSecondaryControlPlane() {
		return task
	}

	mismatch, current, desired, err := controlPlaneAdvertiseAddressMismatch(ctx)
	if err != nil || !mismatch {
		return task
	}

	controlPlaneRejoinMu.Lock()
	defer controlPlaneRejoinMu.Unlock()
	if time.Since(lastControlPlaneRejoinAttempt) < 10*time.Minute {
		return task
	}
	lastControlPlaneRejoinAttempt = time.Now()

	msg := fmt.Sprintf("control-plane advertise address mismatch (%s != %s); forcing rejoin over WireGuard", current, desired)
	log.Printf("Kubernetes: %s", msg)

	
	if err := apiClient.ReportKubernetes(apiKey, client.KubernetesReport{State: "error", Message: msg}); err != nil {
		log.Printf("Kubernetes: failed to report mismatch to API: %v", err)
		return task
	}

	newTask, err := apiClient.GetKubernetesTask(apiKey)
	if err != nil {
		log.Printf("Kubernetes: failed to refetch task after mismatch report: %v", err)
		return task
	}

	na := strings.ToLower(strings.TrimSpace(newTask.Action))
	if na != "join_control_plane" || strings.TrimSpace(newTask.JoinCommand) == "" {
		log.Printf("Kubernetes: expected join_control_plane task after mismatch report, got %q", na)
		return task
	}
	return newTask
}

func controlPlaneAdvertiseAddressMismatch(ctx context.Context) (bool, string, string, error) {
	desired, err := detectWireGuardAdvertiseAddress(ctx)
	if err != nil || desired == "" {
		return false, "", "", err
	}

	current, err := currentControlPlaneAdvertiseAddress()
	if err != nil || current == "" {
		return false, "", "", err
	}

	
	overlayPool, _ := netip.ParsePrefix("10.255.0.0/16")
	if addr, err := netip.ParseAddr(current); err == nil && !overlayPool.Contains(addr) {
		return true, current, desired, nil
	}
	return strings.TrimSpace(current) != strings.TrimSpace(desired), current, desired, nil
}

func currentControlPlaneAdvertiseAddress() (string, error) {
	b, err := os.ReadFile(controlPlaneManifestPath)
	if err != nil {
		return "", err
	}
	s := string(b)
	
	re := regexp.MustCompile(`--advertise-address=([0-9]{1,3}(?:\.[0-9]{1,3}){3})`)
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1]), nil
	}
	
	re2 := regexp.MustCompile(`--apiserver-advertise-address(?:=|\s+)([0-9]{1,3}(?:\.[0-9]{1,3}){3})`)
	m = re2.FindStringSubmatch(s)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1]), nil
	}
	return "", fmt.Errorf("could not find advertise address in %s", controlPlaneManifestPath)
}

func maybeEnsureApiserverAnonymousAuth(ctx context.Context) {
	if !isControlPlaneNode() {
		return
	}
	apiserverManifestMu.Lock()
	defer apiserverManifestMu.Unlock()
	if time.Since(lastApiserverManifestPatch) < 10*time.Minute {
		return
	}
	lastApiserverManifestPatch = time.Now()

	changed, err := patchApiserverManifestFlag("--anonymous-auth=true")
	if err != nil {
		log.Printf("Kubernetes: failed to patch kube-apiserver manifest: %v", err)
		return
	}
	if changed {
		log.Println("Kubernetes: enabled anonymous kube-apiserver access for /livez probes; restarting kubelet")
		_, _ = runLogged(ctx, "systemctl", "restart", "kubelet")
	}
}

func patchApiserverManifestFlag(flag string) (bool, error) {
	b, err := os.ReadFile(controlPlaneManifestPath)
	if err != nil {
		return false, err
	}
	s := string(b)
	if strings.Contains(s, flag) {
		return false, nil
	}

	lines := strings.Split(s, "\n")
	prefix := ""
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "- --") {
			prefix = line[:len(line)-len(trimmed)]
			break
		}
	}
	if prefix == "" {
		prefix = "    "
	}
	newLine := prefix + "- " + flag

	insertAt := -1
	for i, line := range lines {
		if strings.Contains(line, "--authorization-mode=") {
			insertAt = i + 1
			break
		}
	}
	if insertAt == -1 {
		for i, line := range lines {
			if strings.Contains(line, "- kube-apiserver") {
				insertAt = i + 1
				break
			}
		}
	}
	if insertAt == -1 {
		return false, fmt.Errorf("could not locate kube-apiserver command list in %s", controlPlaneManifestPath)
	}

	lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
	out := strings.Join(lines, "\n")

	tmp := controlPlaneManifestPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(out), 0644); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, controlPlaneManifestPath); err != nil {
		return false, err
	}
	return true, nil
}

func localAPIServerLivezStatus(ctx context.Context) (int, error) {
	c := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://127.0.0.1:6443/livez", nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp.StatusCode, nil
}

func formatHTTPStatus(status int, err error) string {
	if err != nil {
		return err.Error()
	}
	if status == 0 {
		return "unknown"
	}
	return fmt.Sprintf("%d %s", status, http.StatusText(status))
}

func isSecondaryControlPlane() bool {
	b, err := os.ReadFile(etcdManifestPath)
	if err != nil {
		return false
	}
	s := strings.ToLower(string(b))
	return strings.Contains(s, "--initial-cluster-state=existing")
}

func containsArg(parts []string, arg string) bool {
	for _, p := range parts {
		if p == arg {
			return true
		}
	}
	return false
}

func hasFlag(parts []string, flag string) bool {
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if p == flag {
			return true
		}
		if strings.HasPrefix(p, flag+"=") {
			return true
		}
	}
	return false
}

type ipAddrShowEntry struct {
	IfName   string `json:"ifname"`
	AddrInfo []struct {
		Family    string `json:"family"`
		Local     string `json:"local"`
		PrefixLen int    `json:"prefixlen"`
	} `json:"addr_info"`
}

func detectWireGuardAdvertiseAddress(ctx context.Context) (string, error) {
	
	if v := strings.TrimSpace(os.Getenv("GLUON_K8S_ADVERTISE_ADDRESS")); v != "" {
		return v, nil
	}

	
	loopbackPool, _ := netip.ParsePrefix("10.255.0.0/22")
	overlayPool, _ := netip.ParsePrefix("10.255.0.0/16")

	out, err := output(ctx, "ip", "-4", "-j", "addr", "show")
	if err != nil {
		return "", err
	}

	var entries []ipAddrShowEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return "", fmt.Errorf("failed to parse ip addr JSON: %w", err)
	}

	type candidate struct {
		ip    string
		score int
	}
	var best *candidate

	for _, e := range entries {
		for _, ai := range e.AddrInfo {
			if ai.Family != "inet" {
				continue
			}
			addr, err := netip.ParseAddr(strings.TrimSpace(ai.Local))
			if err != nil {
				continue
			}
			if !overlayPool.Contains(addr) {
				continue
			}

			
			score := 100
			if loopbackPool.Contains(addr) {
				score -= 50
			}
			if ai.PrefixLen == 32 {
				score -= 30
			}
			
			if strings.HasPrefix(e.IfName, "wg") {
				score -= 10
			}
			if e.IfName == "lo" {
				score -= 5
			}

			c := candidate{ip: addr.String(), score: score}
			if best == nil || c.score < best.score {
				tmp := c
				best = &tmp
			}
		}
	}

	if best == nil || best.ip == "" {
		return "", fmt.Errorf("no WireGuard/overlay IP found on this node (expected 10.255.0.0/16)")
	}
	return best.ip, nil
}

func detectBrokenKubeadmState() error {
	const (
		caCrt        = "/etc/kubernetes/pki/ca.crt"
		caKey        = "/etc/kubernetes/pki/ca.key"
		apiserverKey = "/etc/kubernetes/pki/apiserver.key"
	)

	if fileExists(caCrt) && !fileExists(caKey) {
		return fmt.Errorf("detected partial kubeadm PKI state (found %s but missing %s); run: kubeadm reset -f && rm -rf /etc/kubernetes /var/lib/etcd", caCrt, caKey)
	}
	if fileExists(adminConfPath) && fileExists(caCrt) && !fileExists(apiserverKey) {
		return fmt.Errorf("detected incomplete kubeadm state (missing %s); run: kubeadm reset -f && rm -rf /etc/kubernetes /var/lib/etcd", apiserverKey)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func maybeKickKubeletForCNI(ctx context.Context) {
	
	if !isJoined() {
		return
	}

	entries, err := os.ReadDir(cniNetDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".conf") || strings.HasSuffix(name, ".conflist") {
			return
		}
	}

	
	
	log.Printf("Kubernetes: CNI config missing in %s; restarting kubelet/containerd to recover", cniNetDir)
	_, _ = runLogged(ctx, "systemctl", "restart", "containerd")
	_, _ = runLogged(ctx, "systemctl", "restart", "kubelet")
}

func maybeEnsureControlPlaneLabels(ctx context.Context) {
	if !isControlPlaneNode() {
		return
	}
	if !fileExists(adminConfPath) {
		return
	}

	controlPlaneLabelMu.Lock()
	defer controlPlaneLabelMu.Unlock()
	if time.Since(lastControlPlaneLabelAttempt) < 2*time.Minute {
		return
	}
	lastControlPlaneLabelAttempt = time.Now()

	ensureControlPlaneLabels(ctx)
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n…(truncated)"
}

func shouldAutoResetOnRoleMismatch() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("GLUON_K8S_AUTO_RESET_ON_ROLE_MISMATCH")))
	if v == "" {
		
		return true
	}
	if v == "0" || v == "false" || v == "no" {
		return false
	}
	return v == "1" || v == "true" || v == "yes"
}

func resetKubeadmState(ctx context.Context) error {
	log.Println("Kubernetes: running kubeadm reset -f")
	if out, err := runLogged(ctx, "kubeadm", "reset", "-f"); err != nil {
		return fmt.Errorf("kubeadm reset failed: %w\n%s", err, truncate(out, 8000))
	}

	paths := []string{
		"/etc/kubernetes",
		"/var/lib/etcd",
		"/var/lib/kubelet/pki",
	}
	for _, p := range paths {
		_ = os.RemoveAll(p)
	}

	_, _ = runLogged(ctx, "systemctl", "restart", "containerd")
	_, _ = runLogged(ctx, "systemctl", "restart", "kubelet")
	return nil
}

func resolveSelfNodeName(ctx context.Context) string {
	hostname, _ := os.Hostname()
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return ""
	}
	short := strings.Split(hostname, ".")[0]
	if short == "" {
		short = hostname
	}

	
	out, err := runKubectlCaptured(ctx, "get", "node", short, "-o", "name")
	if err == nil && strings.TrimSpace(out) != "" {
		return short
	}

	
	jsonPath := "{.items[0].metadata.name}"
	out, err = runKubectlCaptured(ctx, "get", "nodes", "-l", "kubernetes.io/hostname="+short, "-o", "jsonpath="+jsonPath)
	if err == nil {
		name := strings.TrimSpace(out)
		if name != "" {
			return name
		}
	}

	return short
}

func ensureKubeletNodeIPPreJoin(ctx context.Context) {
	desired, err := detectWireGuardAdvertiseAddress(ctx)
	if err != nil || strings.TrimSpace(desired) == "" {
		return
	}

	changed, err := writeKubeletNodeIPDropIn(desired)
	if err != nil {
		log.Printf("Kubernetes: failed to pre-configure kubelet node-ip: %v", err)
		return
	}
	if changed {
		_, _ = runLogged(ctx, "systemctl", "daemon-reload")
	}
}

func maybeEnsureKubeletNodeIP(ctx context.Context) {
	if !isJoined() {
		return
	}

	kubeletNodeIPMu.Lock()
	defer kubeletNodeIPMu.Unlock()
	if time.Since(lastKubeletNodeIPAttempt) < 2*time.Minute {
		return
	}
	lastKubeletNodeIPAttempt = time.Now()

	desired, err := detectWireGuardAdvertiseAddress(ctx)
	if err != nil || strings.TrimSpace(desired) == "" {
		if err != nil {
			log.Printf("Kubernetes: failed to detect WireGuard node IP: %v", err)
		}
		return
	}

	current := readConfiguredKubeletNodeIP()
	if strings.TrimSpace(current) == strings.TrimSpace(desired) {
		return
	}

	updated := false
	if fileExists(kubeletKubeadmFlagsPath) {
		if err := upsertKubeadmFlagsNodeIP(desired); err != nil {
			log.Printf("Kubernetes: failed to update %s: %v", kubeletKubeadmFlagsPath, err)
		} else {
			updated = true
		}
	}

	if !updated {
		if _, err := writeKubeletNodeIPDropIn(desired); err != nil {
			log.Printf("Kubernetes: failed to install kubelet node-ip drop-in: %v", err)
			return
		}
	}

	log.Printf("Kubernetes: setting kubelet --node-ip to WireGuard IP %s (was %s)", desired, nonEmpty(current, "unset"))
	_, _ = runLogged(ctx, "systemctl", "daemon-reload")
	_, _ = runLogged(ctx, "systemctl", "restart", "kubelet")
}

func writeKubeletNodeIPDropIn(nodeIP string) (bool, error) {
	nodeIP = strings.TrimSpace(nodeIP)
	
	
	
	defaultKubeletPath := "/etc/default/kubelet"
	content := fmt.Sprintf("KUBELET_EXTRA_ARGS=--node-ip=%s\n", nodeIP)
	existing, _ := os.ReadFile(defaultKubeletPath)
	if string(existing) == content {
		return false, nil
	}
	tmp := defaultKubeletPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, defaultKubeletPath); err != nil {
		return false, err
	}
	return true, nil
}

func upsertKubeadmFlagsNodeIP(nodeIP string) error {
	b, err := os.ReadFile(kubeletKubeadmFlagsPath)
	if err != nil {
		return err
	}
	s := string(b)

	
	re := regexp.MustCompile(`(?m)^\s*KUBELET_KUBEADM_ARGS\s*=\s*"(.*)"\s*$`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		re2 := regexp.MustCompile(`(?m)^\s*KUBELET_KUBEADM_ARGS\s*=\s*(.*)\s*$`)
		m2 := re2.FindStringSubmatch(s)
		if len(m2) < 2 {
			return fmt.Errorf("could not parse KUBELET_KUBEADM_ARGS")
		}
		m = []string{m2[0], strings.Trim(m2[1], `"' `)}
	}

	args := m[1]
	
	
	reNode := regexp.MustCompile(`(?:^|\s)--node-ip(?:=|\s+)[0-9]{1,3}(?:\.[0-9]{1,3}){3}\b`)
	args = reNode.ReplaceAllString(args, "")
	args = strings.TrimSpace(args)
	args = strings.TrimSpace(args + " --node-ip=" + strings.TrimSpace(nodeIP))

	newLine := fmt.Sprintf(`KUBELET_KUBEADM_ARGS="%s"`, args)
	s = re.ReplaceAllString(s, newLine)
	if !strings.Contains(s, newLine) {
		
		reAny := regexp.MustCompile(`(?m)^\s*KUBELET_KUBEADM_ARGS\s*=.*$`)
		s = reAny.ReplaceAllString(s, newLine)
	}

	tmp := kubeletKubeadmFlagsPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(s), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, kubeletKubeadmFlagsPath)
}

func readConfiguredKubeletNodeIP() string {
	paths := []string{
		kubeletNodeIPDropInPath,
		"/etc/default/kubelet",
		kubeletKubeadmFlagsPath,
	}
	re := regexp.MustCompile(`--node-ip(?:=|\s+)([0-9]{1,3}(?:\.[0-9]{1,3}){3})`)
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		m := re.FindStringSubmatch(string(b))
		if len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
