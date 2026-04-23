```mermaid
graph TD
    subgraph "Security Layers"
        L1[{"layer": "1", "name": "Infrastructure"}]
        L2[{"layer": "2", "name": "Isolation"}]
        L3[{"layer": "3", "name": "Network"}]
        L4[{"layer": "4", "name": "Runtime"}]
        L5[{"layer": "5", "name": "Semantic"}]
        L6[{"layer": "6", "name": "Observability"}]
    end

    Attack --> L1
    Attack --> L2
    Attack --> L3
    Attack --> L4
    Attack --> L5
    Attack --> L6

    L1 -->|"blocked"| Blocked
    L2 -->|"blocked"| Blocked
    L3 -->|"blocked"| Blocked
    L4 -->|"detected"| Blocked
    L5 -->|"blocked"| Blocked
    L6 -->|"alerted"| Monitored

    style Attack fill:#f00,color:#fff
    style Blocked fill:#f00,color:#fff
    style Monitored fill:#ff0
```