package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an F5 BIG-IP API client
type Client struct {
	Host       string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// TokenResponse represents the response from a token authentication request
type TokenResponse struct {
	Token struct {
		Token    string `json:"token"`
		Timeout  int64  `json:"timeout"`
		ExpireAt string `json:"expireTime,omitempty"`
	} `json:"token"`
}

// TokenRequest represents a token request to the F5 BIG-IP
type TokenRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	LoginProviderName string `json:"loginProviderName,omitempty"`
}

// NewClient creates a new F5 BIG-IP client
func NewClient(host, username, password string, insecureSSL bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSSL},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 30,
	}

	// Ensure host starts with https://
	if !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	return &Client{
		Host:       host,
		Username:   username,
		Password:   password,
		HTTPClient: httpClient,
	}
}

// GetToken authenticates to the F5 BIG-IP and retrieves an authentication token
func (c *Client) GetToken(timeout int64) (*TokenResponse, error) {
	// Construct the URL for token authentication
	url := fmt.Sprintf("%s/mgmt/shared/authn/login", c.Host)

	// Create the token request payload
	tokenReq := TokenRequest{
		Username: c.Username,
		Password: c.Password,
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling token request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating token request: %w", err)
	}

	// Set Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making token request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading token response: %w", err)
	}

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error authenticating to F5 BIG-IP: %s - %s", resp.Status, string(body))
	}

	// Parse the response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing token response: %w", err)
	}

	// If a custom timeout is specified, update the token timeout
	if timeout > 0 {
		err = c.UpdateTokenTimeout(tokenResp.Token.Token, timeout)
		if err != nil {
			return nil, fmt.Errorf("error updating token timeout: %w", err)
		}
		tokenResp.Token.Timeout = timeout
	}

	return &tokenResp, nil
}

// UpdateTokenTimeout updates the timeout for a token
func (c *Client) UpdateTokenTimeout(token string, timeout int64) error {
	// Construct the URL for token timeout update
	url := fmt.Sprintf("%s/mgmt/shared/authz/tokens/%s", c.Host, token)

	// Create timeout request payload
	timeoutReq := struct {
		Timeout int64 `json:"timeout"`
	}{
		Timeout: timeout,
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(timeoutReq)
	if err != nil {
		return fmt.Errorf("error marshaling timeout request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating timeout update request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-F5-Auth-Token", token)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making timeout update request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading timeout update response: %w", err)
	}

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error updating token timeout: %s - %s", resp.Status, string(body))
	}

	return nil
}

// RevokeToken revokes an authentication token
func (c *Client) RevokeToken(token string) error {
	// Construct the URL for token revocation
	url := fmt.Sprintf("%s/mgmt/shared/authz/tokens/%s", c.Host, token)

	// Create request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating token revocation request: %w", err)
	}

	// Set headers
	req.Header.Set("X-F5-Auth-Token", token)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making token revocation request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		// Read the response body for error details
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error revoking token: %s - %s", resp.Status, string(body))
	}

	return nil
}

// ValidateToken checks if a token is valid
func (c *Client) ValidateToken(token string) (bool, error) {
	// Construct the URL for a simple validation (getting system version)
	url := fmt.Sprintf("%s/mgmt/tm/sys/version", c.Host)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating validation request: %w", err)
	}

	// Set token header
	req.Header.Set("X-F5-Auth-Token", token)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making validation request: %w", err)
	}
	defer resp.Body.Close()

	// If status is 200, token is valid
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// If status is 401 or 403, token is invalid
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, nil
	}

	// Any other status is an error
	return false, fmt.Errorf("unexpected status when validating token: %s", resp.Status)
}
