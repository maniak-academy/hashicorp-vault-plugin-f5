#!/bin/bash

# Set Vault environment variables
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root

# Connection parameters
CONNECTION_NAME="bigip1"
HOST="172.16.10.10"
USERNAME="admin"
PASSWORD="W3lcome098!"
INSECURE_SSL=true

# Create JSON payload file
cat > connection.json << EOF
{
  "host": "${HOST}",
  "username": "${USERNAME}",
  "password": "${PASSWORD}",
  "insecure_ssl": ${INSECURE_SSL}
}
EOF

# Configure the connection
echo "Configuring connection to ${HOST}..."
vault write "f5token/config/connection/${CONNECTION_NAME}" host="${HOST}" username="${USERNAME}" password="${PASSWORD}" insecure_ssl="${INSECURE_SSL}"

# Check status
if [ $? -eq 0 ]; then
  echo "Connection configured successfully."
else
  echo "Failed to configure connection."
  echo "Trying alternate method..."
  vault write "f5token/config/connection/${CONNECTION_NAME}" @connection.json
  if [ $? -eq 0 ]; then
    echo "Connection configured successfully using JSON file."
  else
    echo "Failed to configure connection using JSON file."
    echo "Listing paths for debugging..."
    vault path-help f5token
    exit 1
  fi
fi

# Clean up
rm -f connection.json

# List connections
echo "Listing configured connections:"
vault list f5token/config/connections 