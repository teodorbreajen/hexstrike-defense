#!/bin/bash
#
# hexstrike-defense Validation Script
# Runs spec validation against running cluster
#
set -euo pipefail

NAMESPACE="${NAMESPACE:-hexstrike-system}"
MONITORING_NS="hexstrike-monitoring"
AGENTS_NS="hexstrike-agents"

echo "============================================="
echo "  hexstrike-defense Validation"
echo "============================================="
echo ""

# Function to check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        echo "ERROR: kubectl not found. Please install kubectl."
        exit 1
    fi
    echo "[OK] kubectl is available"
}

# Function to check cluster connectivity
check_cluster() {
    echo "Checking cluster connectivity..."
    if ! kubectl cluster-info &> /dev/null; then
        echo "ERROR: Cannot connect to cluster. Check kubeconfig."
        exit 1
    fi
    echo "[OK] Cluster is reachable"
}

# Validate Cilium policies
validate_cilium() {
    echo ""
    echo "--- Checking Cilium Network Policies ---"

    local cnp_count
    cnp_count=$(kubectl get cnp -A --no-headers 2>/dev/null | wc -l || echo "0")
    echo "Found $cnp_count CiliumClusterwideNetworkPolicy resources"

    # List all CNPs
    kubectl get cnp -A -o wide 2>/dev/null || echo "No CNPs found (Cilium may not be installed)"

    # Check DNS policy specifically
    if kubectl get cnp dns-whitelist -n kube-system &>/dev/null; then
        echo "[OK] DNS whitelist policy exists"
    else
        echo "[WARN] DNS whitelist policy not found"
    fi
}

# Validate Falco rules
validate_falco() {
    echo ""
    echo "--- Checking Falco Rules ---"

    local rules_count
    rules_count=$(kubectl get falcorules -A --no-headers 2>/dev/null | wc -l || echo "0")
    echo "Found $rules_count FalcoRule resources"

    # List all Falco rules
    kubectl get falcorules -A -o wide 2>/dev/null || echo "No Falco rules found (Falco may not be installed)"

    # Check for execve rules
    if kubectl get falcorules execve-rules -n hexstrike-system &>/dev/null; then
        echo "[OK] execve detection rules exist"
    else
        echo "[WARN] execve rules not found"
    fi
}

# Validate MCP proxy deployment
validate_mcp_proxy() {
    echo ""
    echo "--- Checking MCP Policy Proxy ---"

    # Check namespace exists
    if kubectl get namespace "$NAMESPACE" &>/dev/null; then
        echo "[OK] Namespace '$NAMESPACE' exists"
    else
        echo "[WARN] Namespace '$NAMESPACE' not found"
        return
    fi

    # Check deployment
    local replica_count
    replica_count=$(kubectl get deployment mcp-policy-proxy -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    echo "MCP Policy Proxy ready replicas: $replica_count"

    if [[ "$replica_count" -gt 0 ]]; then
        echo "[OK] MCP Policy Proxy is running"
    else
        echo "[WARN] MCP Policy Proxy is not ready"
    fi

    # Check pod status
    kubectl get pods -n "$NAMESPACE" -l app=mcp-policy-proxy -o wide 2>/dev/null || true
}

# Validate Lakera API connectivity
validate_lakera() {
    echo ""
    echo "--- Testing Lakera API Connectivity ---"

    local lakera_endpoint="http://mcp-policy-proxy.${NAMESPACE}:8080/health"

    if curl -sf "$lakera_endpoint" &>/dev/null; then
        echo "[OK] Lakera API is reachable"
    else
        echo "[WARN] Cannot reach Lakera API at $lakera_endpoint"
    fi
}

# Validate monitoring stack
validate_monitoring() {
    echo ""
    echo "--- Checking Monitoring Stack ---"

    # Check if monitoring namespace exists
    if kubectl get namespace "$MONITORING_NS" &>/dev/null; then
        echo "[OK] Monitoring namespace exists"
    else
        echo "[WARN] Monitoring namespace not found"
    fi

    # Check ServiceMonitor
    if kubectl get servicemonitor mcp-policy-proxy-metrics -n "$MONITORING_NS" &>/dev/null; then
        echo "[OK] Prometheus ServiceMonitor exists"
    else
        echo "[WARN] ServiceMonitor not found"
    fi
}

# Validate agent configurations
validate_agents() {
    echo ""
    echo "--- Checking Agent Configurations ---"

    if kubectl get namespace "$AGENTS_NS" &>/dev/null; then
        echo "[OK] Agents namespace exists"
    else
        echo "[WARN] Agents namespace not found"
    fi

    # Check ConfigMaps
    kubectl get configmap -n "$AGENTS_NS" -l app=hexstrike-defense 2>/dev/null || echo "No agent ConfigMaps found"
}

# Main validation run
main() {
    check_kubectl
    check_cluster
    validate_cilium
    validate_falco
    validate_mcp_proxy
    validate_lakera
    validate_monitoring
    validate_agents

    echo ""
    echo "============================================="
    echo "  Validation Complete"
    echo "============================================="
}

main "$@"