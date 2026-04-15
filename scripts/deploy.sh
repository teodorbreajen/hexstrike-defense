#!/bin/bash
#
# hexstrike-defense Deployment Script
# Deploys the complete defense stack to Kubernetes
#
set -euo pipefail

NAMESPACE="${NAMESPACE:-hexstrike-system}"
MONITORING_NS="hexstrike-monitoring"
AGENTS_NS="hexstrike-agents"
RELEASE_NAME="hexstrike-defense"
CHART_PATH="./manifests/charts/hexstrike-defense"
TIMEOUT="${TIMEOUT:-5m}"

echo "============================================="
echo "  hexstrike-defense Deployment"
echo "============================================="
echo ""
echo "Namespace: $NAMESPACE"
echo "Chart: $CHART_PATH"
echo "Timeout: $TIMEOUT"
echo ""

# Check prerequisites
check_prerequisites() {
    echo "--- Checking Prerequisites ---"

    if ! command -v kubectl &> /dev/null; then
        echo "ERROR: kubectl not found. Please install kubectl."
        exit 1
    fi
    echo "[OK] kubectl"

    if ! command -v helm &> /dev/null; then
        echo "ERROR: helm not found. Please install helm."
        exit 1
    fi
    echo "[OK] helm"

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        echo "ERROR: Cannot connect to cluster. Check kubeconfig."
        exit 1
    fi
    echo "[OK] Cluster connectivity"

    # Check cluster capabilities
    echo ""
    echo "--- Checking Cluster Capabilities ---"
    kubectl version --client 2>/dev/null || true
    kubectl get nodes 2>/dev/null || true
}

# Create namespaces
create_namespaces() {
    echo ""
    echo "--- Creating Namespaces ---"

    for ns in "$NAMESPACE" "$MONITORING_NS" "$AGENTS_NS"; do
        if kubectl get namespace "$ns" &>/dev/null; then
            echo "[SKIP] Namespace '$ns' already exists"
        else
            kubectl create namespace "$ns"
            echo "[OK] Created namespace '$ns'"
        fi
    done
}

# Deploy Cilium policies
deploy_cilium() {
    echo ""
    echo "--- Deploying Cilium Policies ---"

    local cilium_manifests="manifests/cilium"
    if [[ -d "$cilium_manifests" ]]; then
        for manifest in $(ls "$cilium_manifests"/*.yaml 2>/dev/null | sort); do
            echo "Applying $(basename "$manifest")..."
            kubectl apply -f "$manifest" --server-side 2>/dev/null || kubectl apply -f "$manifest"
        done
        echo "[OK] Cilium policies deployed"
    else
        echo "[WARN] Cilium manifests not found at $cilium_manifests"
    fi
}

# Deploy Falco rules
deploy_falco() {
    echo ""
    echo "--- Deploying Falco Rules ---"

    local falco_manifests="manifests/falco"
    if [[ -d "$falco_manifests" ]]; then
        for manifest in $(ls "$falco_manifests"/*.yaml 2>/dev/null | sort); do
            echo "Applying $(basename "$manifest")..."
            kubectl apply -f "$manifest" --server-side 2>/dev/null || kubectl apply -f "$manifest"
        done
        echo "[OK] Falco rules deployed"
    else
        echo "[WARN] Falco manifests not found at $falco_manifests"
    fi
}

# Deploy MCP proxy
deploy_mcp_proxy() {
    echo ""
    echo "--- Deploying MCP Policy Proxy ---"

    local mcp_manifests="manifests/mcp-proxy"
    if [[ -d "$mcp_manifests" ]]; then
        for manifest in $(ls "$mcp_manifests"/*.yaml 2>/dev/null | sort); do
            echo "Applying $(basename "$manifest")..."
            kubectl apply -f "$manifest" --server-side 2>/dev/null || kubectl apply -f "$manifest"
        done
        echo "[OK] MCP proxy deployed"
    else
        echo "[WARN] MCP proxy manifests not found at $mcp_manifests"
    fi
}

# Deploy LangGraph agent configs
deploy_agent_configs() {
    echo ""
    echo "--- Deploying Agent Configurations ---"

    local langgraph_manifests="manifests/langgraph"
    if [[ -d "$langgraph_manifests" ]]; then
        for manifest in $(ls "$langgraph_manifests"/*.yaml 2>/dev/null | sort); do
            echo "Applying $(basename "$manifest")..."
            kubectl apply -f "$manifest" --server-side 2>/dev/null || kubectl apply -f "$manifest"
        done
        echo "[OK] Agent configurations deployed"
    else
        echo "[WARN] LangGraph manifests not found at $langgraph_manifests"
    fi
}

# Install Helm chart (optional - can be used instead of individual manifests)
install_helm_chart() {
    echo ""
    echo "--- Installing Helm Chart (Optional) ---"

    if [[ ! -d "$CHART_PATH" ]]; then
        echo "[WARN] Helm chart not found at $CHART_PATH"
        return
    fi

    # Check if values file exists
    local values_file="$CHART_PATH/values.yaml"
    if [[ ! -f "$values_file" ]]; then
        echo "[WARN] Values file not found at $values_file"
    fi

    # Install or upgrade
    if helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
        echo "Upgrading existing release..."
        helm upgrade "$RELEASE_NAME" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            --values "$values_file" \
            --timeout "$TIMEOUT" \
            --wait
    else
        echo "Installing new release..."
        helm install "$RELEASE_NAME" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            --create-namespace \
            --values "$values_file" \
            --timeout "$TIMEOUT" \
            --wait
    fi
    echo "[OK] Helm chart deployed"
}

# Wait for pods to be ready
wait_for_pods() {
    echo ""
    echo "--- Waiting for Pods ---"

    local timeout_seconds=120
    local label="app=mcp-policy-proxy"

    echo "Waiting for pods with label '$label' in namespace '$NAMESPACE'..."

    if kubectl wait --for=condition=ready pod \
        -l "$label" \
        -n "$NAMESPACE" \
        --timeout="${timeout_seconds}s" 2>/dev/null; then
        echo "[OK] Pods are ready"
    else
        echo "[WARN] Timeout waiting for pods. Checking status..."
        kubectl get pods -n "$NAMESPACE" -l "$label" -o wide
    fi
}

# Health check verification
health_check() {
    echo ""
    echo "--- Running Health Checks ---"

    local mcp_proxy_endpoint="http://mcp-policy-proxy.${NAMESPACE}:8080/health"

    # Try to reach MCP proxy health endpoint
    if curl -sf "$mcp_proxy_endpoint" &>/dev/null; then
        echo "[OK] MCP Policy Proxy health check passed"
    else
        echo "[WARN] MCP Policy Proxy health check failed"
        echo "       Endpoint: $mcp_proxy_endpoint"
    fi

    # Check all pods in namespace
    echo ""
    echo "--- Pod Status ---"
    kubectl get pods -n "$NAMESPACE" -o wide 2>/dev/null || true
}

# Display deployment summary
deployment_summary() {
    echo ""
    echo "============================================="
    echo "  Deployment Summary"
    echo "============================================="
    echo ""
    echo "Namespaces created:"
    echo "  - $NAMESPACE"
    echo "  - $MONITORING_NS"
    echo "  - $AGENTS_NS"
    echo ""
    echo "Deployed resources:"
    kubectl get all -n "$NAMESPACE" 2>/dev/null || echo "  (none in $NAMESPACE)"
    kubectl get all -n "$MONITORING_NS" 2>/dev/null || echo "  (none in $MONITORING_NS)"
    kubectl get all -n "$AGENTS_NS" 2>/dev/null || echo "  (none in $AGENTS_NS)"
    echo ""
    echo "Next steps:"
    echo "  1. Run scripts/validate.sh to verify deployment"
    echo "  2. Run scripts/test-attacks.sh to test security"
    echo "  3. Check logs: kubectl logs -n $NAMESPACE -l app=mcp-policy-proxy"
}

# Main deployment run
main() {
    check_prerequisites
    create_namespaces
    deploy_cilium
    deploy_falco
    deploy_mcp_proxy
    deploy_agent_configs
    # install_helm_chart  # Optional - comment in to enable
    wait_for_pods
    health_check
    deployment_summary

    echo ""
    echo "============================================="
    echo "  Deployment Complete!"
    echo "============================================="
}

main "$@"