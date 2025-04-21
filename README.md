# Aura
Aura is a Solana based onchain data retrieval tool that provides traders with real-time data as well as historic metadata to make data driven decisions when executing trades.

Onchain data is extracted via mainnet RPC node calls, decoding Metaplex PDA data, as well as more optimized RPC calls using Helius.

## High Level Architecture
```mermaid
flowchart TD
    subgraph UI[Frontend]
        WEB[Telegram Interface]
    end
    
    subgraph GCP[GCP Compute Engine VM]
        subgraph Backend[Backend Service]
            subgraph DataProcessing[Data Processing]
                COLLECT[Data Collection] --> PROCESS[Data Processing]
                PROCESS --> |Persist Token/User Data| STORE_DB[Load into DB]
                PROCESS --> |Real-time Wallet Updates| SERVE[Load into API]
            end
            
            subgraph HTTP_SERVER[HTTP Server]
                API[REST Endpoints]
            end
        end
        
        subgraph Docker[Docker Container]
            subgraph Database[Database]
                DB[(Postgres)]
            end
        end
    end
    
    subgraph Blockchain[Solana Network]
        RPC[RPC Node]
        BC_WS[WebSocket Connection]
    end
    
    %% Data Flows
    WEB <--> |HTTP Requests/Responses| API
    API <--> |Query/Store Data| DB
    
    COLLECT --> |Request Based| RPC
    COLLECT --> |Subscription Based Real-time Stream| BC_WS
    STORE_DB --> |Persist Token/User Data| DB
    SERVE --> |Real-time Wallet Updates| API
    
    %% Styling
    classDef frontend fill:#E8EAED,stroke:#BDC1C6,stroke-width:1px,color:#202124
    classDef backend fill:#DAE8FC,stroke:#6C8EBF,stroke-width:1px,color:#333
    classDef blockchain fill:#F5F5F5,stroke:#CCCCCC,stroke-width:1px,color:#333
    classDef database fill:#E1D5E7,stroke:#9673A6,stroke-width:1px,color:#333
    classDef dataProcessingService fill:#f8d894,stroke:#D6B656,stroke-width:1px,color:#333
    classDef processSteps fill:#F8CECC,stroke:#B85450,stroke-width:1px,color:#333
    classDef server fill:#D5E8D4,stroke:#82B366,stroke-width:1px,color:#333
    classDef gcp fill:#F9F9F9,stroke:#999999,stroke-width:1px,color:#333
    classDef docker fill:#F0F8FF,stroke:#1D70B8,stroke-width:1px,color:#333
    
    class WEB frontend
    class API,Backend backend
    class HTTP_SERVER server
    class DataProcessing dataProcessingService
    class COLLECT,PROCESS,STORE_DB,SERVE processSteps
    class RPC,BC_WS blockchain
    class DB database
    class GCP gcp
    class Docker docker
    
    %% Make arrow lines darker
    linkStyle default stroke:#555555,stroke-width:1.5px
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
