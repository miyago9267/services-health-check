package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"services-health-check/internal/core/check"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type PodChecker struct {
	NameValue     string
	Namespace     string
	LabelSelector string
	Kubeconfig    string
	Context       string
	MinReady      int
}

func (c *PodChecker) Name() string {
	return c.NameValue
}

func (c *PodChecker) Check(ctx context.Context) (check.Result, error) {
	clientset, err := c.buildClient()
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: err.Error(), CheckedAt: time.Now()}, err
	}

	ns := c.Namespace
	if ns == "" {
		ns = "default"
	}
	if strings.EqualFold(ns, "all") || ns == "*" {
		ns = metav1.NamespaceAll
	}

	pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: c.LabelSelector})
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: err.Error(), CheckedAt: time.Now()}, err
	}

	total := len(pods.Items)
	ready := 0
	failed := 0
	pending := 0
	var problems []string

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case "Failed", "Unknown":
			failed++
			problems = append(problems, formatPodIssue(pod.Namespace, pod.Name, pod.Spec.NodeName, pod.Status.Phase, "", "", 0))
		case "Pending":
			pending++
			problems = append(problems, formatPodIssue(pod.Namespace, pod.Name, pod.Spec.NodeName, "Pending", "", "", 0))
		}

		allReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allReady = false
				reason := ""
				if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
					reason = cs.State.Waiting.Reason
				} else if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
					reason = cs.State.Terminated.Reason
				}
				problems = append(problems, formatPodIssue(pod.Namespace, pod.Name, pod.Spec.NodeName, pod.Status.Phase, cs.Name, reason, cs.RestartCount))
				break
			}
		}
		if allReady && pod.Status.Phase == "Running" {
			ready++
		}
	}

	status := check.StatusOK
	message := "pods healthy"

	if total == 0 {
		status = check.StatusWarn
		message = "找不到任何 Pod"
	} else if failed > 0 {
		status = check.StatusCrit
		message = fmt.Sprintf("有 %d 個 Pod 失敗", failed)
	} else if c.MinReady > 0 && ready < c.MinReady {
		status = check.StatusCrit
		message = fmt.Sprintf("就緒數不足：%d（最低 %d）", ready, c.MinReady)
	} else if pending > 0 || ready < total {
		status = check.StatusWarn
		message = fmt.Sprintf("就緒 %d/%d", ready, total)
	}

	if len(problems) > 0 {
		limit := 5
		if len(problems) < limit {
			limit = len(problems)
		}
		message = fmt.Sprintf("%s；例: %s", message, strings.Join(problems[:limit], "；"))
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   message,
		Metrics:   map[string]any{"total": total, "ready": ready, "failed": failed, "pending": pending, "problems": problems},
		CheckedAt: time.Now(),
	}, nil
}

func (c *PodChecker) buildClient() (*kubernetes.Clientset, error) {
	if c.Kubeconfig == "" {
		cfg, err := rest.InClusterConfig()
		if err == nil {
			return kubernetes.NewForConfig(cfg)
		}
	}

	kubeconfig := c.Kubeconfig
	if strings.HasPrefix(kubeconfig, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(home, kubeconfig[2:])
		}
	}
	if kubeconfig != "" {
		if _, err := os.Stat(kubeconfig); err != nil {
			return nil, fmt.Errorf("kubeconfig not found: %s", kubeconfig)
		}
	}

	loading := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	overrides := &clientcmd.ConfigOverrides{}
	if c.Context != "" {
		overrides.CurrentContext = c.Context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	cfg, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func formatPodIssue(namespace, podName, nodeName, phase, container, reason string, restarts int32) string {
	fields := []string{fmt.Sprintf("%s/%s", namespace, podName)}
	if phase != "" {
		fields = append(fields, "phase="+phase)
	}
	if container != "" {
		fields = append(fields, "container="+container)
	}
	if reason != "" {
		fields = append(fields, "reason="+reason)
	}
	if restarts > 0 {
		fields = append(fields, fmt.Sprintf("restarts=%d", restarts))
	}
	if nodeName != "" {
		fields = append(fields, "node="+nodeName)
	}
	return strings.Join(fields, " ")
}
