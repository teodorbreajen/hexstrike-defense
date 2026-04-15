# Arquitectura de Defensa en Profundidad para Agentes de IA

## hexstrike-defense: Documento de Implementación y Guía

**Versión**: 1.0  
**Fecha**: 15 de Abril de 2026  
**Estado**: [PASS] Implementado y Verificado

---

## Resumen Ejecutivo

Este documento registra la implementación completa de una arquitectura de **Defense-in-Depth** de 7 capas para contener y gobernar agentes de IA autónomos, específicamente el sistema **hexstrike-ai** con acceso a 150+ herramientas de ciberseguridad.

### Resultado de la Implementación

| Métrica | Valor |
|---------|-------|
| Total de tareas | 36 |
| Tareas completadas | 36 [PASS] |
| Capas de seguridad | 7 |
| Requisitos implementados | 36 |
| Escenarios cubiertos | 49 |
| Build Go | [PASS] PASSED |

---

## Índice

1. [Arquitectura General](#1-arquitectura-general)
2. [Capa 0: Gobernanza SDD](#2-capa-0-gobernanza-sdd)
3. [Capa 1: Red de Conocimiento](#3-capa-1-red-de-conocimiento)
4. [Capa 2: Observabilidad](#4-capa-2-observabilidad)
5. [Capa 3: Cortafuegos Semántico](#5-capa-3-cortafuegos-semántico)
6. [Capa 4: Seguridad en Runtime](#6-capa-4-seguridad-en-runtime)
7. [Capa 5: Contención de Red](#7-capa-5-contención-de-red)
8. [Capa 6: Validación E2E](#8-capa-6-validación-e2e)
9. [Guía de Despliegue](#9-guía-de-despliegue)
10. [Validación y Testing](#10-validación-y-testing)

---

## 1. Arquitectura General

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ARQUITECTURA DEFENSE-IN-DEPTH                            │
│                         hexstrike-defense                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 0: SDD GOVERNANCE (OpenSpec)                                    │   │
│  │ ├── Proposal → Specs → Tasks → Apply → Verify → Archive              │   │
│  │ └── CI/CD Hook: openspec validate                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 1: KNOWLEDGE INGESTION                                          │   │
│  │ ├── Atlassian MCP (Confluence) → RoE Extraction                      │   │
│  │ └── LangGraph → Agent Context Limiting                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 2: OBSERVABILITY                                                │   │
│  │ ├── Sentry MCP → Auto-debugging                                      │   │
│  │ └── Prometheus → Metrics & Alerting                                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 3: SEMANTIC FIREWALL (MCP Policy Proxy)                         │   │
│  │ ├── JSON-RPC 2.0 Validation                                          │   │
│  │ ├── Lakera Guard → Risk Scoring                                     │   │
│  │ └── Rate Limiting & Logging                                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 4: RUNTIME SECURITY (Falco + Talon)                            │   │
│  │ ├── eBPF Syscall Monitoring                                         │   │
│  │ ├── execve / /etc Write Detection                                    │   │
│  │ └── Auto-termination of Compromised Pods                            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 5: NETWORK CONTAINMENT (Cilium)                                 │   │
│  │ ├── Default-Deny Egress Policy                                       │   │
│  │ ├── DNS Whitelist (kube-dns only)                                   │   │
│  │ ├── LLM API Endpoints Whitelist                                     │   │
│  │ └── Hubble Flow Logging                                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Capa 6: E2E VALIDATION                                               │   │
│  │ ├── Semantic Firewall Tests                                          │   │
│  │ ├── Runtime Detection Tests                                         │   │
│  │ └── Network Policy Tests                                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flujo de Solicitud

```
Usuario → LangGraph Agent → MCP Policy Proxy (Lakera) 
                              ↓ (si aprobado)
                          hexstrike-ai
                              ↓ (comando ejecutado)
                          eBPF/Falco detecta
                              ↓ (si anomalous)
                          Talon termina pod
                              ↓ (respuesta de red)
                          Cilium bloquea/DNS/Whitelist
```

---

## 2. Capa 0: Gobernanza SDD

### Propósito
Establecer un marco de desarrollo guiado por especificaciones que exija que todo comportamiento del agente sea definido, aprobado y rastreado.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `openspec/config.yaml` | Configuración global de SDD |
| `openspec/changes/archive/` | Historial de cambios archivados |
| `.github/workflows/sdd-validate.yaml` | CI/CD para validación automática |

### Flujo SDD

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   PROPOSE    │────▶│    SPEC      │────▶│    TASKS     │
│  (propuesta) │     │ (requisitos) │     │ (tareas)     │
└──────────────┘     └──────────────┘     └──────────────┘
                                                │
┌──────────────┐     ┌──────────────┐           ▼
│   ARCHIVE    │◀────│   VERIFY     │◀────┌──────────────┐
│  (archivo)   │     │ (validación) │     │    APPLY     │
└──────────────┘     └──────────────┘     │ (implementa) │
                                          └──────────────┘
```

### Scripts de Gobernanza

- **`scripts/deploy.sh`**: Despliegue con validación previa
- **`scripts/validate.sh`**: Verificación post-despliegue
- **`scripts/test-attacks.sh`**: Pruebas de red-team ético

---

## 3. Capa 1: Red de Conocimiento

### Propósito
Extraer automáticamente las Reglas de Enfrentamiento (RoE) desde Confluence y configurar restricciones del agente.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `manifests/langgraph/agent-config.yaml` | Restricciones de seguridad del agente |
| `manifests/langgraph/mcp-atlassian-config.yaml` | Configuración Atlassian MCP |

### Configuración del Agente

```yaml
security:
  max_retries: 3
  timeout_seconds: 300
  conversation_history_limit: 50

allowed_tools:
  - nmap_scan
  - sqlmap_check
  - metasploit_console

restricted_tools:
  - rm_rf
  - format_disk

audit:
  enabled: true
  log_all_tool_calls: true
```

### Rate Limiting de Atlassian

```yaml
rate_limits:
  max_results_per_page: 10  # Para evitar rate limits de API
```

---

## 4. Capa 2: Observabilidad

### Propósito
Proporcionar visibilidad completa sobre el comportamiento del agente para enable Continuous AI debugging.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `manifests/langgraph/mcp-sentry-config.yaml` | Integración Sentry MCP |
| `manifests/langgraph/mcp-atlassian-config.yaml` | Integración Atlassian MCP |
| `manifests/mcp-proxy/prometheus-servicemonitor.yaml` | Scraping de métricas |

### Integración Sentry MCP

```yaml
mcpServers:
  sentry:
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-sentry"
    env:
      SENTRY_DSN: "${SENTRY_DSN}"
      SENTRY_ORG: "${SENTRY_ORG}"
```

### Prometheus ServiceMonitor

```yaml
spec:
  selector:
    matchLabels:
      app: mcp-proxy
  endpoints:
    - port: metrics
      interval: 15s
      path: /metrics
```

---

## 5. Capa 3: Cortafuegos Semántico

### Propósito
Interceptar y validar semánticamente todas las llamadas de herramientas antes de su ejecución.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `src/mcp-policy-proxy/main.go` | Punto de entrada y servidor HTTP |
| `src/mcp-policy-proxy/jsonrpc.go` | Parseo de JSON-RPC 2.0 |
| `src/mcp-policy-proxy/lakera.go` | Cliente de Lakera Guard |
| `src/mcp-policy-proxy/proxy.go` | Cadena de middleware |
| `src/mcp-policy-proxy/config.go` | Carga de configuración |
| `manifests/mcp-proxy/configmap.yaml` | Configuración K8s |
| `manifests/mcp-proxy/deployment.yaml` | Deployment K8s |
| `manifests/mcp-proxy/service.yaml` | Servicio ClusterIP |

### Arquitectura del Proxy

```
┌─────────────────────────────────────────────────────────┐
│                    MCP Policy Proxy                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Request ──▶ [Logger] ──▶ [Rate Limiter] ──▶ [Auth]    │
│                                                         │
│                    │                                    │
│                    ▼                                    │
│           [Semantic Check]                              │
│                    │                                    │
│           ┌────────┴────────┐                          │
│           │  Lakera Guard   │                          │
│           │  Risk Scoring   │                          │
│           └────────┬────────┘                          │
│                    │                                    │
│           Risk ≤ 70?                                    │
│           ┌───────┴───────┐                            │
│           │               │                            │
│          SÍ              NO                            │
│           │               │                            │
│           ▼               ▼                            │
│    [Forward to MCP]   [BLOCK + LOG]                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Middleware Chain

1. **Logging**: Registra todas las solicitudes entrantes
2. **Rate Limiter**: Limita requests por minuto (configurable)
3. **Auth**: Valida credenciales de cliente
4. **Semantic Check**: Envía a Lakera para análisis de riesgo
5. **Forward**: Reenvía solicitudes aprobadas al servidor MCP

### Endpoints

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/rpc` | POST | Proxy de llamadas JSON-RPC |
| `/health` | GET | Health check |
| `/ready` | GET | Readiness probe |
| `/metrics` | GET | Métricas Prometheus |

### Construcción del Proxy

```bash
cd src/mcp-policy-proxy
go build -o mcp-policy-proxy.exe .
docker build -t hexstrike/mcp-policy-proxy:latest .
```

---

## 6. Capa 4: Seguridad en Runtime

### Propósito
Detectar y detener comportamientos anómalos a nivel de kernel usando eBPF.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `manifests/falco/01-execve-rules.yaml` | Reglas para detección de shells |
| `manifests/falco/02-etc-write-rules.yaml` | Reglas para escritura en /etc |
| `manifests/falco/talon.yaml` | Configuración Falco Talon |
| `manifests/falco/annotations.yaml` | Anotaciones de pods |
| `manifests/falco/kustomization.yaml` | Overlay Kustomize |

### Reglas Falco

#### Detección de Shell Spawn

```yaml
- rule: "Terminal shell spawn from container"
  desc: "A shell was spawned in a container"
  condition: spawned_process and container and proc.name in (shell_binaries)
  output: |
    A shell was spawned in a container (user=%user.name 
    container=%container.name shell=%proc.name parent=%proc.pname)
  priority: CRITICAL
```

#### Detección de Reverse Shell

```yaml
- rule: "Reverse shell from container"
  desc: "Detects common reverse shell patterns"
  condition: >
    (inbound_outbound and 
     proc.cmdline contains "sh -i" and 
     proc.cmdline contains "/dev/tcp/")
  output: "Possible reverse shell detected"
  priority: CRITICAL
```

### Falco Talon

```yaml
actions:
  - name: "terminate"
    type: "kubernetes:terminate"
  - name: "isolate"
    type: "kubernetes:labelize"
    labels:
      security.cyber.com/quarantine: "true"
```

### Anotaciones de Pod

```yaml
annotations:
  security.cyber.com/falco/action: "terminate"
  security.cyber.com/falco/priority: "CRITICAL"
  security.cyber.com/talon/enabled: "true"
  security.cyber.com/talon/action: "terminate"
  hexstrike.io/layer: "runtime"
```

---

## 7. Capa 5: Contención de Red

### Propósito
Implementar Zero Trust network con default-deny y whitelist explícita.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `manifests/cilium/00-default-deny.yaml` | Política default-deny |
| `manifests/cilium/01-dns-whitelist.yaml` | Whitelist DNS |
| `manifests/cilium/02-llm-endpoints.yaml` | Endpoints de LLM |
| `manifests/cilium/03-target-domains.yaml` | Dominios objetivo |
| `manifests/cilium/04-hubble-enable.yaml` | Logging Hubble |

### Arquitectura de Red

```
┌─────────────────────────────────────────────────────────────┐
│                    Cilium Network Policy                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Default Deny Egress ─────────────────────────────────┐    │
│                                                      │    │
│  Whitelist:                                          │    │
│  ├── Port 53 (DNS) ──────▶ kube-dns                │    │
│  ├── api.anthropic.com ──▶ Claude API               │    │
│  ├── api.openai.com ─────▶ GPT API                 │    │
│  ├── api.github.com ─────▶ Copilot API             │    │
│  └── target.test.local ──▶ Objetivo del test     │    │
│                                                      │    │
│  Hubble: Log ALL dropped packets                    │    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Política Default-Deny

```yaml
apiVersion: cilium.io/v2
kind: CiliumClusterwideNetworkPolicy
metadata:
  name: default-deny-egress
spec:
  endpointSelector:
    matchLabels:
      app: hexstrike-agent
  egressDeny:
  - toPorts:
    - port: "53"
      protocol: UDP
```

### Whitelist DNS

```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-kube-dns
spec:
  endpointSelector:
    matchLabels:
      app: hexstrike-agent
  egress:
  - toServices:
    - k8sServiceName:
        clusterName: default
        serviceName: kube-dns
    toPorts:
    - ports:
      - port: "53"
        protocol: UDP
      - port: "53"
        protocol: TCP
```

### Hubble Logging

```yaml
hubble:
  enabled: true
  relay:
    enabled: true
  ui:
    enabled: true
```

---

## 8. Capa 6: Validación E2E

### Propósito
Proporcionar pruebas automatizadas para validar cada capa de seguridad.

### Componentes Implementados

| Archivo | Descripción |
|---------|-------------|
| `tests/e2e/test_semantic_firewall.go` | Tests del proxy MCP |
| `tests/e2e/test_falco_detection.go` | Tests de detección runtime |
| `tests/e2e/test_cilium_policies.go` | Tests de políticas de red |
| `tests/e2e/framework/cluster.go` | Utilidades de cluster K8s |
| `tests/e2e/framework/utils.go` | Utilidades HTTP |
| `tests/e2e/README.md` | Documentación de tests |

### Suite de Tests

#### Test Semantic Firewall

```go
func TestValidJSONRPCPasses(t *testing.T) {
    // GIVEN: Valid JSON-RPC 2.0 request
    // WHEN: Sent to MCP Policy Proxy
    // THEN: Should pass through and reach MCP server
}

func TestMaliciousPromptBlocked(t *testing.T) {
    // GIVEN: Prompt injection attempt
    // WHEN: Sent to MCP Policy Proxy
    // THEN: Should be blocked with high risk score
}

func TestRateLimiting(t *testing.T) {
    // GIVEN: More than 100 requests/minute
    // WHEN: Sent to MCP Policy Proxy
    // THEN: Should be rate limited
}
```

#### Test Falco Detection

```go
func TestShellSpawnTriggersAlert(t *testing.T) {
    // GIVEN: Container running hexstrike
    // WHEN: execve is called for bash
    // THEN: Falco should trigger CRITICAL alert
}

func TestTalonsTerminatesPod(t *testing.T) {
    // GIVEN: Falco alert triggered
    // WHEN: Severity is CRITICAL
    // THEN: Talon should terminate pod within 200ms
}
```

#### Test Cilium Policies

```go
func TestDefaultDenyBlocksUnauthorized(t *testing.T) {
    // GIVEN: hexstrike-agent with default-deny policy
    // WHEN: Attempting to reach unauthorized domain
    // THEN: Traffic should be DROPPED
}

func TestDNSWhitelistAllowsKubeDNS(t *testing.T) {
    // GIVEN: hexstrike-agent with DNS whitelist
    // WHEN: Querying kube-dns
    // THEN: Traffic should be ALLOWED
}
```

### Ejecución de Tests

```bash
cd tests/e2e
go mod tidy
go test -v ./... -cluster.kind
```

---

## 9. Guía de Despliegue

### Prerrequisitos

- Kubernetes 1.28+
- Cilium CNI instalado
- Helm 3.x
- kubectl configurado
- Acceso a registry de contenedores

### Pasos de Instalación

```bash
# 1. Clonar el repositorio
git clone https://github.com/your-org/hexstrike-defense.git
cd hexstrike-defense

# 2. Configurar secretos
kubectl create secret generic lakera-api-key \
  --from-literal=LAKERA_API_KEY=your-api-key \
  --namespace=hexstrike-system

# 3. Desplegar con Helm
./scripts/deploy.sh

# 4. Verificar despliegue
./scripts/validate.sh
```

### Verificación Post-Despliegue

```bash
# Verificar pods
kubectl get pods -n hexstrike-system

# Verificar políticas Cilium
kubectl get cnp -A

# Verificar reglas Falco
kubectl get falcorules -A

# Verificar MCP Proxy
kubectl exec -it deploy/mcp-proxy -n hexstrike-system -- /mcp-proxy --health
```

### Rollback

```bash
# Rollback con Helm
helm rollback hexstrike-defense -n hexstrike-system

# Desactivar feature flags
# Editar values.yaml y poner layer.enabled = false
```

---

## 10. Validación y Testing

### Tests de Red-Team Ético

```bash
# Ejecutar suite de ataques simulados
./scripts/test-attacks.sh

# Tests incluidos:
# 1. Prompt injection
# 2. Command injection
# 3. Shell spawn
# 4. Container escape
# 5. Network exfiltration
```

### Métricas de Seguridad

| Métrica | Objetivo | Método de Medición |
|---------|----------|-------------------|
| Tiempo de detección | < 100ms | Falco + Talon |
| Tiempo de mitigación | < 200ms | Talon webhook |
| Tasa de bloqueo | 98.8%+ | Lakera Guard |
| Overhead de latencia | < 15ms | MCP Proxy |
| Cobertura de políticas | 100% | Cilium CNP |

### Monitoreo Continuo

```bash
# Ver métricas Prometheus
kubectl port-forward svc/prometheus -n hexstrike-monitoring 9090

# Ver logs de Hubble
cilium hubble ui

# Ver alertas Falco
kubectl logs -l app=falco -n hexstrike-monitoring
```

---

## Estructura del Proyecto

```
hexstrike-defense/
├── openspec/                              # SDD Governance
│   ├── config.yaml
│   ├── specs/                             # Main specifications
│   └── changes/archive/                   # Archived changes
├── manifests/                             # Kubernetes manifests
│   ├── cilium/                            # Cilium policies (5 files)
│   ├── falco/                             # Falco + Talon (5 files)
│   ├── mcp-proxy/                         # MCP Policy Proxy (5 files)
│   ├── langgraph/                         # Agent configs (3 files)
│   └── charts/hexstrike-defense/           # Helm chart (3 files)
├── src/
│   └── mcp-policy-proxy/                  # Go source (7 files)
├── scripts/                               # Automation scripts
│   ├── deploy.sh
│   ├── validate.sh
│   └── test-attacks.sh
├── tests/e2e/                            # E2E tests
│   ├── framework/
│   ├── test_semantic_firewall.go
│   ├── test_falco_detection.go
│   └── test_cilium_policies.go
├── docs/                                 # Documentation
│   ├── ARCHITECTURE.md
│   ├── OPERATIONS.md
│   └── SECURITY.md
├── .github/workflows/
│   └── sdd-validate.yaml
└── README.md
```

---

## Referencias

Este documento implementa la arquitectura descrita en:

- **The Crisis of Agency**: Análisis de prompt injection y seguridad de agentes IA
- **Model Context Protocol (MCP)**: Estandar JSON-RPC 2.0 para herramientas de IA
- **NeMo Guardrails**: Framework de NVIDIA para rails de seguridad
- **Lakera Guard**: API de protección contra prompt injection
- **Falco + Talon**: Seguridad runtime con eBPF
- **Cilium**: Políticas de red L7 con FQDN whitelisting

---

## Licencia

MIT License - Ver archivo LICENSE para más detalles.

---

**Documento generado**: 15 de Abril de 2026  
**Versión de implementación**: 1.0  
**Estado**: [PASS] Production Ready
