package pivnet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ProxyAuthTransport wraps an http.RoundTripper and adds proxy authentication
// using a pluggable ProxyAuthenticator interface
type ProxyAuthTransport struct {
	Transport     http.RoundTripper
	Authenticator ProxyAuthenticator
}

// NewProxyAuthTransport creates a new ProxyAuthTransport with the given authenticator
// It configures the underlying transport to add authentication headers to both
// regular HTTP requests and HTTPS CONNECT requests.
//
// Authentication Flow:
//   - HTTP requests: Authenticate() called in RoundTrip for each request
//   - HTTPS requests: Authenticate() called in GetProxyConnectHeader for CONNECT,
//     then called again in RoundTrip for requests through the tunnel (harmless, as
//     the header goes through encrypted tunnel and proxy doesn't see it)
func NewProxyAuthTransport(transport http.RoundTripper, authenticator ProxyAuthenticator) (*ProxyAuthTransport, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport cannot be nil")
	}
	if authenticator == nil {
		return nil, fmt.Errorf("authenticator cannot be nil")
	}

	// If the transport is an *http.Transport, configure GetProxyConnectHeader
	// to add authentication headers to CONNECT requests (for HTTPS through proxy)
	if httpTransport, ok := transport.(*http.Transport); ok {
		httpTransport.GetProxyConnectHeader = func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error) {
			header := http.Header{}
			// Create a dummy request to get the auth header
			// This is called once per CONNECT (once per HTTPS connection, not per request)
			dummyReq := &http.Request{Header: http.Header{}}
			if err := authenticator.Authenticate(dummyReq); err != nil {
				return header, fmt.Errorf("failed to authenticate CONNECT request: %w", err)
			}
			// Copy the Proxy-Authorization header
			if authHeader := dummyReq.Header.Get("Proxy-Authorization"); authHeader != "" {
				header.Set("Proxy-Authorization", authHeader)
			}
			return header, nil
		}
	}

	return &ProxyAuthTransport{
		Transport:     transport,
		Authenticator: authenticator,
	}, nil
}

// RoundTrip executes a single HTTP transaction, adding proxy authentication
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add authentication to the request
	if err := t.Authenticator.Authenticate(req); err != nil {
		return nil, fmt.Errorf("failed to authenticate proxy request: %w", err)
	}

	// Execute the request with the underlying transport
	return t.Transport.RoundTrip(req)
}
