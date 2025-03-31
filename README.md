# HashiCorp Vault F5 BIG-IP Token Plugin

A HashiCorp Vault plugin for managing authentication tokens for F5 BIG-IP devices.

## Overview

This plugin enables HashiCorp Vault to:
- Securely store F5 BIG-IP admin credentials
- Generate REST API tokens for F5 BIG-IP authentication
- Manage token lifecycles with configurable TTLs
- Support multiple F5 BIG-IP devices from a single Vault instance
- List and track all active tokens

Instead of creating temporary user accounts on F5 BIG-IP systems, this plugin obtains and manages authentication tokens. Applications like Ansible, CI/CD pipelines, or other automation tools can request tokens from Vault when they need to interact with F5 devices.

## Features

- **Secure Credential Storage**: Admin credentials are securely stored in Vault
- **Token Generation**: Create authentication tokens with configurable TTLs
- **Multiple Connections**: Manage tokens for multiple F5 BIG-IP devices
- **Auto-Cleanup**: Expired tokens are automatically cleaned up
- **Token Tracking**: List all active tokens across connections
- **UI Support**: Generate tokens via Vault UI or CLI
- **TTL Management**: Configure token lifetimes based on your security requirements

## Why Use API Tokens?

Using API tokens instead of temporary user accounts provides several advantages:

1. **No User Account Management**: No need to create and delete user accounts
2. **Lower Overhead**: Faster authentication with fewer resources
3. **Fine-grained Control**: Precise token expiration timing
4. **Native Support**: Uses the built-in F5 BIG-IP token API
5. **Improved Auditing**: Better tracking of who is accessing what

## Installation

### Prerequisites

- HashiCorp Vault (v1.4.0+)
- Go 1.19+ (for building)
- F5 BIG-IP device(s) with API access

### Building the Plugin

```bash
# Build for local platform
make build-token

# Build for Linux (required for Docker)
make build-token-linux
```

### Registering the Plugin with Vault

```bash
# Using the make target (recommended)
make register-token-plugin

# Or manually
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
SHA256=$(shasum -a 256 vault-plugin-f5-token-linux | cut -d ' ' -f1)
vault plugin register -sha256=$SHA256 secret vault-plugin-f5-token-linux
vault secrets enable -path=f5token vault-plugin-f5-token-linux
```

## Usage

### Command Line Interface

#### Configure a Connection to F5 BIG-IP

```bash
vault write f5token/config/connection/bigip1 \
  host="172.16.10.10" \
  username="admin" \
  password="password" \
  insecure_ssl=true
```

#### List All Configured Connections

```bash
vault list f5token/config/connections
```

#### Generate an API Token

```bash
# With default TTL (1 hour)
vault read f5token/token/bigip1

# With custom TTL (in seconds)
vault read f5token/token/bigip1 ttl=300
```

Example output:
```
Key           Value
---           -----
expires_at    2025-03-31T18:57:01Z
host          bigip1
token         2CLQGKIQGBH42P7LEVYA7YO3NO
token_id      token_bigip1_1743443821
ttl           1h
```

#### List All Active Tokens

```bash
vault read f5token/tokens
```

Example output:
```
Key      Value
---      -----
tokens   [map[active:true created_at:2025-03-31T18:52:06Z expires_at:2025-03-31T19:52:06Z host:bigip1 token_id:token_bigip1_1743443821]]
```

This shows all currently active tokens across all F5 BIG-IP connections, including:
- Which connection the token is for (`host`)
- When the token was created (`created_at`)
- When the token will expire (`expires_at`)
- The token's identifier (`token_id`)
- Whether the token is active (`active`)

### Vault UI

You can also manage connections and generate tokens through the Vault UI:

1. **Configure Connections**:
   - Navigate to `f5token/config/connection/YOUR_CONNECTION_NAME`
   - Click "Create"
   - Fill in the connection details and click "Save"

2. **Generate Tokens**:
   - Navigate to `f5token/token/YOUR_CONNECTION_NAME`
   - Optionally set a custom TTL
   - Click "Read" to generate a token

3. **View All Active Tokens**:
   - Navigate to `f5token/tokens`
   - Click "Read" to see all active tokens

### Using with Applications

Applications can request tokens from Vault when needed:

1. Authenticate to Vault
2. Request a token for the F5 BIG-IP
3. Use the token for all F5 BIG-IP API calls
4. Let the token expire automatically

#### Example with Ansible

```yaml
- name: Get F5 token from Vault
  uri:
    url: "{{ vault_addr }}/v1/f5token/token/bigip1"
    method: GET
    headers:
      X-Vault-Token: "{{ vault_token }}"
    status_code: 200
    validate_certs: no
  register: f5_token_response

- name: Set F5 token variables
  set_fact:
    f5_token: "{{ f5_token_response.json.data.token }}"
    f5_host: "{{ f5_token_response.json.data.host }}"

- name: Use F5 token in an API call
  uri:
    url: "https://{{ f5_bigip_address }}/mgmt/tm/sys/version"
    method: GET
    headers:
      X-F5-Auth-Token: "{{ f5_token }}"
    status_code: 200
    validate_certs: no
  register: f5_response
```

#### Example with Python

```python
import requests
import json

# Get token from Vault
vault_addr = "http://127.0.0.1:8200"
vault_token = "YOUR_VAULT_TOKEN_HERE"  # Replace with your actual token when using
headers = {"X-Vault-Token": vault_token}
response = requests.get(f"{vault_addr}/v1/f5token/token/bigip1", headers=headers)
token_data = response.json()["data"]
f5_token = token_data["token"]

# Use token with F5 API
f5_addr = "https://172.16.10.10"
headers = {"X-F5-Auth-Token": f5_token}
response = requests.get(f"{f5_addr}/mgmt/tm/sys/version", headers=headers, verify=False)
print(json.dumps(response.json(), indent=2))
```

## Token Lifecycle Management

The plugin automatically manages the lifecycle of tokens:

1. **Creation**: Tokens are created with a specified TTL (default: 1 hour)
2. **Storage**: Token details are securely stored in Vault
3. **Expiration**: Tokens automatically expire after their TTL
4. **Cleanup**: A periodic function removes expired tokens

You can always view active tokens with:
```bash
vault read f5token/tokens
```

## Testing

A test script is provided to verify functionality:

```bash
./test_minimal.sh
```

This script:
1. Configures a connection to an F5 BIG-IP
2. Lists the connection
3. Reads the connection details
4. Generates tokens with default and custom TTLs
5. Lists active tokens

## Setting Up UI Access

To ensure full access to the plugin in the Vault UI:

```bash
vault policy write f5token-admin - <<EOF
path "f5token/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF

vault token create -policy=f5token-admin
```

Use the generated token to log into the Vault UI.

## Verifying on F5 BIG-IP

You can verify token activity on your F5 BIG-IP device:

1. SSH into your F5 device
2. Check logs for token creation and usage:
   ```bash
   tail -f /var/log/restjavad.0.log
   tail -f /var/log/secure
   ```

## Documentation

- [Plugin Testing](PLUGIN_TESTING.md): Documentation on testing the plugin
- [UI Testing](UI_TESTING.md): Documentation on using the plugin with the Vault UI
- [Project Structure](PROJECT_STRUCTURE.md): Overview of the project structure
- [Flow Diagrams](FLOW_DIAGRAM.md): Visual diagrams of token flow and plugin architecture (Mermaid format)
- [ASCII Diagrams](ASCII_DIAGRAMS.md): Simple ASCII art diagrams of token flow and architecture
- [Git Instructions](GIT_INSTRUCTIONS.md): Instructions for working with Git

## Source Control

This project is managed with Git. To contribute:

1. Clone the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

See [GIT_INSTRUCTIONS.md](GIT_INSTRUCTIONS.md) for detailed Git workflow instructions.

## License

This plugin is licensed under the MIT License.