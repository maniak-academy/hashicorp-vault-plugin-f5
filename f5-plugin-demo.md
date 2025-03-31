---
author: "Sebastian Maniak"
title: "HashiCorp Vault F5 BIG-IP Token Plugin"
date: "2025"
---

# HashiCorp Vault F5 BIG-IP Token Plugin

* Generate and manage F5 BIG-IP authentication tokens
* No need for temporary user accounts
* Secure credential storage
* Automated token lifecycle management

<!-- end_slide -->

# What is the F5 BIG-IP Token Plugin?

* A Vault plugin for F5 BIG-IP authentication
* Securely stores admin credentials in Vault
* Generates REST API tokens with configurable TTLs
* Manages token lifecycle automatically
* Supports multiple F5 BIG-IP devices

<!-- end_slide -->

# Why Use API Tokens?

* **No User Account Management**
  - No temporary accounts needed
* **Lower Resource Overhead**
  - Faster authentication process
* **Fine-grained Control**
  - Precise TTL setting
* **Native F5 Support**
  - Uses built-in token API
* **Better Audit Trail**
  - Improved access tracking

<!-- end_slide -->

# Token Generation Flow (1/2)

```
+-------------+          +---------------+          +------------+
| Application |          | HashiCorp     |          | F5 Token   |
|             |          | Vault         |          | Plugin     |
+------+------+          +-------+-------+          +-----+------+
       |                         |                        |      
       | Authenticate            |                        |      
       +------------------------>|                        |      
       |                         |                        |      
       |<------------------------+                        |      
       | Vault Token             |                        |      
       |                         |                        |      
       | Request F5 Token        |                        |      
       +------------------------>|                        |      
       |                         | Forward Request        |      
       |                         +----------------------->|      
```

<!-- end_slide -->

# Token Generation Flow (2/2)

```
+-------------+          +---------------+          +------------+          +---------+
| Application |          | HashiCorp     |          | F5 Token   |          | F5      |
|             |          | Vault         |          | Plugin     |          | BIG-IP  |
+------+------+          +-------+-------+          +-----+------+          +----+----+
                                                          |                      |
                                                          | Authenticate with    |
                                                          | stored credentials   |
                                                          +--------------------->|
                                                          |                      |
                                                          |<---------------------+
                                                          | Generate Token       |
                                                          |                      |
                                                          | Store token details  |
                                                          +---+                  |
                                                          |   |                  |
                                                          |<--+                  |
```

<!-- end_slide -->

# Token Usage Flow

```
+-------------+          +---------------+          +------------+          +---------+
| Application |          | HashiCorp     |          | F5 Token   |          | F5      |
|             |          | Vault         |          | Plugin     |          | BIG-IP  |
+------+------+          +-------+-------+          +-----+------+          +----+----+
       |                         |                        |                      |
       |<------------------------+                        |                      |
       | F5 token + metadata     |                        |                      |
       |                         |                        |                      |
       | API calls with F5 token |                        |                      |
       +-------------------------------------------------------------->|        |
       |                         |                        |            |        |
       |<--------------------------------------------------------------+        |
       | API responses           |                        |                      |
       |                         |                        |                      |
       |                         |                        | Periodic cleanup     |
       |                         |                        | of expired tokens    |
       |                         |                        +---+                  |
       |                         |                        |   |                  |
       |                         |                        |<--+                  |
+------+------+          +-------+-------+          +-----+------+          +----+----+
```

<!-- end_slide -->

# Plugin Architecture (1/2)

* **Configuration Paths** 
  - `/config/connection/*`
  - Store F5 BIG-IP connection details

* **Token Paths**
  - `/token/*`
  - Generate tokens for specific connections

* **Tokens List Path**
  - `/tokens`
  - List all active tokens across connections

<!-- end_slide -->

# Plugin Architecture (2/2)

```
+-------------------------------------------+    +----------------------------+
|           HashiCorp Vault                 |    |                            |
|                                           |    |  +------------------------+|
|   +---------------+                       |    |  | F5 BIG-IP Device 1     ||
|   |   Vault API   |                       |    |  | 172.16.10.10           ||
|   +-------+-------+                       |    |  +------------------------+|
|           |                               |    |                            |
|           v                               |    |  +------------------------+|
|   +-------+------------------------+      |    |  | F5 BIG-IP Device 2     ||
|   |      F5 Token Plugin           +----------->| 172.16.10.11           ||
|   +--------------------------------+      |    |  +------------------------+|
|                                           |    |                            |
+-------------------------------------------+    |  +------------------------+|
                                                 |  | F5 BIG-IP Device N     ||
                                                 |  +------------------------+|
                                                 +----------------------------+
```

<!-- end_slide -->

# Token Lifecycle (1/2)

* **Requested**: Application requests token via Vault
* **Generated**: Plugin authenticates to F5 and gets token
* **Active**: Token stored in Vault and ready for use
* **Expired**: TTL reached, token no longer valid

<!-- end_slide -->

# Token Lifecycle (2/2)

```
                   +------------+
                   |            |
                   |  Requested |<-------------------+
                   |            |                    |
                   +-----+------+                    |
                         |                           |
                         | Plugin authenticates to F5|
                         v                           |
                   +------------+                    |
                   |            |                    |
                   | Generated  |                    |
                   |            |                    |
                   +-----+------+                    |
                         |                           |
                         | Token stored in Vault     |
                         v                           |
                   +------------+                    |
                   |            |                    |
            +----->|  Active    +--------------------+
            |      |            |                    |
            |      +-----+------+                    |
            |            |                           |
            |            | TTL reached               |
 App uses   |            v                           |
 token with |      +------------+                    |
 F5 BIG-IP  |      |            |                    |
            +------+  Expired   |                    |
                   |            |                    |
                   +------------+                    |
```

<!-- end_slide -->

# Demo: Test Overview

* `test_minimal.sh` demonstrates basic functionality:
  - Configure F5 BIG-IP connection
  - List available connections
  - Read connection details
  - Generate tokens (default and custom TTL)
  - Use token with F5 BIG-IP API
  - Verify token activity in logs

<!-- end_slide -->

# Running test_minimal.sh

```bash
./test_minimal.sh
```

Output starts with:

```
===== F5 BIG-IP Token Plugin Basic Test =====
1. Testing connection configuration...
```

<!-- end_slide -->

# Demo: Configure Connection

```bash
vault write f5token/config/connection/bigip1 \
  host="172.16.10.10" \
  username="admin" \
  password="W3lcome098!" \
  insecure_ssl=true
```

Output:
```
Success! Data written to: f5token/config/connection/bigip1
```

* Securely stores connection details
* Password is stored encrypted
* 'insecure_ssl=true' allows self-signed certificates

<!-- end_slide -->

# Demo: List Connections

```bash
vault list f5token/config/connections
```

Output:
```
2. Listing configured connections...
Keys
----
bigip1
```

* Shows all configured connections
* Each connection is a separate path
* Multiple connections can be configured

<!-- end_slide -->

# Demo: Read Connection Details

```bash
vault read f5token/config/connection/bigip1
```

Output:
```
3. Reading connection details (password redacted)...
Key             Value
---             -----
host            172.16.10.10
insecure_ssl    true
password        <sensitive>
username        admin
```

* Notice password is marked as sensitive
* Connection parameters are shown
* Verify connection is configured correctly

<!-- end_slide -->

# Demo: Generate Token (Default TTL)

```bash
TOKEN_INFO=$(vault read -format=json f5token/token/bigip1)
echo $TOKEN_INFO | jq
```

Output:
```
4. Generating token with default TTL (1 hour)...
{
  "request_id": "5c21001a-99e1-4b25-1f14-93c15cdbe9a1",
  "lease_id": "",
  "lease_duration": 0,
  "renewable": false,
  "data": {
    "expires_at": "2025-03-31T19:47:01Z",
    "host": "bigip1",
    "token": "662K5JMBE3PGM7DPQPJYB4FOF7",
    "token_id": "token_bigip1_1743446821",
    "ttl": 3600
  },
  "warnings": null,
  "mount_type": "vault-plugin-f5-token-linux"
}
```

* Default TTL is 1 hour
* Token can be extracted using jq

<!-- end_slide -->

# Demo: Extract and Use Token

```bash
TOKEN=$(echo $TOKEN_INFO | jq -r .data.token)
echo "5. Generated token: $TOKEN"
```

Output:
```
5. Generated token: 662K5JMBE3PGM7DPQPJYB4FOF7
```

```bash
curl -k -H "X-F5-Auth-Token: $TOKEN" \
  https://172.16.10.10/mgmt/tm/sys/version
```

* Token is used with the X-F5-Auth-Token header
* All F5 BIG-IP API calls use the same token
* No need to authenticate with each request

<!-- end_slide -->

# Demo: Generate Token (Custom TTL)

```bash
vault read f5token/token/bigip1 ttl=300
```

Output:
```
6. Generating another token with custom TTL (5 minutes)...
Key           Value
---           -----
expires_at    2025-03-31T18:52:01Z
host          bigip1
token         6W7YS6OXBP6XRX7LBPRFHV2PBA
token_id      token_bigip1_1743446821
ttl           5m
```

* Custom TTL of 300 seconds (5 minutes)
* Ideal for short-lived automation tasks
* Each token has a unique token_id

<!-- end_slide -->

# Test Completion

```
===== Test completed successfully! =====
```

* All test steps completed without errors
* Demo successfully showed token generation and usage
* Plugin is functioning as expected

<!-- end_slide -->

# Using the Token in Applications

* Authenticate to Vault
* Request token for specific F5 device 
* Use token in API requests
* Let it expire automatically
* No cleanup needed after use

<!-- end_slide -->

# Integration: Ansible

```yaml
- name: Get F5 token from Vault
  uri:
    url: "{{ vault_addr }}/v1/f5token/token/bigip1"
    method: GET
    headers:
      X-Vault-Token: "{{ vault_token }}"
  register: f5_token_response

- name: Use F5 token in an API call
  uri:
    url: "https://172.16.10.10/mgmt/tm/sys/version"
    method: GET
    headers:
      X-F5-Auth-Token: "{{ f5_token_response.json.data.token }}"
```

<!-- end_slide -->

# Integration: Python

```python
import requests

# Get token from Vault
vault_addr = "http://127.0.0.1:8200"
vault_token = "YOUR_VAULT_TOKEN"
headers = {"X-Vault-Token": vault_token}
response = requests.get(
    f"{vault_addr}/v1/f5token/token/bigip1", 
    headers=headers
)
f5_token = response.json()["data"]["token"]

# Use token with F5 API
f5_addr = "https://172.16.10.10"
api_headers = {"X-F5-Auth-Token": f5_token}
response = requests.get(
    f"{f5_addr}/mgmt/tm/sys/version", 
    headers=api_headers, 
    verify=False
)
```

<!-- end_slide -->

# Key Benefits

* **Secure**
  - Credentials never leave Vault
  - Only tokens are exposed to applications

* **Automated**
  - Token lifecycle is fully managed
  - Expired tokens are cleaned up automatically

* **Scalable**
  - Support for multiple F5 BIG-IP devices
  - Single Vault instance for all tokens

<!-- end_slide -->

# Summary

* Configure F5 BIG-IP connections in Vault
* Generate tokens with custom TTLs
* Use tokens with any F5 BIG-IP API client
* Track all active tokens across connections
* Simplify F5 BIG-IP authentication

<!-- end_slide -->

# Thank You!

Questions?

<!-- end_slide -->