package pivnet

import (
	"fmt"
	"net/http"
)

// ProxyAuthType represents the type of proxy authentication mechanism
type ProxyAuthType string

const (
	// ProxyAuthTypeBasic indicates HTTP Basic proxy authentication
	ProxyAuthTypeBasic ProxyAuthType = "basic"
	// ProxyAuthTypeSPNEGO indicates Kerberos/SPNEGO proxy authentication
	ProxyAuthTypeSPNEGO ProxyAuthType = "spnego"
)

// ProxyAuthenticator is an interface for proxy authentication mechanisms
type ProxyAuthenticator interface {
	// Authenticate adds authentication headers to the HTTP request
	Authenticate(req *http.Request) error
}

// NewProxyAuthenticator creates a ProxyAuthenticator based on the auth type
func NewProxyAuthenticator(config ProxyAuthConfig) (ProxyAuthenticator, error) {
	// Validate auth type is not empty
	if config.AuthType == "" {
		return nil, fmt.Errorf("proxy authentication type cannot be empty")
	}

	switch config.AuthType {
	case ProxyAuthTypeBasic:
		return NewBasicProxyAuth(config.Username, config.Password), nil

	case ProxyAuthTypeSPNEGO:
		auth, err := NewSPNEGOProxyAuth(config.Username, config.Password, config.ProxyURL, config.Krb5Config)
		if err != nil {
			return nil, err // Explicitly return nil interface
		}
		return auth, nil

	default:
		return nil, fmt.Errorf("unsupported proxy authentication type: %s (supported types: basic, spnego)", config.AuthType)
	}
}
