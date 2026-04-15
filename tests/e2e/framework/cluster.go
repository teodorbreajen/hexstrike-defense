package framework

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// ClusterConfig holds the configuration for connecting to a Kubernetes cluster
type ClusterConfig struct {
	Kubeconfig    string
	Context       string
	Namespace     string
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

// DefaultClusterConfig returns a default cluster configuration
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		Kubeconfig:    os.Getenv("KUBECONFIG"),
		Context:       os.Getenv("KUBE_CONTEXT"),
		Namespace:     "hexstrike-agents",
		Timeout:       5 * time.Minute,
		RetryAttempts: 30,
		RetryDelay:    2 * time.Second,
	}
}

// Client holds the Kubernetes clientset and configuration
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *ClusterConfig
}

// NewClient creates a new Kubernetes client for e2e testing
func NewClient(cfg *ClusterConfig) (*Client, error) {
	if cfg == nil {
		cfg = DefaultClusterConfig()
	}

	var restConfig *restclient.Config
	var err error

	if cfg.Kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		restConfig, err = restclient.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    cfg,
	}, nil
}

// WaitForPodReady waits for a pod to be in the Ready state
func (c *Client) WaitForPodReady(ctx context.Context, namespace, podName string) error {
	return c.waitForPodCondition(ctx, namespace, podName, func(pod *corev1.Pod) bool {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	})
}

// WaitForPodTerminated waits for a pod to be terminated
func (c *Client) WaitForPodTerminated(ctx context.Context, namespace, podName string) error {
	return c.waitForPodCondition(ctx, namespace, podName, func(pod *corev1.Pod) bool {
		return pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed
	})
}

// waitForPodCondition waits for a pod to meet a specific condition
func (c *Client) waitForPodCondition(ctx context.Context, namespace, podName string, condition func(*corev1.Pod) bool) error {
	client := c.Clientset.CoreV1().Pods(namespace)

	for i := 0; i < c.Config.RetryAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pod, err := client.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			// Pod might not exist yet
			time.Sleep(c.Config.RetryDelay)
			continue
		}

		if condition(pod) {
			return nil
		}

		time.Sleep(c.Config.RetryDelay)
	}

	return fmt.Errorf("timeout waiting for pod %s/%s to meet condition", namespace, podName)
}

// GetPodLogs retrieves logs from a pod container
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName, containerName string) (string, error) {
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		TailLines: int64Ptr(100),
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 32*1024)
	var output bytes.Buffer
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return output.String(), nil
}

// CreateNamespace creates a new namespace
func (c *Client) CreateNamespace(ctx context.Context, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"hexstrike.io/managed": "true",
			},
		},
	}

	_, err := c.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// DeleteNamespace deletes a namespace
func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	return c.Clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

// ListPods lists all pods in a namespace
func (c *Client) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// PodExists checks if a pod exists and returns its status
func (c *Client) PodExists(ctx context.Context, namespace, podName string) (*corev1.Pod, bool, error) {
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, false, nil
	}
	return pod, true, nil
}

// ExecInPod executes a command in a pod container
func (c *Client) ExecInPod(ctx context.Context, namespace, podName, containerName string, command []string) (string, string, error) {
	req := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	opts := &corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
	}

	req.VersionedParams(opts, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(&c.Clientset.RESTClient().Config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, &remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), err
}

// Helper to get pointer to int64
func int64Ptr(i int64) *int64 {
	return &i
}
