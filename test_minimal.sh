#!/bin/bash
set -e

echo "===== F5 BIG-IP Token Plugin Basic Test ====="
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root

echo "1. Testing connection configuration..."
vault write f5token/config/connection/bigip1 \
  host="172.16.10.10" \
  username="admin" \
  password="W3lcome098!" \
  insecure_ssl=true

echo "2. Listing configured connections..."
vault list f5token/config/connections

echo "3. Reading connection details (password redacted)..."
vault read f5token/config/connection/bigip1

echo "4. Generating token with default TTL (1 hour)..."
TOKEN_INFO=$(vault read -format=json f5token/token/bigip1)
echo $TOKEN_INFO | jq

# Extract token
TOKEN=$(echo $TOKEN_INFO | jq -r .data.token)
echo "5. Generated token: $TOKEN"

echo "6. Generating another token with custom TTL (5 minutes)..."
vault read f5token/token/bigip1 ttl=300

echo "===== Test completed successfully! ====="
echo ""
echo "This token can now be used with the F5 BIG-IP REST API:"
echo "curl -k -H \"X-F5-Auth-Token: $TOKEN\" https://172.16.10.10/mgmt/tm/sys/version"
echo ""
echo "You can verify token activity by checking the logs on your F5 device:"
echo "- /var/log/restjavad.0.log"
echo "- /var/log/secure"

# Try this more comprehensive policy
vault policy write f5token-admin - <<EOF
path "f5token/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF

# Create a token with this policy
vault token create -policy=f5token-admin 

presenterm f5-plugin-demo.md 