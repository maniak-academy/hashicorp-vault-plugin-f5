# F5 BIG-IP Token Plugin Flow Diagrams

This document contains visual diagrams of the F5 BIG-IP Token Plugin workflows.

## Token Generation and Usage Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Vault as HashiCorp Vault
    participant Plugin as F5 Token Plugin
    participant F5 as F5 BIG-IP
    
    App->>Vault: Authenticate (get Vault token)
    Vault-->>App: Vault token
    
    App->>Vault: Request F5 token (f5token/token/bigip1)
    Vault->>Plugin: Forward request
    
    Plugin->>F5: Authenticate with stored credentials
    F5-->>Plugin: Generate token
    
    Plugin->>Plugin: Store token details in Vault
    Plugin-->>Vault: Return token details
    Vault-->>App: F5 token + metadata
    
    App->>F5: API calls with F5 token
    F5-->>App: API responses
    
    Note over Plugin,F5: Token automatically expires after TTL
    Plugin->>Plugin: Periodic cleanup of expired tokens
```

## Plugin Architecture

```mermaid
flowchart TB
    subgraph Vault["HashiCorp Vault"]
        direction TB
        API[Vault API] --> Plugin
        
        subgraph Plugin["F5 Token Plugin"]
            direction TB
            ConfigPath["Configuration Paths\n/config/connection/*"] --> Backend
            TokenPath["Token Paths\n/token/*"] --> Backend
            TokensList["Tokens List Path\n/tokens"] --> Backend
            
            Backend["Plugin Backend"] --> Storage
            Backend --> F5Client
        end
    end
    
    Storage[(Vault Storage)]
    
    subgraph F5["F5 BIG-IP Devices"]
        F5_1["Device 1\n172.16.10.10"]
        F5_2["Device 2\n172.16.10.11"]
        F5_N["Device N"]
    end
    
    F5Client["F5 API Client"] -- "REST API\nAuthN Tokens" --> F5
    
    App1["Application 1"] -- "1. Get token" --> API
    App1 -- "2. Use token" --> F5
    
    App2["Application 2"] -- "1. Get token" --> API
    App2 -- "2. Use token" --> F5
```

## Token Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Requested: Application requests token
    
    Requested --> Generated: Plugin authenticates to F5
    Generated --> Active: Token stored in Vault
    
    Active --> Used: Application uses token with F5
    Used --> Active: Token still valid
    
    Active --> Expired: TTL reached
    Expired --> Cleaned: Plugin periodic function
    
    Active --> Revoked: Manual revocation
    Revoked --> Cleaned
    
    Cleaned --> [*]
``` 