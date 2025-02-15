# aura Architecture

This project implements a real-time Solana transaction monitoring system with the following key components:

```mermaid
flowchart TD
    subgraph UI[Frontend]
        WEB[Web Interface]
        WS_CLIENT[WebSocket Client]
    end

    subgraph Backend[Backend Service]
        subgraph HTTP_SERVER[HTTP Server]
            API[REST Endpoints]
            WS[WebSocket Upgrade]
        end
        MONITOR[Transaction Monitor]
    end

    subgraph Database[Database]
        DB[(Postgres)]
    end

    subgraph Blockchain[Solana Network]
        RPC[RPC Node]
        BC_WS[Blockchain WebSocket]
    end

    %% Bi-directional Flows
    WEB  |HTTP Req/Res| API
    WS_CLIENT  |WebSocket Messages| WS
    API  DB
    RPC --> |Query Data| API
    WS --> MONITOR
    MONITOR  BC_WS
    MONITOR --> DB

    %% Modern, softer color palette
    classDef frontend fill:#7CB9E8,stroke:#4682B4,stroke-width:2px,color:white
    classDef backend fill:#98FB98,stroke:#3CB371,stroke-width:2px,color:#333
    classDef blockchain fill:#FFB6C1,stroke:#DB7093,stroke-width:2px,color:#333
    classDef database fill:#FFA07A,stroke:#FF8C00,stroke-width:2px,color:#333
    classDef server fill:#B0C4DE,stroke:#4169E1,stroke-width:4px,color:#333

    class WEB,WS_CLIENT frontend
    class API,WS,MONITOR backend
    class HTTP_SERVER server
    class RPC,BC_WS blockchain
    class DB database
```