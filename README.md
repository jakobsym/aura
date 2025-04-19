# Aura
Aura is a Solana based onchain data retrieval tool that provides traders with real-time data as well as historic metadata to make data driven decisions when executing trades.

Onchain data is extracted via mainnet RPC node calls, decoding Metaplex PDA data, as well as more optimized RPC calls using Helius.

## Aura Architecture
```mermaid
flowchart TD
    subgraph UI[Frontend]
        WEB[Telegram Interface]
    end

    subgraph Backend[Backend Service]
        subgraph ETL[Pseudo Pipeline]
            EXTRACT[Extract] --> TRANSFORM[Transform]
            TRANSFORM --> |Load to DB| LOAD_DB[Load]
            TRANSFORM --> |Load to HTTP Server| LOAD_HTTP[Load]
        end
        
        subgraph HTTP_SERVER[HTTP Server]
            API[REST Endpoints]
        end
    end

    subgraph Database[Database]
        DB[(Postgres)]
    end

    subgraph Blockchain[Solana Network]
        RPC[RPC Node]
        BC_WS[WebSocket Connection]
    end

    %% Data Flows
    WEB <--> |HTTP Requests/Responses| API
    API <--> |Query Data| DB
    
    EXTRACT --> RPC
    EXTRACT --> |Real-time Stream| BC_WS
    LOAD_DB --> |Persist Data| DB
    LOAD_HTTP --> |Real-time Updates| API

    %% Styling
    classDef frontend fill:#7CB9E8,stroke:#4682B4,stroke-width:2px,color:white
    classDef backend fill:#98FB98,stroke:#3CB371,stroke-width:2px,color:#333
    classDef blockchain fill:#FFB6C1,stroke:#DB7093,stroke-width:2px,color:#333
    classDef database fill:#FFA07A,stroke:#FF8C00,stroke-width:1px,color:#333
    classDef etl fill:#DDA0DD,stroke:#9400D3,stroke-width:2px,color:#333
    classDef server fill:#B0C4DE,stroke:#4169E1,stroke-width:4px,color:#333

    class WEB frontend
    class API backend
    class HTTP_SERVER server
    class EXTRACT,TRANSFORM,LOAD_DB,LOAD_HTTP etl
    class RPC,BC_WS blockchain
    class DB database
```
## Features
- 🔎 Real-time wallet monitoring
- 📊 SPL token and wallet metadata retrieval
- 🖥️ Telegram interface for mobile or desktop based access

## Requirements
- `Go: 1.23.X +`
- [Helius API key](https://dashboard.helius.dev/)

## Quick Start (Locally)
- Below are instructions to set up a local copy of the Go backend server. You’ll need to configure a data store (e.g., PostgreSQL, Redis, or an in-memory solution). 

- The project uses a *Repository Pattern*, so the data store implementation can be swapped without modifying core logic—just replace the repository layer with your preferred storage solution.
```
$ git clone https://github.com/jakobsym/aura.git
$ make
$ ./bin/aura
```

## Usage Example(s)
- Locally you can access specific endpoints of the internal API
    


<user_id> wishes to track <solana_wallet_address>
```
$ curl -X POST localhost:3000/v0/track/<solana_wallet_address> \
    -H "Content-Type: application/json" \
    -d '{
        "user_id" : <user_id>
    }'
```

Receive metadata for <token_address>
```
$ curl -X GET localhost:3000/v0/token/<token_address>

```
Response:
```
{
  "token_address": <token_address>,
  "name": "Solana",
  "symbol": "SOL",
  "created_at": "2024-05-05T06:18:01Z",
  "supply": 926910034.835728,
  "price": 0.00147629,
  "fdv": 1368388.0153276369,
  "socials": "https://x.com/search?q=6yjNqPzTSanBWSa6dxVEgTjePXBrZ2FoHLDQwYwEsyM6"
}
```
