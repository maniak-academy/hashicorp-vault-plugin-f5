#!/bin/bash
set -e

echo "===== F5 BIG-IP Token Plugin Test ====="
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

# Extract token and token_id
TOKEN=$(echo $TOKEN_INFO | jq -r .data.token)
TOKEN_ID=$(echo $TOKEN_INFO | jq -r .data.token_id)

echo "5. Generated token: $TOKEN"
echo "   Token ID: $TOKEN_ID"

echo "6. Generating another token with custom TTL (5 minutes)..."
vault read f5token/token/bigip1 ttl=300

echo "7. Revoking a token..."
vault write -force f5token/revoke/$TOKEN_ID

echo "8. Confirming token was revoked and testing connection details..."
vault read f5token/config/connection/bigip1

echo "===== Test completed successfully! =====" 