package pivnet

import (
	"encoding/base64"
	"net/http"
)

// BasicProxyAuth implements HTTP Basic authentication for proxies
type BasicProxyAuth struct {
	username string
	password string
}

// NewBasicProxyAuth creates a new BasicProxyAuth authenticator
func NewBasicProxyAuth(username, password string) *BasicProxyAuth {
	return &BasicProxyAuth{
		username: username,
		password: password,
	}
}

// Authenticate adds the Proxy-Authorization header with Basic auth
func (b *BasicProxyAuth) Authenticate(req *http.Request) error {
	if b.username == "" && b.password == "" {
		return nil
	}

	// Create Basic auth string: "username:password" encoded in base64
	auth := b.username + ":" + b.password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Proxy-Authorization", "Basic "+encodedAuth)

	return nil
}
