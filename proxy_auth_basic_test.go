package pivnet_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

func TestNewBasicProxyAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "with valid credentials",
			username: "testuser",
			password: "testpass",
		},
		{
			name:     "with empty username",
			username: "",
			password: "testpass",
		},
		{
			name:     "with empty password",
			username: "testuser",
			password: "",
		},
		{
			name:     "with both empty",
			username: "",
			password: "",
		},
		{
			name:     "with special characters",
			username: "user@domain.com",
			password: "p@ssw0rd!#$%",
		},
		{
			name:     "with spaces",
			username: "user name",
			password: "pass word",
		},
		{
			name:     "with colon in password",
			username: "user",
			password: "pass:word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := pivnet.NewBasicProxyAuth(tt.username, tt.password)
			if auth == nil {
				t.Error("expected non-nil BasicProxyAuth")
			}
		})
	}
}

func TestBasicProxyAuth_Authenticate(t *testing.T) {
	tests := []struct {
		name                string
		username            string
		password            string
		expectHeader        bool
		expectedHeaderValue string
	}{
		{
			name:                "valid credentials",
			username:            "testuser",
			password:            "testpass",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass")),
		},
		{
			name:                "empty username and password - no header",
			username:            "",
			password:            "",
			expectHeader:        false,
			expectedHeaderValue: "",
		},
		{
			name:                "empty username with password",
			username:            "",
			password:            "testpass",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte(":testpass")),
		},
		{
			name:                "username with empty password",
			username:            "testuser",
			password:            "",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:")),
		},
		{
			name:                "credentials with special characters",
			username:            "user@domain.com",
			password:            "p@ssw0rd!#$%^&*()",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("user@domain.com:p@ssw0rd!#$%^&*()")),
		},
		{
			name:                "credentials with colon",
			username:            "user:name",
			password:            "pass:word",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("user:name:pass:word")),
		},
		{
			name:                "credentials with spaces",
			username:            "user name",
			password:            "pass word",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("user name:pass word")),
		},
		{
			name:                "credentials with newlines",
			username:            "user\nname",
			password:            "pass\nword",
			expectHeader:        true,
			expectedHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("user\nname:pass\nword")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := pivnet.NewBasicProxyAuth(tt.username, tt.password)
			req := httptest.NewRequest("GET", "http://example.com", nil)

			err := auth.Authenticate(req)
			if err != nil {
				t.Errorf("expected no error, but got: %v", err)
				return
			}

			headerValue := req.Header.Get("Proxy-Authorization")
			if tt.expectHeader {
				if headerValue == "" {
					t.Error("expected Proxy-Authorization header to be set, but it was empty")
					return
				}
				if headerValue != tt.expectedHeaderValue {
					t.Errorf("expected header value '%s', got '%s'", tt.expectedHeaderValue, headerValue)
				}

				// Verify the header can be decoded
				if strings.HasPrefix(headerValue, "Basic ") {
					encodedCreds := strings.TrimPrefix(headerValue, "Basic ")
					decodedBytes, err := base64.StdEncoding.DecodeString(encodedCreds)
					if err != nil {
						t.Errorf("failed to decode base64: %v", err)
					}
					expectedCreds := tt.username + ":" + tt.password
					if string(decodedBytes) != expectedCreds {
						t.Errorf("decoded credentials '%s' don't match expected '%s'", string(decodedBytes), expectedCreds)
					}
				}
			} else {
				if headerValue != "" {
					t.Errorf("expected no Proxy-Authorization header, but got: %s", headerValue)
				}
			}
		})
	}
}

func TestBasicProxyAuth_WithMockProxyServer(t *testing.T) {
	t.Run("successful authentication with mock proxy", func(t *testing.T) {
		authAttempts := 0
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authAttempts++

			// Check for Proxy-Authorization header
			authHeader := r.Header.Get("Proxy-Authorization")
			if authHeader == "" {
				w.Header().Set("Proxy-Authenticate", "Basic realm=\"Mock Proxy\"")
				w.WriteHeader(http.StatusProxyAuthRequired)
				w.Write([]byte("Proxy authentication required"))
				return
			}

			// Decode and verify credentials
			expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("mockuser:mockpass"))
			if authHeader != expectedAuth {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Invalid proxy credentials"))
				return
			}

			// Success
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Authenticated through proxy"))
		}))
		defer proxyServer.Close()

		// Create auth and transport
		auth := pivnet.NewBasicProxyAuth("mockuser", "mockpass")
		transport, err := pivnet.NewProxyAuthTransport(http.DefaultTransport, auth)
		if err != nil {
			t.Fatalf("failed to create transport: %v", err)
		}

		// Make request
		client := &http.Client{Transport: transport}
		resp, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		if authAttempts != 1 {
			t.Errorf("expected 1 auth attempt, got %d", authAttempts)
		}
	})

	t.Run("proxy rejects invalid credentials", func(t *testing.T) {
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Proxy-Authorization")

			// Only accept specific credentials
			validAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("validuser:validpass"))
			if authHeader != validAuth {
				w.Header().Set("Proxy-Authenticate", "Basic realm=\"Mock Proxy\"")
				w.WriteHeader(http.StatusProxyAuthRequired)
				w.Write([]byte("Invalid credentials"))
				return
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer proxyServer.Close()

		// Use invalid credentials
		auth := pivnet.NewBasicProxyAuth("invaliduser", "invalidpass")
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

		if resp.StatusCode != http.StatusProxyAuthRequired {
			t.Errorf("expected status 407, got %d", resp.StatusCode)
		}
	})

	t.Run("proxy with special characters in credentials", func(t *testing.T) {
		specialUser := "user@domain.com"
		specialPass := "p@ss:w0rd!#$%"

		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Proxy-Authorization")
			expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(specialUser+":"+specialPass))

			if authHeader != expectedAuth {
				w.WriteHeader(http.StatusProxyAuthRequired)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Special chars handled"))
		}))
		defer proxyServer.Close()

		auth := pivnet.NewBasicProxyAuth(specialUser, specialPass)
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
	})
}
