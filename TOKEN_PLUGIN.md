# F5 BIG-IP Token Vault Plugin

This plugin allows HashiCorp Vault to manage F5 BIG-IP authentication tokens. Instead of creating temporary user accounts, this plugin authenticates to the F5 BIG-IP and obtains API tokens that can be used for subsequent API calls.

## Features

- Configure multiple F5 BIG-IP connections
- Generate authentication tokens with customizable TTL (lease time)
- Automatically revoke expired tokens
- Secure storage of F5 admin credentials in Vault
- Manual token revocation

## Why Use API Tokens?

Using API tokens instead of temporary user accounts provides several advantages:

1. No need to create/delete user accounts on the F5 BIG-IP
2. Lower overhead and faster authentication
3. Fine-grained control over token expiration
4. Native support in the F5 BIG-IP API
5. Easier to audit and track

## Installation

### Building the Plugin

```shell
make build-token          # Build for your local platform
make build-token-linux    # Build for Linux (required for Docker)
```

### Registering the Plugin with Vault

With Vault running:

```shell
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

### Configure a Connection to F5 BIG-IP

```shell
vault write f5token/config/connection/bigip1 \
    host="10.0.0.1" \
    username="admin" \
    password="password" \
    insecure_ssl=true
```

### List All Configured Connections

```shell
vault list f5token/config/connections
```

### Generate an API Token

```shell
vault read f5token/token/bigip1 ttl=3600
```

This will return:
```
Key         Value
---         -----
token_id    token_bigip1_1610000000
token       ABCDEF123456...
host        bigip1
expires_at  2023-01-01T00:00:00Z
ttl         3600
```

### Revoke a Token

```shell
vault write f5token/revoke/token_bigip1_1610000000
```

## Using with Applications

Applications can request tokens from Vault when needed, typically at the start of their operation:

1. Authenticate to Vault
2. Request a token for the F5 BIG-IP
3. Use the token for all F5 BIG-IP API calls
4. Optionally revoke the token when done, or let it expire

### Example with Ansible

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
    f5_token_id: "{{ f5_token_response.json.data.token_id }}"
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

## Automatic Token Cleanup

The plugin automatically revokes expired tokens in the background. This ensures that tokens are properly cleaned up even if the client doesn't explicitly revoke them. 