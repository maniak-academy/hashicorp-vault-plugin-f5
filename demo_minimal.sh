#!/bin/bash
set -e

# Function to explain the step, wait, and then run command
explain_and_run() {
    local explanation=$1
    local command=$2
    
    echo -e "\n\033[1;36m==================================================\033[0m"
    echo -e "\033[1;36m$explanation\033[0m"
    echo -e "\033[1;36m==================================================\033[0m\n"
    
    echo -e "\033[1;33mCommand:\033[0m $command"
    echo -e "\033[1;33mExecuting in 5 seconds...\033[0m"
    sleep 5
    echo -e "\033[1;32mOutput:\033[0m\n"
    eval "$command"
    echo ""
    sleep 1
}

# Set up environment
echo -e "\n\033[1;35m===== F5 BIG-IP Token Plugin Interactive Demo =====\033[0m\n"
echo "Setting up environment variables for Vault..."
sleep 2

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root

echo -e "VAULT_ADDR=$VAULT_ADDR\nVAULT_TOKEN=$VAULT_TOKEN\n"
sleep 3

# Step 1: Configure connection
explain_and_run "Step 1: Configuring a connection to F5 BIG-IP device
This step securely stores F5 BIG-IP connection details in Vault.
The password will be encrypted in Vault's storage." \
"vault write f5token/config/connection/bigip1 \\
  host=\"172.16.10.10\" \\
  username=\"admin\" \\
  password=\"W3lcome098!\" \\
  insecure_ssl=true"

# Step 2: List connections
explain_and_run "Step 2: Listing all configured F5 BIG-IP connections
This shows all connections that have been set up in the plugin.
Each connection can generate tokens independently." \
"vault list f5token/config/connections"

# Step 3: Read connection details
explain_and_run "Step 3: Reading connection details
This shows the connection parameters (note: password will be redacted).
We can verify our connection is configured correctly." \
"vault read f5token/config/connection/bigip1"

# Step 4: Generate token with default TTL
explain_and_run "Step 4: Generating a token with default TTL (1 hour)
This authenticates to F5 with stored credentials and creates a token.
The token will be valid for 1 hour by default." \
"vault read -format=json f5token/token/bigip1 | jq"

# Extract token to a variable
echo "Extracting token for later use..."
TOKEN_INFO=$(vault read -format=json f5token/token/bigip1)
TOKEN=$(echo $TOKEN_INFO | jq -r .data.token)
sleep 2

echo "Retrieved token: $TOKEN"
sleep 3

# Step 5: Generate token with custom TTL
explain_and_run "Step 5: Generating another token with custom TTL (5 minutes)
This creates a token with a shorter lifespan of 300 seconds (5 minutes).
Custom TTLs allow for fine-grained control over token lifetimes." \
"vault read f5token/token/bigip1 ttl=300"

# Step 6: How to use the token
explain_and_run "Step 6: Demonstrating how to use the token
The extracted token can be used in API calls to F5 BIG-IP.
Here's the curl command you would use:" \
"echo \"curl -k -H \\\"X-F5-Auth-Token: $TOKEN\\\" https://172.16.10.10/mgmt/tm/sys/version\""

# Step 7: Show policy configuration
explain_and_run "Step 7: Setting up a Vault policy for accessing the plugin
This creates a policy with full access to the plugin paths.
You can use this policy to create tokens for administrators." \
"vault policy write f5token-admin - <<EOF
path \"f5token/*\" {
  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\", \"sudo\"]
}
EOF"

# Step 8: Create a token with the policy
explain_and_run "Step 8: Creating a Vault token with the F5 token plugin policy
This generates a Vault token with permissions to use the F5 token plugin.
You can distribute this token to authorized users/applications." \
"vault token create -policy=f5token-admin"

# Conclusion
echo -e "\n\033[1;35m===== Demo completed successfully! =====\033[0m\n"
echo "This demo showed the basic functionality of the F5 BIG-IP Token Plugin:"
echo " - Configuring connections to F5 BIG-IP devices"
echo " - Generating tokens with different TTLs"
echo " - Using tokens with the F5 BIG-IP API"
echo " - Setting up policies for plugin access"
echo ""
echo "For more information, see the README.md file."
echo "" 