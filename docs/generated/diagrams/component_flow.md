```mermaid
flowchart LR
    subgraph Proxy
        P[MCP Policy Proxy]
    end

    subgraph Backend
        MCP[MCP Server]
        LK[Lakera Guard]
    end

    Client --> P
    P --> MCP
    P --> LK

```