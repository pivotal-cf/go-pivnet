package pivnet_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	pivnet "github.com/pivotal-cf/go-pivnet/v9"
)

// mockAuthenticator is a mock implementation of ProxyAuthenticator for testing
type mockAuthenticator struct {
	authenticateCalled int
	authenticateError  error
	headerToSet        string
	headerValue        string
}

func (m *mockAuthenticator) Authenticate(req *http.Request) error {
	m.authenticateCalled++
	if m.authenticateError != nil {
		return m.authenticateError
	}
	if m.headerToSet != "" {
		req.Header.Set(m.headerToSet, m.headerValue)
	}
	return nil
}

// mockRoundTripper is a mock implementation of http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripCalled int
	response        *http.Response
	err             error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.roundTripCalled++
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

func TestNewProxyAuthTransport(t *testing.T) {
	tests := []struct {
		name          string
		transport     http.RoundTripper
		authenticator pivnet.ProxyAuthenticator
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid transport and authenticator",
			transport:     &mockRoundTripper{},
			authenticator: &mockAuthenticator{},
			expectError:   false,
		},
		{
			name:          "nil transport",
			transport:     nil,
			authenticator: &mockAuthenticator{},
			expectError:   true,
			errorContains: "transport cannot be nil",
		},
		{
			name:          "nil authenticator",
			transport:     &mockRoundTripper{},
			authenticator: nil,
			expectError:   true,
			errorContains: "authenticator cannot be nil",
		},
		{
			name:          "both nil",
			transport:     nil,
			authenticator: nil,
			expectError:   true,
			errorContains: "transport cannot be nil",
		},
		{
			name:          "with http.Transport",
			transport:     &http.Transport{},
			authenticator: &mockAuthenticator{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := pivnet.NewProxyAuthTransport(tt.transport, tt.authenticator)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', but got no error", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errorContains, err)
				}
				if transport != nil {
					t.Errorf("expected nil transport on error, but got: %v", transport)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
					return
				}
				if transport == nil {
					t.Errorf("expected transport to be non-nil")
				}
			}
		})
	}
}

func TestProxyAuthTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name                   string
		authenticateError      error
		roundTripError         error
		expectError            bool
		errorContains          string
		expectedAuthCalls      int
		expectedRoundTripCalls int
	}{
		{
			name:                   "successful round trip",
			authenticateError:      nil,
			roundTripError:         nil,
			expectError:            false,
			expectedAuthCalls:      1,
			expectedRoundTripCalls: 1,
		},
		{
			name:                   "authentication fails",
			authenticateError:      fmt.Errorf("auth failed"),
			roundTripError:         nil,
			expectError:            true,
			errorContains:          "failed to authenticate proxy request",
			expectedAuthCalls:      1,
			expectedRoundTripCalls: 0,
		},
		{
			name:                   "round trip fails",
			authenticateError:      nil,
			roundTripError:         fmt.Errorf("connection failed"),
			expectError:            true,
			errorContains:          "connection failed",
			expectedAuthCalls:      1,
			expectedRoundTripCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthenticator{
				authenticateError: tt.authenticateError,
				headerToSet:       "Proxy-Authorization",
				headerValue:       "Bearer test-token",
			}
			mockRT := &mockRoundTripper{
				err: tt.roundTripError,
			}

			transport, err := pivnet.NewProxyAuthTransport(mockRT, mockAuth)
			if err != nil {
				t.Fatalf("failed to create transport: %v", err)
			}

			req := httptest.NewRequest("GET", "http://example.com", nil)
			resp, err := transport.RoundTrip(req)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', but got no error", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
					return
				}
				if resp == nil {
					t.Errorf("expected response to be non-nil")
				}
			}

			if mockAuth.authenticateCalled != tt.expectedAuthCalls {
				t.Errorf("expected %d authenticate calls, got %d", tt.expectedAuthCalls, mockAuth.authenticateCalled)
			}
			if mockRT.roundTripCalled != tt.expectedRoundTripCalls {
				t.Errorf("expected %d round trip calls, got %d", tt.expectedRoundTripCalls, mockRT.roundTripCalled)
			}
		})
	}
}

func TestProxyAuthTransport_GetProxyConnectHeader(t *testing.T) {
	t.Run("http.Transport sets GetProxyConnectHeader", func(t *testing.T) {
		mockAuth := &mockAuthenticator{
			headerToSet: "Proxy-Authorization",
			headerValue: "Bearer connect-token",
		}
		httpTransport := &http.Transport{}

		transport, err := pivnet.NewProxyAuthTransport(httpTransport, mockAuth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		// Verify GetProxyConnectHeader was set
		if httpTransport.GetProxyConnectHeader == nil {
			t.Fatal("expected GetProxyConnectHeader to be set")
		}

		// Test the GetProxyConnectHeader function
		ctx := context.Background()
		proxyURL, _ := url.Parse("http://proxy.example.com:8080")
		headers, err := httpTransport.GetProxyConnectHeader(ctx, proxyURL, "example.com:443")

		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}

		if headers.Get("Proxy-Authorization") != "Bearer connect-token" {
			t.Errorf("expected Proxy-Authorization header to be set, got: %s", headers.Get("Proxy-Authorization"))
		}

		if mockAuth.authenticateCalled != 1 {
			t.Errorf("expected 1 authenticate call, got %d", mockAuth.authenticateCalled)
		}

		// Ensure transport is properly wrapped
		if transport == nil {
			t.Error("expected transport to be non-nil")
		}
	})

	t.Run("GetProxyConnectHeader handles authentication error", func(t *testing.T) {
		mockAuth := &mockAuthenticator{
			authenticateError: fmt.Errorf("auth error"),
		}
		httpTransport := &http.Transport{}

		_, err := pivnet.NewProxyAuthTransport(httpTransport, mockAuth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		ctx := context.Background()
		proxyURL, _ := url.Parse("http://proxy.example.com:8080")
		_, err = httpTransport.GetProxyConnectHeader(ctx, proxyURL, "example.com:443")

		if err == nil {
			t.Error("expected error, but got none")
		}
		if !contains(err.Error(), "failed to authenticate CONNECT request") {
			t.Errorf("expected error to contain 'failed to authenticate CONNECT request', got: %v", err)
		}
	})

	t.Run("non-http.Transport doesn't set GetProxyConnectHeader", func(t *testing.T) {
		mockAuth := &mockAuthenticator{}
		mockRT := &mockRoundTripper{}

		transport, err := pivnet.NewProxyAuthTransport(mockRT, mockAuth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		if transport == nil {
			t.Error("expected transport to be non-nil")
		}
	})
}

func TestProxyAuthTransport_WithMockProxy(t *testing.T) {
	t.Run("mock proxy requiring authentication", func(t *testing.T) {
		// Create a mock proxy server that requires authentication
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for Proxy-Authorization header
			authHeader := r.Header.Get("Proxy-Authorization")
			if authHeader == "" {
				w.Header().Set("Proxy-Authenticate", "Basic realm=\"Test Proxy\"")
				w.WriteHeader(http.StatusProxyAuthRequired)
				w.Write([]byte("Proxy authentication required"))
				return
			}

			// Validate the auth header
			expectedAuth := "Basic dGVzdHVzZXI6dGVzdHBhc3M=" // testuser:testpass
			if authHeader != expectedAuth {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Invalid credentials"))
				return
			}

			// Authentication successful - proxy the request
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Proxy authenticated successfully"))
		}))
		defer proxyServer.Close()

		// Create authenticator
		auth := pivnet.NewBasicProxyAuth("testuser", "testpass")

		// Create transport with proxy
		transport, err := pivnet.NewProxyAuthTransport(http.DefaultTransport, auth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		// Make request through the "proxy" (simulated)
		client := &http.Client{Transport: transport}
		resp, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("mock proxy rejecting invalid credentials", func(t *testing.T) {
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Proxy-Authorization")

			// Always reject with wrong credentials
			if authHeader != "Basic Y29ycmVjdDpjcmVkcw==" { // correct:creds
				w.Header().Set("Proxy-Authenticate", "Basic realm=\"Test Proxy\"")
				w.WriteHeader(http.StatusProxyAuthRequired)
				w.Write([]byte("Authentication required"))
				return
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer proxyServer.Close()

		// Use wrong credentials
		auth := pivnet.NewBasicProxyAuth("wrong", "credentials")

		transport, err := pivnet.NewProxyAuthTransport(http.DefaultTransport, auth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		client := &http.Client{Transport: transport}
		resp, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should get 407 Proxy Authentication Required
		if resp.StatusCode != http.StatusProxyAuthRequired {
			t.Errorf("expected status 407, got %d", resp.StatusCode)
		}
	})

	t.Run("mock proxy forwarding to target server", func(t *testing.T) {
		// Create target server
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Target server response"))
		}))
		defer targetServer.Close()

		// Create proxy server that actually forwards to target
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Proxy-Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusProxyAuthRequired)
				return
			}

			// Actually forward the request to target server
			targetResp, err := http.Get(targetServer.URL)
			if err != nil {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(fmt.Sprintf("Proxy failed to reach target: %v", err)))
				return
			}
			defer targetResp.Body.Close()

			// Copy response from target
			body := make([]byte, 1024)
			n, _ := targetResp.Body.Read(body)
			w.WriteHeader(targetResp.StatusCode)
			w.Write(body[:n])
		}))
		defer proxyServer.Close()

		auth := pivnet.NewBasicProxyAuth("proxyuser", "proxypass")
		transport, err := pivnet.NewProxyAuthTransport(http.DefaultTransport, auth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		client := &http.Client{Transport: transport}
		resp, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Verify we got the response from the target server
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		responseBody := string(body[:n])
		if responseBody != "Target server response" {
			t.Errorf("expected 'Target server response', got '%s'", responseBody)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
