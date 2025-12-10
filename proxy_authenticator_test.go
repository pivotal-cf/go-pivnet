package pivnet_test

import (
	"strings"
	"testing"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

func TestNewProxyAuthenticator(t *testing.T) {
	tests := []struct {
		name           string
		config         pivnet.ProxyAuthConfig
		expectError    bool
		errorContains  string
		expectAuthType string // Expected type of authenticator
	}{
		// Basic Authentication - Success Cases
		{
			name: "basic auth with valid credentials",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeBasic,
				Username: "testuser",
				Password: "testpass",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:    false,
			expectAuthType: "BasicProxyAuth",
		},
		{
			name: "basic auth with empty username and password",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeBasic,
				Username: "",
				Password: "",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:    false,
			expectAuthType: "BasicProxyAuth",
		},
		{
			name: "basic auth with special characters in password",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeBasic,
				Username: "user",
				Password: "p@ssw0rd!#$%",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:    false,
			expectAuthType: "BasicProxyAuth",
		},
		// SPNEGO Authentication - Error Cases (can't test success without real Kerberos, will handle this in integration tests)
		{
			name: "spnego with empty username",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeSPNEGO,
				Username: "",
				Password: "password",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "username, password, and proxyURL are required",
		},
		{
			name: "spnego with empty password",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeSPNEGO,
				Username: "user@REALM.COM",
				Password: "",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "username, password, and proxyURL are required",
		},
		{
			name: "spnego with empty proxy URL",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeSPNEGO,
				Username: "user@REALM.COM",
				Password: "password",
				ProxyURL: "",
			},
			expectError:   true,
			errorContains: "username, password, and proxyURL are required",
		},
		{
			name: "spnego with invalid proxy URL scheme",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeSPNEGO,
				Username: "user@REALM.COM",
				Password: "password",
				ProxyURL: "ftp://proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "proxy URL must start with http:// or https://",
		},
		{
			name: "spnego with proxy URL without scheme",
			config: pivnet.ProxyAuthConfig{
				AuthType: pivnet.ProxyAuthTypeSPNEGO,
				Username: "user@REALM.COM",
				Password: "password",
				ProxyURL: "proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "proxy URL must start with http:// or https://",
		},
		{
			name: "spnego with path traversal in krb5 config",
			config: pivnet.ProxyAuthConfig{
				AuthType:   pivnet.ProxyAuthTypeSPNEGO,
				Username:   "user@REALM.COM",
				Password:   "password",
				ProxyURL:   "http://proxy.example.com:8080",
				Krb5Config: "/etc/../../../etc/passwd",
			},
			expectError:   true,
			errorContains: "krb5 config path contains invalid path traversal",
		},
		{
			name: "spnego with relative path traversal in krb5 config",
			config: pivnet.ProxyAuthConfig{
				AuthType:   pivnet.ProxyAuthTypeSPNEGO,
				Username:   "user@REALM.COM",
				Password:   "password",
				ProxyURL:   "http://proxy.example.com:8080",
				Krb5Config: "../krb5.conf",
			},
			expectError:   true,
			errorContains: "krb5 config path contains invalid path traversal",
		},

		// General Error Cases
		{
			name: "empty auth type",
			config: pivnet.ProxyAuthConfig{
				AuthType: "",
				Username: "user",
				Password: "pass",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "proxy authentication type cannot be empty",
		},
		{
			name: "unsupported auth type",
			config: pivnet.ProxyAuthConfig{
				AuthType: "digest",
				Username: "user",
				Password: "pass",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError:   true,
			errorContains: "unsupported proxy authentication type",
		},
		{
			name: "case insensitive basic auth type",
			config: pivnet.ProxyAuthConfig{
				AuthType: "BASIC",
				Username: "testuser",
				Password: "testpass",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError: true,
		},
		{
			name: "case insensitive spnego auth type (will fail on krb5 config)",
			config: pivnet.ProxyAuthConfig{
				AuthType: "SPNEGO",
				Username: "user@REALM.COM",
				Password: "password",
				ProxyURL: "http://proxy.example.com:8080",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator, err := pivnet.NewProxyAuthenticator(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', but got no error", tt.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errorContains, err)
				}
				if authenticator != nil {
					t.Errorf("expected nil authenticator on error, but got: %v", authenticator)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
					return
				}
				if authenticator == nil {
					t.Errorf("expected authenticator to be non-nil")
				}
			}
		})
	}
}
