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
	ProblemLimit  int
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

	deploys, err := clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{LabelSelector: c.LabelSelector})
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: err.Error(), CheckedAt: time.Now()}, err
	}

	total := len(deploys.Items)
	ready := 0
	unready := 0
	var problems []string

	for _, deploy := range deploys.Items {
		desired := int32(1)
		if deploy.Spec.Replicas != nil {
			desired = *deploy.Spec.Replicas
		}
		readyReplicas := deploy.Status.ReadyReplicas
		if readyReplicas >= desired && desired > 0 {
			ready++
			continue
		}
		unready++
		problems = append(problems, fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name))
	}

	status := check.StatusOK
	message := "deployments healthy"

	if total == 0 {
		status = check.StatusWarn
		message = "找不到任何 Deployment"
	} else if c.MinReady > 0 && ready < c.MinReady {
		status = check.StatusCrit
		message = fmt.Sprintf("就緒數不足：%d（最低 %d）", ready, c.MinReady)
	} else if unready > 0 || ready < total {
		status = check.StatusWarn
		message = fmt.Sprintf("就緒 %d/%d", ready, total)
	}

	if len(problems) > 0 {
		limit := c.ProblemLimit
		if limit <= 0 {
			limit = 5
		}
		if len(problems) < limit {
			limit = len(problems)
		}
		message = fmt.Sprintf("%s；例: %s", message, strings.Join(problems[:limit], "；"))
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   message,
		Metrics:   map[string]any{"total": total, "ready": ready, "unready": unready, "problems": problems},
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

// formatPodIssue removed: use deployment-level readiness only.
