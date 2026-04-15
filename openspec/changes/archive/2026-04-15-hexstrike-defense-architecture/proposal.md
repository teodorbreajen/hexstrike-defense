# Proposal: hexstrike-defense-architecture

## Intent

Implementar una arquitectura de Defense-in-Depth de 7 capas para contener y gobernar un agente autónomo de IA (hexstrike-ai) con acceso a 150+ herramientas de ciberseguridad. El objetivo es crear barreras de seguridad que inspecten, validen y contengan el comportamiento del agente en cada etapa de ejecución.

## Scope

### In Scope
- Manifiestos Kubernetes modulares por cada capa de defensa
- MCP Policy Proxy (microservicio Go) con validación semántica de comandos
- Configuración Falco con reglas específicas para detección y terminación de shells no autorizados
- Políticas Cilium con default-deny egress para restricción L7 Zero Trust
- Helm charts para componentes reutilizables
- Scripts de automatización de despliegue y validación e2e
- Documentación de arquitectura y guías de operación

### Out of Scope
- Modificación del código fuente de hexstrike-ai (solo contenemos)
- Implementación de instancias de producción reales
- Auditorías de compliance específicas (GDPR, SOC2, etc.)

## Approach

1. Crear manifiestos Kubernetes independientes por capa
2. Implementar MCP Policy Proxy como microservicio Go con validación semántica
3. Usar Helm charts para componentes reutilizables
4. Crear scripts de validación e2e automatizados
5. Documentar cada capa con justificación técnica

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `manifests/cilium/` | New | CiliumNetworkPolicy default-deny egress |
| `manifests/falco/` | New | Falco rules + Talon config para detección shells |
| `manifests/mcp-proxy/` | New | Semantic firewall proxy microservice |
| `manifests/langgraph/` | New | LangGraph agent configuration |
| `scripts/` | New | Deployment y validation scripts |
| `docs/` | New | Architecture docs y runbooks |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Complejidad operacional alta | High | Documentación exhaustiva + scripts automatizados |
| Performance overhead del proxy | Medium | Benchmarking y optimización progresiva |
| Falsos positivos en Falco | Medium | Tuning gradual con whitelist específica |
| Compatibilidad con otros CNIs | Low | Documentar requisitos de Cilium |

## Rollback Plan

- Cada componente es independiente y puede deshabilitarse individualmente
- Realizar backup de políticas existentes antes de aplicar cambios
- Feature flags en cada capa para activar/desactivar sin afectar otras
- Manifiestos versionados para rollback rápido: `kubectl apply -f manifests/v1.x/`

## Dependencies

- Kubernetes 1.28+ con soporte para eBPF
- Cilium CNI instalado y configurado
- Helm 3.x
- kubectl configurado con acceso al cluster

## Success Criteria

- [ ] Los 7 manifiestos de capa están implementados y aplicados
- [ ] MCP Policy Proxy filtra comandos maliciosos/sospechosos
- [ ] Falco detecta y termina shells no autorizados en tiempo real
- [ ] Cilium bloquea tráfico no autorizado en capa L7
- [ ] Documentación completa de arquitectura con diagramas
- [ ] Scripts de validación ejecutables y funcionales