package pivnet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type AuthResp struct {
	Token string `json:"access_token"`
}

type TokenFetcher struct {
	Endpoint          string
	RefreshToken      string
	SkipSSLValidation bool
	UserAgent         string
	ProxyAuthConfig   ProxyAuthConfig
}

func NewTokenFetcher(endpoint, refreshToken string, skipSSLValidation bool, userAgent string, proxyAuthConfig ProxyAuthConfig) *TokenFetcher {
	return &TokenFetcher{endpoint, refreshToken, skipSSLValidation, userAgent, proxyAuthConfig}
}

func (t TokenFetcher) GetToken() (string, error) {
	var transport http.RoundTripper
	var err error

	// If proxy authentication is configured, use it; otherwise use standard transport
	if t.ProxyAuthConfig.AuthType != "" {
		// Create base transport
		baseTransport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: t.SkipSSLValidation,
			},
		}

		// Parse proxy URL
		if t.ProxyAuthConfig.ProxyURL == "" {
			return "", fmt.Errorf("proxy URL is required when proxy authentication is specified")
		}
		proxyURL, err := url.Parse(t.ProxyAuthConfig.ProxyURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		baseTransport.Proxy = http.ProxyURL(proxyURL)

		// Create authenticator
		authenticator, err := NewProxyAuthenticator(t.ProxyAuthConfig)
		if err != nil {
			return "", fmt.Errorf("failed to create proxy authenticator: %w", err)
		}

		// Wrap transport with proxy authentication
		transport, err = NewProxyAuthTransport(baseTransport, authenticator)
		if err != nil {
			return "", fmt.Errorf("failed to initialize proxy authentication: %w", err)
		}
	} else {
		// Use standard transport with environment proxy support
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: t.SkipSSLValidation,
			},
			Proxy: http.ProxyFromEnvironment,
		}
	}

	httpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	body := AuthBody{RefreshToken: t.RefreshToken}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API token request body: %s", err.Error())
	}
	req, err := http.NewRequest("POST", t.Endpoint+"/authentication/access_tokens", bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json")

	if t.UserAgent != "" {
		req.Header.Add("User-Agent", t.UserAgent)
	}

	if err != nil {
		return "", fmt.Errorf("failed to construct API token request: %s", err.Error())
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API token request failed: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch API token - received status %v", resp.StatusCode)
	}

	var response AuthResp
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode API token response: %s", err.Error())
	}

	return response.Token, nil
}
