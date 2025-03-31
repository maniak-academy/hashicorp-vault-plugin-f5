package bigiptoken

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/smaniak/vault-plugin-f5/pkg/bigiptoken/api"
)

// Factory returns a new backend as logical.Backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// f5TokenBackend defines the F5 BIG-IP token backend structure
type f5TokenBackend struct {
	*framework.Backend
	lock sync.RWMutex
}

// Connection represents a connection to an F5 BIG-IP device
type Connection struct {
	Host        string `json:"host"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	InsecureSSL bool   `json:"insecure_ssl"`
}

// TokenEntry represents a stored F5 BIG-IP token
type TokenEntry struct {
	Token     string    `json:"token"`
	Host      string    `json:"host"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// Backend creates a new f5TokenBackend
func Backend() *f5TokenBackend {
	var b f5TokenBackend

	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(backendHelp),
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config/connection/",
				"tokens/",
			},
		},
		Paths: framework.PathAppend(
			[]*framework.Path{
				pathConfigConnection(&b),
				pathConfigConnectionList(&b),
				pathToken(&b),
				pathTokensList(&b),
			},
		),
		PeriodicFunc: b.cleanupExpiredTokens,
	}

	return &b
}

// pathConfigConnection defines the path for F5 BIG-IP connection configuration
func pathConfigConnection(b *f5TokenBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config/connection/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Unique name for the F5 BIG-IP connection",
				Required:    true,
			},
			"host": {
				Type:        framework.TypeString,
				Description: "F5 BIG-IP hostname or IP address",
				Required:    true,
			},
			"username": {
				Type:        framework.TypeString,
				Description: "Username for F5 BIG-IP authentication",
				Required:    true,
			},
			"password": {
				Type:        framework.TypeString,
				Description: "Password for F5 BIG-IP authentication",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Sensitive: true,
				},
			},
			"insecure_ssl": {
				Type:        framework.TypeBool,
				Description: "Allow insecure SSL connections (not recommended)",
				Default:     false,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConnectionRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConnectionWrite,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConnectionWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConnectionDelete,
			},
		},

		ExistenceCheck: b.connectionExistenceCheck,

		HelpSynopsis:    "Configure a connection to an F5 BIG-IP system",
		HelpDescription: "This endpoint configures connection details for an F5 BIG-IP system, including host, credentials, and SSL settings.",
	}
}

// pathConfigConnectionList defines the path for listing available F5 BIG-IP connections
func pathConfigConnectionList(b *f5TokenBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config/connections/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathConnectionList,
			},
		},

		HelpSynopsis:    "List all configured F5 BIG-IP connections",
		HelpDescription: "This endpoint lists all configured F5 BIG-IP connections by name.",
	}
}

// pathToken defines the path for generating F5 BIG-IP tokens
func pathToken(b *f5TokenBackend) *framework.Path {
	return &framework.Path{
		Pattern: "token/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the F5 BIG-IP connection to use",
				Required:    true,
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "TTL for the token (in seconds)",
				Default:     3600, // 1 hour default
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathTokenRead,
			},
		},

		HelpSynopsis:    "Generate an F5 BIG-IP authentication token",
		HelpDescription: "This endpoint generates and returns an F5 BIG-IP authentication token with the specified TTL.",
	}
}

// pathTokensList defines the path for listing all tokens
func pathTokensList(b *f5TokenBackend) *framework.Path {
	return &framework.Path{
		Pattern: "tokens/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathTokensListRead,
			},
		},

		HelpSynopsis:    "List all F5 BIG-IP tokens",
		HelpDescription: "This endpoint lists all F5 BIG-IP tokens by ID.",
	}
}

// connectionExistenceCheck checks if a connection exists
func (b *f5TokenBackend) connectionExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	name := data.Get("name").(string)
	entry, err := req.Storage.Get(ctx, "config/connection/"+name)
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

// pathConnectionWrite handles config/connection write operations
func (b *f5TokenBackend) pathConnectionWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.Backend.Logger().Debug("pathConnectionWrite called", "request_path", req.Path, "operation", req.Operation)

	name := data.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("connection name cannot be empty"), nil
	}

	var host, username, password string
	var insecureSSL bool

	// Handle both direct parameters and JSON input
	if h, ok := data.GetOk("host"); ok {
		host = h.(string)
	}
	if u, ok := data.GetOk("username"); ok {
		username = u.(string)
	}
	if p, ok := data.GetOk("password"); ok {
		password = p.(string)
	}
	if i, ok := data.GetOk("insecure_ssl"); ok {
		insecureSSL = i.(bool)
	}

	if host == "" || username == "" || password == "" {
		return logical.ErrorResponse("host, username, and password are required"), nil
	}

	// Log what we're doing (without sensitive info)
	b.Backend.Logger().Info("configuring connection", "name", name, "host", host)

	// Create configuration entry
	connection := &Connection{
		Host:        host,
		Username:    username,
		Password:    password,
		InsecureSSL: insecureSSL,
	}

	// Test the connection by getting a token
	client := api.NewClient(host, username, password, insecureSSL)
	tokenResp, err := client.GetToken(60) // Short-lived test token
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("failed to connect to F5 BIG-IP: %s", err)), nil
	}

	// Revoke the test token, we don't need it
	if err := client.RevokeToken(tokenResp.Token.Token); err != nil {
		// Just log this error, don't fail the operation
		b.Backend.Logger().Warn("failed to revoke test token", "error", err)
	}

	// Store the connection config
	entry, err := logical.StorageEntryJSON("config/connection/"+name, connection)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"success": true,
			"host":    host,
			"status":  "Connection configured and tested successfully",
		},
	}, nil
}

// pathConnectionRead handles config/connection read operations
func (b *f5TokenBackend) pathConnectionRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("connection name cannot be empty"), nil
	}

	entry, err := req.Storage.Get(ctx, "config/connection/"+name)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var connection Connection
	if err := entry.DecodeJSON(&connection); err != nil {
		return nil, err
	}

	// Return all but the password
	resp := &logical.Response{
		Data: map[string]interface{}{
			"host":         connection.Host,
			"username":     connection.Username,
			"insecure_ssl": connection.InsecureSSL,
		},
	}

	return resp, nil
}

// pathConnectionDelete handles config/connection delete operations
func (b *f5TokenBackend) pathConnectionDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("connection name cannot be empty"), nil
	}

	// Remove the connection configuration
	if err := req.Storage.Delete(ctx, "config/connection/"+name); err != nil {
		return nil, err
	}

	return nil, nil
}

// pathConnectionList handles config/connections list operations
func (b *f5TokenBackend) pathConnectionList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	connections, err := req.Storage.List(ctx, "config/connection/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(connections), nil
}

// getF5Client creates an F5 API client from a named connection configuration
func (b *f5TokenBackend) getF5Client(ctx context.Context, storage logical.Storage, name string) (*api.Client, error) {
	entry, err := storage.Get(ctx, "config/connection/"+name)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("connection %s not found", name)
	}

	var connection Connection
	if err := entry.DecodeJSON(&connection); err != nil {
		return nil, err
	}

	return api.NewClient(connection.Host, connection.Username, connection.Password, connection.InsecureSSL), nil
}

// pathTokenRead handles token/ read operations to generate tokens
func (b *f5TokenBackend) pathTokenRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	ttl := data.Get("ttl").(int)

	if name == "" {
		return logical.ErrorResponse("connection name cannot be empty"), nil
	}

	// Generate a token ID
	tokenID := fmt.Sprintf("token_%s_%d", name, time.Now().Unix())

	// Retrieve the F5 client for the specified host
	client, err := b.getF5Client(ctx, req.Storage, name)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("error getting F5 client: %s", err)), nil
	}

	// Get token from F5 BIG-IP
	tokenResp, err := client.GetToken(int64(ttl))
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("error generating token: %s", err)), nil
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)

	// Create and store token record
	tokenEntry := &TokenEntry{
		Token:     tokenResp.Token.Token,
		Host:      name,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	// Store the token
	entry, err := logical.StorageEntryJSON("tokens/"+tokenID, tokenEntry)
	if err != nil {
		// Attempt to revoke the token if we can't store it
		_ = client.RevokeToken(tokenResp.Token.Token)
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		// Attempt to revoke the token if we can't store it
		_ = client.RevokeToken(tokenResp.Token.Token)
		return nil, err
	}

	// Return the token and metadata
	resp := &logical.Response{
		Data: map[string]interface{}{
			"token_id":   tokenID,
			"token":      tokenResp.Token.Token,
			"host":       name,
			"expires_at": expiresAt.Format(time.RFC3339),
			"ttl":        ttl,
		},
	}

	return resp, nil
}

// cleanupExpiredTokens is a periodic function to clean up expired tokens
func (b *f5TokenBackend) cleanupExpiredTokens(ctx context.Context, req *logical.Request) error {
	tokenIDs, err := req.Storage.List(ctx, "tokens/")
	if err != nil {
		return err
	}

	now := time.Now()

	for _, tokenID := range tokenIDs {
		tokenEntryRaw, err := req.Storage.Get(ctx, "tokens/"+tokenID)
		if err != nil {
			b.Backend.Logger().Error("error retrieving token", "token_id", tokenID, "error", err)
			continue
		}
		if tokenEntryRaw == nil {
			continue
		}

		var tokenEntry TokenEntry
		if err := tokenEntryRaw.DecodeJSON(&tokenEntry); err != nil {
			b.Backend.Logger().Error("error decoding token", "token_id", tokenID, "error", err)
			continue
		}

		// Skip if already inactive
		if !tokenEntry.IsActive {
			continue
		}

		// Check if expired
		if now.After(tokenEntry.ExpiresAt) {
			// Get the F5 client
			client, err := b.getF5Client(ctx, req.Storage, tokenEntry.Host)
			if err != nil {
				b.Backend.Logger().Error("error getting F5 client for cleanup", "token_id", tokenID, "error", err)
				continue
			}

			// Revoke the token in F5
			if err := client.RevokeToken(tokenEntry.Token); err != nil {
				b.Backend.Logger().Warn("failed to revoke expired token", "token_id", tokenID, "error", err)
			}

			// Mark token as inactive
			tokenEntry.IsActive = false

			// Update the token entry
			entry, err := logical.StorageEntryJSON("tokens/"+tokenID, tokenEntry)
			if err != nil {
				b.Backend.Logger().Error("error creating storage entry", "token_id", tokenID, "error", err)
				continue
			}

			if err := req.Storage.Put(ctx, entry); err != nil {
				b.Backend.Logger().Error("error updating token entry", "token_id", tokenID, "error", err)
			}
		}
	}

	return nil
}

// pathTokensListRead handles tokens/ read operations
func (b *f5TokenBackend) pathTokensListRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Get list of tokens
	tokens, err := req.Storage.List(ctx, "tokens/")
	if err != nil {
		return nil, err
	}

	// For each token, get details and return a more informative list
	var tokenDetails []map[string]interface{}
	for _, tokenID := range tokens {
		tokenEntryRaw, err := req.Storage.Get(ctx, "tokens/"+tokenID)
		if err != nil {
			continue
		}
		if tokenEntryRaw == nil {
			continue
		}

		var tokenEntry TokenEntry
		if err := tokenEntryRaw.DecodeJSON(&tokenEntry); err != nil {
			continue
		}

		// Only show active tokens
		if !tokenEntry.IsActive {
			continue
		}

		// Check if token is expired
		if time.Now().After(tokenEntry.ExpiresAt) {
			continue
		}

		detail := map[string]interface{}{
			"token_id":   tokenID,
			"host":       tokenEntry.Host,
			"created_at": tokenEntry.CreatedAt.Format(time.RFC3339),
			"expires_at": tokenEntry.ExpiresAt.Format(time.RFC3339),
			"active":     tokenEntry.IsActive,
		}
		tokenDetails = append(tokenDetails, detail)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"tokens": tokenDetails,
		},
	}, nil
}

// Help text
const backendHelp = `
The F5 BIG-IP Token secrets backend dynamically generates and manages
authentication tokens for F5 BIG-IP devices.

After configuring a connection to your F5 BIG-IP system, you can request
an authentication token which can be used for API access. The tokens have
configurable TTLs and are automatically revoked when they expire.
`
