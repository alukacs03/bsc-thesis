package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

type k8sObjectMeta struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

type k8sList[T any] struct {
	Items []T `json:"items"`
}

type k8sDeployment struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		Replicas *int32 `json:"replicas"`
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		Replicas        *int32 `json:"replicas"`
		ReadyReplicas   *int32 `json:"readyReplicas"`
		UpdatedReplicas *int32 `json:"updatedReplicas"`
	} `json:"status"`
}

type k8sStatefulSet struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		Replicas *int32 `json:"replicas"`
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		Replicas      *int32 `json:"replicas"`
		ReadyReplicas *int32 `json:"readyReplicas"`
	} `json:"status"`
}

type k8sDaemonSet struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		DesiredNumberScheduled *int32 `json:"desiredNumberScheduled"`
		NumberReady            *int32 `json:"numberReady"`
	} `json:"status"`
}

type k8sJob struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Status   struct {
		Active    *int32 `json:"active"`
		Succeeded *int32 `json:"succeeded"`
		Failed    *int32 `json:"failed"`
	} `json:"status"`
}

type k8sPod struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		NodeName   string `json:"nodeName"`
		Containers []struct {
			Image string `json:"image"`
			Name  string `json:"name"`
		} `json:"containers"`
	} `json:"spec"`
	Status struct {
		Phase             string `json:"phase"`
		Reason            string `json:"reason"`
		Message           string `json:"message"`
		ContainerStatuses []struct {
			Name         string `json:"name"`
			RestartCount int32  `json:"restartCount"`
			State        struct {
				Waiting struct {
					Reason  string `json:"reason"`
					Message string `json:"message"`
				} `json:"waiting"`
				Terminated struct {
					Reason  string `json:"reason"`
					Message string `json:"message"`
				} `json:"terminated"`
			} `json:"state"`
			LastState struct {
				Terminated struct {
					Reason  string `json:"reason"`
					Message string `json:"message"`
				} `json:"terminated"`
			} `json:"lastState"`
		} `json:"containerStatuses"`
		Conditions []struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			Reason string `json:"reason"`
		} `json:"conditions"`
	} `json:"status"`
}

type workloadNamespaceSummary struct {
	Namespace string `json:"namespace"`

	DeploymentsTotal int `json:"deployments_total"`
	DeploymentsReady int `json:"deployments_ready"`

	StatefulSetsTotal int `json:"statefulsets_total"`
	StatefulSetsReady int `json:"statefulsets_ready"`

	DaemonSetsTotal int `json:"daemonsets_total"`
	DaemonSetsReady int `json:"daemonsets_ready"`

	JobsTotal     int `json:"jobs_total"`
	JobsActive    int `json:"jobs_active"`
	JobsSucceeded int `json:"jobs_succeeded"`
	JobsFailed    int `json:"jobs_failed"`

	PodsTotal     int `json:"pods_total"`
	PodsRunning   int `json:"pods_running"`
	PodsPending   int `json:"pods_pending"`
	PodsSucceeded int `json:"pods_succeeded"`
	PodsFailed    int `json:"pods_failed"`
	PodsUnhealthy int `json:"pods_unhealthy"`

	RestartsTotal int `json:"restarts_total"`
}

type workloadPodIssue struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Node      string `json:"node,omitempty"`
	Phase     string `json:"phase"`
	Reason    string `json:"reason,omitempty"`
	Message   string `json:"message,omitempty"`
	Images    []string `json:"images,omitempty"`
	Restarts  int    `json:"restarts"`
	AgeSeconds int64 `json:"age_seconds"`
}

type workloadNodeSummary struct {
	Node          string `json:"node"`
	Pods          int    `json:"pods"`
	UnhealthyPods int    `json:"unhealthy_pods"`
}

type workloadResourceInfo struct {
	Namespace     string   `json:"namespace"`
	Name          string   `json:"name"`
	Kind          string   `json:"kind"`
	Ready         string   `json:"ready"`
	Images        []string `json:"images,omitempty"`
	AgeSeconds    int64    `json:"age_seconds"`
}

type kubernetesWorkloadsResponse struct {
	GeneratedAt   time.Time                  `json:"generated_at"`
	Namespaces    []workloadNamespaceSummary `json:"namespaces"`
	Nodes         []workloadNodeSummary      `json:"nodes"`
	UnhealthyPods []workloadPodIssue         `json:"unhealthy_pods"`
	Resources     []workloadResourceInfo     `json:"resources"`
}

func AdminGetKubernetesWorkloads(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	var deployments k8sList[k8sDeployment]
	if err := kubectlJSON(ctx, []string{"get", "deployments", "-A", "-o", "json"}, &deployments); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	var daemonsets k8sList[k8sDaemonSet]
	if err := kubectlJSON(ctx, []string{"get", "daemonsets", "-A", "-o", "json"}, &daemonsets); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	var statefulsets k8sList[k8sStatefulSet]
	if err := kubectlJSON(ctx, []string{"get", "statefulsets", "-A", "-o", "json"}, &statefulsets); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	var jobs k8sList[k8sJob]
	if err := kubectlJSON(ctx, []string{"get", "jobs", "-A", "-o", "json"}, &jobs); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	var pods k8sList[k8sPod]
	if err := kubectlJSON(ctx, []string{"get", "pods", "-A", "-o", "json"}, &pods); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	byNS := map[string]*workloadNamespaceSummary{}
	getNS := func(ns string) *workloadNamespaceSummary {
		if ns == "" {
			ns = "default"
		}
		if v, ok := byNS[ns]; ok {
			return v
		}
		v := &workloadNamespaceSummary{Namespace: ns}
		byNS[ns] = v
		return v
	}

	for _, d := range deployments.Items {
		ns := getNS(d.Metadata.Namespace)
		ns.DeploymentsTotal++
		desired := derefInt32(d.Spec.Replicas)
		ready := derefInt32(d.Status.ReadyReplicas)
		if desired == 0 && ready == 0 {
			
		} else if ready >= desired {
			ns.DeploymentsReady++
		}
	}

	for _, ds := range daemonsets.Items {
		ns := getNS(ds.Metadata.Namespace)
		ns.DaemonSetsTotal++
		desired := derefInt32(ds.Status.DesiredNumberScheduled)
		ready := derefInt32(ds.Status.NumberReady)
		if desired == 0 && ready == 0 {
			
		} else if ready >= desired {
			ns.DaemonSetsReady++
		}
	}

	for _, s := range statefulsets.Items {
		ns := getNS(s.Metadata.Namespace)
		ns.StatefulSetsTotal++
		desired := derefInt32(s.Spec.Replicas)
		ready := derefInt32(s.Status.ReadyReplicas)
		if desired == 0 && ready == 0 {
			
		} else if ready >= desired {
			ns.StatefulSetsReady++
		}
	}

	for _, j := range jobs.Items {
		ns := getNS(j.Metadata.Namespace)
		ns.JobsTotal++
		ns.JobsActive += int(derefInt32(j.Status.Active))
		ns.JobsSucceeded += int(derefInt32(j.Status.Succeeded))
		ns.JobsFailed += int(derefInt32(j.Status.Failed))
	}

	nodeCounts := map[string]*workloadNodeSummary{}
	var unhealthy []workloadPodIssue

	for _, p := range pods.Items {
		ns := getNS(p.Metadata.Namespace)
		ns.PodsTotal++

		switch strings.ToLower(strings.TrimSpace(p.Status.Phase)) {
		case "running":
			ns.PodsRunning++
		case "pending":
			ns.PodsPending++
		case "succeeded":
			ns.PodsSucceeded++
		case "failed":
			ns.PodsFailed++
		default:
			
		}

		restarts := 0
		waitingReason := ""
		waitingMessage := ""
		for _, cs := range p.Status.ContainerStatuses {
			restarts += int(cs.RestartCount)
			if cs.State.Waiting.Reason != "" && waitingReason == "" {
				waitingReason = cs.State.Waiting.Reason
				waitingMessage = cs.State.Waiting.Message
			}
			if cs.State.Terminated.Reason != "" && waitingReason == "" {
				waitingReason = cs.State.Terminated.Reason
				waitingMessage = cs.State.Terminated.Message
			}
		}
		ns.RestartsTotal += restarts

		nodeName := strings.TrimSpace(p.Spec.NodeName)
		if nodeName != "" {
			if _, ok := nodeCounts[nodeName]; !ok {
				nodeCounts[nodeName] = &workloadNodeSummary{Node: nodeName}
			}
			nodeCounts[nodeName].Pods++
		}

		ready := podReady(&p)
		phaseLower := strings.ToLower(strings.TrimSpace(p.Status.Phase))
		unhealthyPod := phaseLower == "failed" || phaseLower == "pending" || (phaseLower == "running" && !ready)
		if unhealthyPod {
			ns.PodsUnhealthy++
			reason := strings.TrimSpace(p.Status.Reason)
			msg := strings.TrimSpace(p.Status.Message)
			if waitingReason != "" {
				reason = waitingReason
				if waitingMessage != "" {
					msg = waitingMessage
				}
			}
			imagesSet := map[string]struct{}{}
			for _, c := range p.Spec.Containers {
				img := strings.TrimSpace(c.Image)
				if img == "" {
					continue
				}
				imagesSet[img] = struct{}{}
			}
			images := make([]string, 0, len(imagesSet))
			for img := range imagesSet {
				images = append(images, img)
			}
			sort.Strings(images)
			age := time.Since(p.Metadata.CreationTimestamp)
			issue := workloadPodIssue{
				Namespace: p.Metadata.Namespace,
				Name:      p.Metadata.Name,
				Node:      nodeName,
				Phase:     p.Status.Phase,
				Reason:    reason,
				Message:   msg,
				Images:    images,
				Restarts:  restarts,
				AgeSeconds: int64(age.Seconds()),
			}
			unhealthy = append(unhealthy, issue)
			if nodeName != "" {
				nodeCounts[nodeName].UnhealthyPods++
			}
		}
	}

	namespaces := make([]workloadNamespaceSummary, 0, len(byNS))
	for _, v := range byNS {
		namespaces = append(namespaces, *v)
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Namespace < namespaces[j].Namespace })

	nodes := make([]workloadNodeSummary, 0, len(nodeCounts))
	for _, v := range nodeCounts {
		nodes = append(nodes, *v)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Pods > nodes[j].Pods })

	sort.Slice(unhealthy, func(i, j int) bool {
		
		pi := strings.ToLower(unhealthy[i].Phase)
		pj := strings.ToLower(unhealthy[j].Phase)
		score := func(p string) int {
			switch p {
			case "failed":
				return 0
			case "pending":
				return 1
			default:
				return 2
			}
		}
		if score(pi) != score(pj) {
			return score(pi) < score(pj)
		}
		if unhealthy[i].Restarts != unhealthy[j].Restarts {
			return unhealthy[i].Restarts > unhealthy[j].Restarts
		}
		return unhealthy[i].AgeSeconds < unhealthy[j].AgeSeconds
	})
	if len(unhealthy) > 50 {
		unhealthy = unhealthy[:50]
	}

	
	var resources []workloadResourceInfo
	for _, d := range deployments.Items {
		images := extractImages(d.Spec.Template.Spec.Containers)
		resources = append(resources, workloadResourceInfo{
			Namespace:  d.Metadata.Namespace,
			Name:       d.Metadata.Name,
			Kind:       "Deployment",
			Ready:      fmt.Sprintf("%d/%d", derefInt32(d.Status.ReadyReplicas), derefInt32(d.Spec.Replicas)),
			Images:     images,
			AgeSeconds: int64(time.Since(d.Metadata.CreationTimestamp).Seconds()),
		})
	}
	for _, ds := range daemonsets.Items {
		images := extractImages(ds.Spec.Template.Spec.Containers)
		resources = append(resources, workloadResourceInfo{
			Namespace:  ds.Metadata.Namespace,
			Name:       ds.Metadata.Name,
			Kind:       "DaemonSet",
			Ready:      fmt.Sprintf("%d/%d", derefInt32(ds.Status.NumberReady), derefInt32(ds.Status.DesiredNumberScheduled)),
			Images:     images,
			AgeSeconds: int64(time.Since(ds.Metadata.CreationTimestamp).Seconds()),
		})
	}
	for _, s := range statefulsets.Items {
		images := extractImages(s.Spec.Template.Spec.Containers)
		resources = append(resources, workloadResourceInfo{
			Namespace:  s.Metadata.Namespace,
			Name:       s.Metadata.Name,
			Kind:       "StatefulSet",
			Ready:      fmt.Sprintf("%d/%d", derefInt32(s.Status.ReadyReplicas), derefInt32(s.Spec.Replicas)),
			Images:     images,
			AgeSeconds: int64(time.Since(s.Metadata.CreationTimestamp).Seconds()),
		})
	}

	sort.Slice(resources, func(i, j int) bool {
		if resources[i].Namespace != resources[j].Namespace {
			return resources[i].Namespace < resources[j].Namespace
		}
		if resources[i].Kind != resources[j].Kind {
			return resources[i].Kind < resources[j].Kind
		}
		return resources[i].Name < resources[j].Name
	})

	return c.JSON(kubernetesWorkloadsResponse{
		GeneratedAt:   time.Now(),
		Namespaces:    namespaces,
		Nodes:         nodes,
		UnhealthyPods: unhealthy,
		Resources:     resources,
	})
}

func extractImages(containers []struct {
	Image string `json:"image"`
}) []string {
	var images []string
	seen := map[string]bool{}
	for _, c := range containers {
		img := strings.TrimSpace(c.Image)
		if img != "" && !seen[img] {
			images = append(images, img)
			seen[img] = true
		}
	}
	return images
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func podReady(p *k8sPod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == "Ready" {
			return strings.ToLower(c.Status) == "true"
		}
	}
	return false
}

func kubectlJSON(ctx context.Context, args []string, out any) error {
	kubeconfig := strings.TrimSpace(os.Getenv("GLUON_KUBECONFIG"))
	if kubeconfig == "" {
		kubeconfig = "/etc/kubernetes/admin.conf"
	}

	cmdArgs := append([]string{"--kubeconfig", kubeconfig}, args...)
	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	b, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(b))
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("kubectl timeout")
		}
		return fmt.Errorf("kubectl failed: %s", msg)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("failed to parse kubectl output: %w", err)
	}
	return nil
}

func kubectlRaw(ctx context.Context, args []string) (string, error) {
	kubeconfig := strings.TrimSpace(os.Getenv("GLUON_KUBECONFIG"))
	if kubeconfig == "" {
		kubeconfig = "/etc/kubernetes/admin.conf"
	}

	cmdArgs := append([]string{"--kubeconfig", kubeconfig}, args...)
	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	b, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(b))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return output, fmt.Errorf("kubectl timeout")
		}
		if output == "" {
			output = err.Error()
		}
		return output, fmt.Errorf("kubectl failed: %s", output)
	}
	return output, nil
}

type applyManifestInput struct {
	YAML string `json:"yaml"`
}

type applyManifestResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func AdminApplyKubernetesManifest(c *fiber.Ctx) error {
	var input applyManifestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	yaml := strings.TrimSpace(input.YAML)
	if yaml == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "YAML content is required"})
	}

	
	tmpFile, err := os.CreateTemp("", "kubectl-apply-*.yaml")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create temp file"})
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yaml); err != nil {
		tmpFile.Close()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write temp file"})
	}
	tmpFile.Close()

	
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	output, err := kubectlRaw(ctx, []string{"apply", "-f", tmpFile.Name()})
	if err != nil {
		return c.JSON(applyManifestResponse{
			Success: false,
			Output:  output,
			Error:   err.Error(),
		})
	}

	return c.JSON(applyManifestResponse{
		Success: true,
		Output:  output,
	})
}

type getResourceYAMLResponse struct {
	YAML  string `json:"yaml"`
	Error string `json:"error,omitempty"`
}

type k8sResourceWithAnnotations struct {
	Metadata struct {
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`
}


type k8sService struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		Type         string `json:"type"`
		ClusterIP    string `json:"clusterIP"`
		ExternalIP   string `json:"externalIP,omitempty"`
		Ports        []struct {
			Name       string `json:"name,omitempty"`
			Port       int32  `json:"port"`
			TargetPort any    `json:"targetPort"` 
			NodePort   int32  `json:"nodePort,omitempty"`
			Protocol   string `json:"protocol"`
		} `json:"ports"`
		Selector map[string]string `json:"selector,omitempty"`
	} `json:"spec"`
	Status struct {
		LoadBalancer struct {
			Ingress []struct {
				IP       string `json:"ip,omitempty"`
				Hostname string `json:"hostname,omitempty"`
			} `json:"ingress,omitempty"`
		} `json:"loadBalancer,omitempty"`
	} `json:"status"`
}

type k8sIngress struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     struct {
		IngressClassName *string `json:"ingressClassName,omitempty"`
		TLS              []struct {
			Hosts      []string `json:"hosts,omitempty"`
			SecretName string   `json:"secretName,omitempty"`
		} `json:"tls,omitempty"`
		Rules []struct {
			Host string `json:"host,omitempty"`
			HTTP *struct {
				Paths []struct {
					Path     string `json:"path,omitempty"`
					PathType string `json:"pathType,omitempty"`
					Backend  struct {
						Service *struct {
							Name string `json:"name"`
							Port struct {
								Number int32  `json:"number,omitempty"`
								Name   string `json:"name,omitempty"`
							} `json:"port"`
						} `json:"service,omitempty"`
					} `json:"backend"`
				} `json:"paths,omitempty"`
			} `json:"http,omitempty"`
		} `json:"rules,omitempty"`
	} `json:"spec"`
	Status struct {
		LoadBalancer struct {
			Ingress []struct {
				IP       string `json:"ip,omitempty"`
				Hostname string `json:"hostname,omitempty"`
			} `json:"ingress,omitempty"`
		} `json:"loadBalancer,omitempty"`
	} `json:"status"`
}

type serviceInfo struct {
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ClusterIP  string   `json:"cluster_ip"`
	ExternalIP string   `json:"external_ip,omitempty"`
	Ports      []string `json:"ports"`
	AgeSeconds int64    `json:"age_seconds"`
}

type ingressInfo struct {
	Namespace      string   `json:"namespace"`
	Name           string   `json:"name"`
	IngressClass   string   `json:"ingress_class,omitempty"`
	Hosts          []string `json:"hosts"`
	Address        string   `json:"address,omitempty"`
	TLS            bool     `json:"tls"`
	AgeSeconds     int64    `json:"age_seconds"`
}

type ingressRuleInfo struct {
	Host        string `json:"host"`
	Path        string `json:"path"`
	PathType    string `json:"path_type,omitempty"`
	ServiceName string `json:"service_name"`
	ServicePort string `json:"service_port"`
}

type ingressDetailInfo struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	IngressClass string            `json:"ingress_class,omitempty"`
	TLS          bool              `json:"tls"`
	TLSHosts     []string          `json:"tls_hosts,omitempty"`
	Rules        []ingressRuleInfo `json:"rules"`
	Address      string            `json:"address,omitempty"`
	AgeSeconds   int64             `json:"age_seconds"`
}

type kubernetesNetworkingResponse struct {
	GeneratedAt time.Time           `json:"generated_at"`
	Services    []serviceInfo       `json:"services"`
	Ingresses   []ingressDetailInfo `json:"ingresses"`
}

func AdminGetKubernetesNetworking(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	var services k8sList[k8sService]
	if err := kubectlJSON(ctx, []string{"get", "services", "-A", "-o", "json"}, &services); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	var ingresses k8sList[k8sIngress]
	if err := kubectlJSON(ctx, []string{"get", "ingresses", "-A", "-o", "json"}, &ingresses); err != nil {
		
		ingresses = k8sList[k8sIngress]{}
	}

	
	var svcList []serviceInfo
	for _, s := range services.Items {
		var ports []string
		for _, p := range s.Spec.Ports {
			portStr := fmt.Sprintf("%d/%s", p.Port, strings.ToUpper(p.Protocol))
			if p.NodePort > 0 {
				portStr = fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, strings.ToUpper(p.Protocol))
			}
			ports = append(ports, portStr)
		}

		externalIP := ""
		if len(s.Status.LoadBalancer.Ingress) > 0 {
			if s.Status.LoadBalancer.Ingress[0].IP != "" {
				externalIP = s.Status.LoadBalancer.Ingress[0].IP
			} else if s.Status.LoadBalancer.Ingress[0].Hostname != "" {
				externalIP = s.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		svcList = append(svcList, serviceInfo{
			Namespace:  s.Metadata.Namespace,
			Name:       s.Metadata.Name,
			Type:       s.Spec.Type,
			ClusterIP:  s.Spec.ClusterIP,
			ExternalIP: externalIP,
			Ports:      ports,
			AgeSeconds: int64(time.Since(s.Metadata.CreationTimestamp).Seconds()),
		})
	}

	sort.Slice(svcList, func(i, j int) bool {
		if svcList[i].Namespace != svcList[j].Namespace {
			return svcList[i].Namespace < svcList[j].Namespace
		}
		return svcList[i].Name < svcList[j].Name
	})

	
	var ingList []ingressDetailInfo
	for _, ing := range ingresses.Items {
		ingressClass := ""
		if ing.Spec.IngressClassName != nil {
			ingressClass = *ing.Spec.IngressClassName
		}

		address := ""
		if len(ing.Status.LoadBalancer.Ingress) > 0 {
			if ing.Status.LoadBalancer.Ingress[0].IP != "" {
				address = ing.Status.LoadBalancer.Ingress[0].IP
			} else if ing.Status.LoadBalancer.Ingress[0].Hostname != "" {
				address = ing.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		hasTLS := len(ing.Spec.TLS) > 0
		var tlsHosts []string
		for _, tls := range ing.Spec.TLS {
			tlsHosts = append(tlsHosts, tls.Hosts...)
		}

		var rules []ingressRuleInfo
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				serviceName := ""
				servicePort := ""
				if path.Backend.Service != nil {
					serviceName = path.Backend.Service.Name
					if path.Backend.Service.Port.Number > 0 {
						servicePort = fmt.Sprintf("%d", path.Backend.Service.Port.Number)
					} else if path.Backend.Service.Port.Name != "" {
						servicePort = path.Backend.Service.Port.Name
					}
				}
				rules = append(rules, ingressRuleInfo{
					Host:        rule.Host,
					Path:        path.Path,
					PathType:    path.PathType,
					ServiceName: serviceName,
					ServicePort: servicePort,
				})
			}
		}

		ingList = append(ingList, ingressDetailInfo{
			Namespace:    ing.Metadata.Namespace,
			Name:         ing.Metadata.Name,
			IngressClass: ingressClass,
			TLS:          hasTLS,
			TLSHosts:     tlsHosts,
			Rules:        rules,
			Address:      address,
			AgeSeconds:   int64(time.Since(ing.Metadata.CreationTimestamp).Seconds()),
		})
	}

	sort.Slice(ingList, func(i, j int) bool {
		if ingList[i].Namespace != ingList[j].Namespace {
			return ingList[i].Namespace < ingList[j].Namespace
		}
		return ingList[i].Name < ingList[j].Name
	})

	return c.JSON(kubernetesNetworkingResponse{
		GeneratedAt: time.Now(),
		Services:    svcList,
		Ingresses:   ingList,
	})
}

func AdminGetKubernetesResourceYAML(c *fiber.Ctx) error {
	namespace := c.Query("namespace")
	kind := c.Query("kind")
	name := c.Query("name")

	if namespace == "" || kind == "" || name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "namespace, kind, and name are required"})
	}

	
	validKinds := map[string]bool{
		"deployment":  true,
		"statefulset": true,
		"daemonset":   true,
		"service":     true,
		"configmap":   true,
		"secret":      true,
		"ingress":     true,
		"job":         true,
		"cronjob":     true,
		"pod":         true,
	}
	kindLower := strings.ToLower(kind)
	if !validKinds[kindLower] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid resource kind"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	output, err := kubectlRaw(ctx, []string{"get", kindLower, name, "-n", namespace, "-o", "yaml"})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(getResourceYAMLResponse{
			Error: output,
		})
	}

	var resource k8sResourceWithAnnotations
	if err := yaml.Unmarshal([]byte(output), &resource); err == nil {
		if applied := strings.TrimSpace(resource.Metadata.Annotations["kubectl.kubernetes.io/last-applied-configuration"]); applied != "" {
			var appliedObj any
			if err := json.Unmarshal([]byte(applied), &appliedObj); err == nil {
				if formatted, err := yaml.Marshal(appliedObj); err == nil {
					return c.JSON(getResourceYAMLResponse{
						YAML: string(formatted),
					})
				}
			}
			return c.JSON(getResourceYAMLResponse{
				YAML: applied,
			})
		}
	}

	return c.JSON(getResourceYAMLResponse{
		YAML: output,
	})
}

type deleteResourceResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func AdminDeleteKubernetesResource(c *fiber.Ctx) error {
	namespace := c.Query("namespace")
	kind := c.Query("kind")
	name := c.Query("name")

	if namespace == "" || kind == "" || name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "namespace, kind, and name are required"})
	}

	
	validKinds := map[string]bool{
		"deployment":  true,
		"statefulset": true,
		"daemonset":   true,
		"service":     true,
		"configmap":   true,
		"secret":      true,
		"ingress":     true,
		"job":         true,
		"cronjob":     true,
		"pod":         true,
	}
	kindLower := strings.ToLower(kind)
	if !validKinds[kindLower] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid resource kind"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	output, err := kubectlRaw(ctx, []string{"delete", kindLower, name, "-n", namespace})
	if err != nil {
		return c.JSON(deleteResourceResponse{
			Success: false,
			Output:  output,
			Error:   err.Error(),
		})
	}

	return c.JSON(deleteResourceResponse{
		Success: true,
		Output:  output,
	})
}
