# HashiCorp Vault F5 BIG-IP Token Plugin - Project Structure

This document outlines the structure and organization of the F5 BIG-IP Token Vault plugin project.

## Directory Structure

```
.
├── cmd/
│   └── f5-token-plugin/         # Plugin entry point
│       └── main.go              # Main executable
├── pkg/
│   └── bigiptoken/              # Plugin package
│       ├── api/                 # F5 API client 
│       │   └── client.go        # Token-based API client
│       └── backend.go           # Vault plugin backend implementation
├── Makefile                     # Build and development tasks
├── README.md                    # Project documentation
├── TOKEN_PLUGIN.md              # Token plugin usage documentation
└── go.mod                       # Go module dependencies
```

## Package Responsibilities

### cmd/f5-token-plugin

Contains the main plugin executable that Vault will execute. It's responsible for registering the plugin with Vault and starting the plugin server.

### pkg/bigiptoken

Contains the core plugin implementation:

- **backend.go**: Implements the Vault plugin backend, defining the paths and operations available in the plugin
- **api/client.go**: Custom F5 API client that handles token-based authentication and management

## Key Components

### F5 Token API Client

The `api/client.go` file implements a custom HTTP client for the F5 BIG-IP REST API, specifically focused on token-based authentication:

- **GetToken**: Authenticates to F5 BIG-IP and obtains an authentication token
- **UpdateTokenTimeout**: Updates the expiration timeout for a token
- **RevokeToken**: Revokes a token explicitly
- **ValidateToken**: Checks if a token is still valid

### Vault Plugin Backend

The `backend.go` file defines the Vault plugin structure and paths:

- **config/connection/{name}**: Configure an F5 BIG-IP connection
- **config/connections**: List all configured connections
- **token/{name}**: Generate a token for the specified connection
- **revoke/{id}**: Revoke a specific token

### Storage Structure

The plugin uses Vault's storage to maintain:

1. **Connection Configurations**: Stored at `config/connection/{name}`
   - Host information
   - Admin credentials (encrypted)
   - SSL settings

2. **Active Tokens**: Stored at `tokens/{token_id}`
   - Token value
   - Associated connection
   - Creation and expiration timestamps
   - Active status

### Periodic Functions

The plugin includes a periodic function that runs automatically to clean up expired tokens:

- **cleanupExpiredTokens**: Scans for tokens past their expiration time and revokes them

## Authentication Flow

1. Application requests a token from Vault
2. Vault calls the plugin's `pathTokenRead` function
3. Plugin retrieves the connection configuration for the specified F5 device
4. Plugin uses the API client to authenticate to F5 and obtain a token
5. Token details are stored in Vault and returned to the application
6. Application uses the token for F5 API calls
7. Token is automatically revoked when it expires

## Development Workflow

The Makefile provides commands for:

- Building the plugin: `make build-token`
- Building for Linux (Docker): `make build-token-linux`
- Starting a Vault Docker container: `make docker-dev`
- Registering the plugin: `make register-token-plugin`
- Cleaning up: `make clean` 