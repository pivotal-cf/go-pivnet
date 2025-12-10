package pivnet_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

func TestNewSPNEGOProxyAuth(t *testing.T) {
	// Create a temporary krb5.conf for testing
	tmpDir := t.TempDir()
	krb5ConfPath := filepath.Join(tmpDir, "krb5.conf")

	krb5Conf := `[libdefaults]
    default_realm = EXAMPLE.COM
    dns_lookup_realm = false
    dns_lookup_kdc = false

[realms]
    EXAMPLE.COM = {
        kdc = kdc.example.com
        admin_server = kdc.example.com
    }

[domain_realm]
    .example.com = EXAMPLE.COM
    example.com = EXAMPLE.COM
`
	if err := ioutil.WriteFile(krb5ConfPath, []byte(krb5Conf), 0644); err != nil {
		t.Fatalf("failed to create test krb5.conf: %v", err)
	}

	tests := []struct {
		name           string
		username       string
		password       string
		proxyURL       string
		krb5ConfigPath string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "empty username",
			username:       "",
			password:       "password",
			proxyURL:       "http://proxy.example.com:8080",
			krb5ConfigPath: krb5ConfPath,
			expectError:    true,
			errorContains:  "username, password, and proxyURL are required",
		},
		{
			name:           "empty password",
			username:       "user@EXAMPLE.COM",
			password:       "",
			proxyURL:       "http://proxy.example.com:8080",
			krb5ConfigPath: krb5ConfPath,
			expectError:    true,
			errorContains:  "username, password, and proxyURL are required",
		},
		{
			name:           "empty proxy URL",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "",
			krb5ConfigPath: krb5ConfPath,
			expectError:    true,
			errorContains:  "username, password, and proxyURL are required",
		},
		{
			name:           "invalid proxy URL scheme",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "ftp://proxy.example.com:8080",
			krb5ConfigPath: krb5ConfPath,
			expectError:    true,
			errorContains:  "proxy URL must start with http:// or https://",
		},
		{
			name:           "proxy URL without scheme",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "proxy.example.com:8080",
			krb5ConfigPath: krb5ConfPath,
			expectError:    true,
			errorContains:  "proxy URL must start with http:// or https://",
		},
		{
			name:           "path traversal in krb5 config",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "http://proxy.example.com:8080",
			krb5ConfigPath: "/etc/../../../etc/passwd",
			expectError:    true,
			errorContains:  "krb5 config path contains invalid path traversal",
		},
		{
			name:           "relative path traversal in krb5 config",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "http://proxy.example.com:8080",
			krb5ConfigPath: "../krb5.conf",
			expectError:    true,
			errorContains:  "krb5 config path contains invalid path traversal",
		},
		{
			name:           "non-existent krb5 config file",
			username:       "user@EXAMPLE.COM",
			password:       "password",
			proxyURL:       "http://proxy.example.com:8080",
			krb5ConfigPath: "/nonexistent/krb5.conf",
			expectError:    true,
			errorContains:  "failed to load Kerberos config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := pivnet.NewSPNEGOProxyAuth(tt.username, tt.password, tt.proxyURL, tt.krb5ConfigPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', but got no error", tt.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errorContains, err)
				}
				if auth != nil {
					t.Errorf("expected nil auth on error, but got: %v", auth)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
					return
				}
				if auth == nil {
					t.Errorf("expected auth to be non-nil")
				}
			}
		})
	}
}

func TestNewSPNEGOProxyAuth_ConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing default_realm in config", func(t *testing.T) {
		krb5ConfPath := filepath.Join(tmpDir, "no_realm.conf")
		krb5Conf := `[libdefaults]
    dns_lookup_realm = false
    dns_lookup_kdc = false
`
		if err := ioutil.WriteFile(krb5ConfPath, []byte(krb5Conf), 0644); err != nil {
			t.Fatalf("failed to create test krb5.conf: %v", err)
		}

		_, err := pivnet.NewSPNEGOProxyAuth(
			"user",
			"pass",
			"http://proxy.example.com:8080",
			krb5ConfPath,
		)

		if err == nil {
			t.Error("expected error for missing realm, but got none")
			return
		}

		if !strings.Contains(err.Error(), "domain/realm is required") {
			t.Errorf("expected error about missing realm, got: %v", err)
		}
	})

	t.Run("empty config file", func(t *testing.T) {
		krb5ConfPath := filepath.Join(tmpDir, "empty.conf")
		if err := ioutil.WriteFile(krb5ConfPath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create test krb5.conf: %v", err)
		}

		_, err := pivnet.NewSPNEGOProxyAuth(
			"user",
			"pass",
			"http://proxy.example.com:8080",
			krb5ConfPath,
		)

		if err == nil {
			t.Error("expected error for empty config, but got none")
		}
	})

	t.Run("malformed config file", func(t *testing.T) {
		krb5ConfPath := filepath.Join(tmpDir, "malformed.conf")
		if err := ioutil.WriteFile(krb5ConfPath, []byte("invalid config content [[["), 0644); err != nil {
			t.Fatalf("failed to create test krb5.conf: %v", err)
		}

		_, err := pivnet.NewSPNEGOProxyAuth(
			"user",
			"pass",
			"http://proxy.example.com:8080",
			krb5ConfPath,
		)

		if err == nil {
			t.Error("expected error for malformed config, but got none")
		}
	})
}

func TestSPNEGOProxyAuth_DefaultConfigPath(t *testing.T) {
	t.Run("uses KRB5_CONFIG environment variable", func(t *testing.T) {
		tmpDir := t.TempDir()
		krb5ConfPath := filepath.Join(tmpDir, "custom_krb5.conf")

		krb5Conf := `[libdefaults]
    default_realm = CUSTOM.REALM

[realms]
    CUSTOM.REALM = {
        kdc = kdc.custom.realm
    }
`
		if err := ioutil.WriteFile(krb5ConfPath, []byte(krb5Conf), 0644); err != nil {
			t.Fatalf("failed to create test krb5.conf: %v", err)
		}

		// Set KRB5_CONFIG environment variable
		oldEnv := os.Getenv("KRB5_CONFIG")
		os.Setenv("KRB5_CONFIG", krb5ConfPath)
		defer os.Setenv("KRB5_CONFIG", oldEnv)

		_, err := pivnet.NewSPNEGOProxyAuth(
			"user",
			"pass",
			"http://proxy.custom.realm:8080",
			"", // Empty config path should use environment variable
		)

		// Will fail on KDC connection, but should have loaded the config
		if err != nil && !strings.Contains(err.Error(), "failed to login to Kerberos") {
			if strings.Contains(err.Error(), "failed to load Kerberos config") {
				t.Errorf("failed to use KRB5_CONFIG environment variable: %v", err)
			}
		}
	})
}
