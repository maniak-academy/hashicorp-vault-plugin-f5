.PHONY: build build-linux test test-integration clean build-token build-token-linux

BINARY_NAME=vault-plugin-f5
LINUX_BINARY_NAME=vault-plugin-f5-linux
TOKEN_BINARY_NAME=vault-plugin-f5-token
TOKEN_LINUX_BINARY_NAME=vault-plugin-f5-token-linux

build:
	go build -o $(BINARY_NAME) ./cmd/f5-vault-plugin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_BINARY_NAME) ./cmd/f5-vault-plugin

build-token:
	go build -o $(TOKEN_BINARY_NAME) ./cmd/f5-token-plugin

build-token-linux:
	GOOS=linux GOARCH=amd64 go build -o $(TOKEN_LINUX_BINARY_NAME) ./cmd/f5-token-plugin

test:
	go test -v ./pkg/f5 ./pkg/bigiptoken/...

test-integration:
	go test -v -tags=integration ./pkg/f5 ./pkg/bigiptoken/...

clean:
	rm -f $(BINARY_NAME) $(LINUX_BINARY_NAME) $(TOKEN_BINARY_NAME) $(TOKEN_LINUX_BINARY_NAME)

# Docker-based testing
docker-dev:
	docker run --name vault-dev -d -p 8200:8200 \
		-v $(shell pwd):/vault/plugins \
		-e VAULT_DEV_ROOT_TOKEN_ID=root \
		-e "VAULT_LOCAL_CONFIG={\"plugin_directory\":\"/vault/plugins\"}" \
		-e "VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200" \
		hashicorp/vault:latest

register-plugin: build-linux
	@echo "Calculating SHA256 sum..."
	@SHA256=$$(shasum -a 256 $(LINUX_BINARY_NAME) | cut -d ' ' -f1) && \
	echo "Plugin SHA256: $$SHA256" && \
	export VAULT_ADDR=http://127.0.0.1:8200 && \
	export VAULT_TOKEN=root && \
	vault plugin register -sha256=$$SHA256 secret $(LINUX_BINARY_NAME) && \
	vault secrets enable -path=f5 $(LINUX_BINARY_NAME)

register-token-plugin: build-token-linux
	@echo "Calculating SHA256 sum..."
	@SHA256=$$(shasum -a 256 $(TOKEN_LINUX_BINARY_NAME) | cut -d ' ' -f1) && \
	echo "Plugin SHA256: $$SHA256" && \
	export VAULT_ADDR=http://127.0.0.1:8200 && \
	export VAULT_TOKEN=root && \
	vault plugin register -sha256=$$SHA256 secret $(TOKEN_LINUX_BINARY_NAME) && \
	vault secrets enable -path=f5token $(TOKEN_LINUX_BINARY_NAME)

docker-clean:
	docker stop vault-dev || true
	docker rm vault-dev || true 