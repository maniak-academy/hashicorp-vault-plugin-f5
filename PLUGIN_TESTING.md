# HashiCorp Vault F5 BIG-IP Token Plugin Testing

This document describes the successful testing of the HashiCorp Vault F5 BIG-IP Token Plugin.

## Plugin Functionality Tested

The following functionality has been successfully tested:

1. **Connection Configuration**: Storing F5 BIG-IP connection details securely in Vault
2. **Connection Reading**: Retrieving connection details (with password redacted)
3. **Connection Listing**: Listing all configured F5 BIG-IP connections
4. **Token Generation**: Creating authentication tokens for F5 BIG-IP with configurable TTLs
5. **Token Management**: Listing, viewing details, and revoking tokens
6. **Token Revocation**: Manually revoking tokens before they expire

## Test Environment

- Vault running in development mode in Docker container
- F5 BIG-IP configuration:
  - Host: 172.16.10.10
  - Credentials: admin/W3lcome098!
  - Insecure SSL: enabled for testing purposes

## Using the Plugin

### Configuration

Set up a connection to an F5 BIG-IP device:

```bash
vault write f5token/config/connection/CONNECTION_NAME \
  host="F5_HOST_OR_IP" \
  username="USERNAME" \
  password="PASSWORD" \
  insecure_ssl=BOOLEAN
```

### Token Generation

Generate an authentication token:

```bash
# With default TTL (1 hour)
vault read f5token/token/CONNECTION_NAME

# With custom TTL (in seconds)
vault read f5token/token/CONNECTION_NAME ttl=300
```

### Token Management

List all active tokens:

```bash
vault read f5token/tokens
```

View details for a specific token:

```bash
vault read f5token/token-info/TOKEN_ID
```

### Token Revocation

Revoke a token:

```bash
vault write -force f5token/revoke/TOKEN_ID
```

### Connection Management

List all connections:

```bash
vault list f5token/config/connections
```

View connection details:

```bash
vault read f5token/config/connection/CONNECTION_NAME
```

Delete a connection:

```bash
vault delete f5token/config/connection/CONNECTION_NAME
```

## Using the Plugin from the Vault UI

The F5 BIG-IP Token Plugin is fully compatible with the Vault UI. You can:

1. **Configure Connections**: Navigate to the `f5token/config/connection/` path and create a new connection by selecting "Create new version" and entering your connection details.

2. **Generate Tokens**: Go to the `f5token/token/CONNECTION_NAME` path and select "Read" to generate a token. You can specify the TTL in the "ttl" field.

3. **List Tokens**: View all active tokens by browsing to the `f5token/tokens` path and clicking "Read".

4. **View Token Details**: Get detailed information about a specific token by navigating to `f5token/token-info/TOKEN_ID` and clicking "Read".

5. **Revoke Tokens**: Revoke a token by going to `f5token/revoke/TOKEN_ID` and clicking "Create".

## Test Scripts

The repository includes two scripts to help demonstrate the plugin functionality:

1. `configure_connection.sh`: Sets up a connection to an F5 BIG-IP device
2. `test_plugin.sh`: Performs a complete test of all plugin functionality, including token management

## Notes

- Tokens have a default TTL of 1 hour if not specified
- The plugin automatically revokes expired tokens through a periodic function
- All operations are logged on the F5 BIG-IP device for audit purposes
- The token plugin is more efficient than creating temporary users for authentication
- The plugin securely stores tokens in Vault's storage system, allowing for proper management and tracking of tokens 