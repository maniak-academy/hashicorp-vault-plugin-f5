# Using the F5 BIG-IP Token Plugin with Vault UI

The F5 BIG-IP Token Plugin can be used directly through the Vault UI. This document explains how to perform the key operations through the web interface.

## Accessing the Vault UI

1. Open a web browser and navigate to `http://localhost:8200`
2. Log in with your Vault token (in dev mode, use `root`)

## Managing F5 BIG-IP Connections

### Creating a Connection

1. Browse to `f5token/config/connection/CONNECTION_NAME`
2. Click "Create" 
3. Fill in the following fields:
   - host: F5 BIG-IP hostname or IP (e.g., "172.16.10.10")
   - username: Admin username (e.g., "admin")
   - password: Admin password
   - insecure_ssl: "true" or "false"
4. Click "Save"

### Viewing Connections

1. Browse to `f5token/config/connections`
2. Click "List" to see all configured connections

### Viewing Connection Details

1. Browse to `f5token/config/connection/CONNECTION_NAME`
2. Click "Read" to view the connection details (password will be redacted)

## Managing F5 BIG-IP Tokens

### Generating a Token

1. Browse to `f5token/token/CONNECTION_NAME`
2. Click "Read"
3. Optionally specify a TTL in seconds in the "ttl" field
4. Click "Read" to generate the token
5. The result will contain:
   - token: The actual token to use with F5 BIG-IP
   - token_id: The identifier for this token in Vault
   - expires_at: When the token will expire
   - ttl: The time-to-live for the token

## Using the Token with F5 BIG-IP

After generating a token, you can use it to authenticate to the F5 BIG-IP REST API:

```
curl -k -H "X-F5-Auth-Token: YOUR_TOKEN" https://YOUR_F5_HOST/mgmt/tm/sys/version
```

The token will be automatically expired by the F5 BIG-IP system after the TTL period has passed. The plugin also maintains a record of all tokens that have been issued and will automatically clean up expired tokens.

## Verification on the F5 BIG-IP

To verify the token on the F5 BIG-IP side:

1. SSH into your F5 BIG-IP device
2. Check `/var/log/restjavad.0.log` for token creation and usage
3. Check `/var/log/secure` for authentication operations

## Important Notes

1. Tokens have a default TTL of 1 hour if not specified
2. The plugin automatically cleans up expired tokens in its internal storage
3. All token operations are securely managed within Vault
4. In production, make sure to use proper TLS settings by setting insecure_ssl to false 