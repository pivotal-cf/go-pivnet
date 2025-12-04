package pivnet

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// SPNEGOProxyAuth implements Kerberos/SPNEGO authentication for proxies
type SPNEGOProxyAuth struct {
	username       string
	password       string
	proxyURL       string
	kerberosClient *client.Client
}

// NewSPNEGOProxyAuth creates a new SPNEGOProxyAuth authenticator
func NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath string) (*SPNEGOProxyAuth, error) {
	if username == "" || password == "" || proxyURL == "" {
		return nil, fmt.Errorf("username, password, and proxyURL are required for SPNEGO authentication")
	}

	// Validate proxy URL scheme
	if !strings.HasPrefix(proxyURL, "http://") && !strings.HasPrefix(proxyURL, "https://") {
		return nil, fmt.Errorf("proxy URL must start with http:// or https://")
	}

	// Validate krb5ConfigPath for path traversal
	if krb5ConfigPath != "" {
		// Check for path traversal patterns in the original path
		if strings.Contains(krb5ConfigPath, "..") {
			return nil, fmt.Errorf("krb5 config path contains invalid path traversal")
		}
	}

	// Load Kerberos configuration
	var krb5conf *config.Config
	var err error

	if krb5ConfigPath != "" {
		// Load configuration from file
		krb5conf, err = config.Load(krb5ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load Kerberos config from %s: %w", krb5ConfigPath, err)
		}
	} else {
		// Use default config path
		defaultConfigPath := configPath()
		krb5conf, err = config.Load(defaultConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load Kerberos config from %s: %w", defaultConfigPath, err)
		}
	}

	// Extract default realm from krb5.conf
	domain := krb5conf.LibDefaults.DefaultRealm
	if domain == "" {
		return nil, fmt.Errorf("domain/realm is required: please configure default_realm in krb5.conf")
	}

	// Create Kerberos client with username and password
	kerberosClient := client.NewWithPassword(
		username,
		strings.ToUpper(domain),
		password,
		krb5conf,
		client.DisablePAFXFAST(true),
	)

	// Login to Kerberos
	err = kerberosClient.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to login to Kerberos: %w", err)
	}

	return &SPNEGOProxyAuth{
		username:       username,
		password:       password,
		proxyURL:       proxyURL,
		kerberosClient: kerberosClient,
	}, nil
}

// configPath returns the path to Kerberos configuration file
func configPath() string {
	// Check KRB5_CONFIG environment variable first (works on all platforms)
	if krb5Config := os.Getenv("KRB5_CONFIG"); krb5Config != "" {
		if _, err := os.Stat(krb5Config); err == nil {
			return krb5Config
		}
	}

	// Platform-specific paths
	if runtime.GOOS == "windows" {
		// Windows paths (check both krb5.conf and krb5.ini)
		paths := []string{
			filepath.Join(os.Getenv("ProgramData"), "Kerberos", "krb5.conf"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.conf"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos", "krb5.conf"),
			filepath.Join(os.Getenv("WINDIR"), "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "Kerberos", "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos", "krb5.ini"),
		}

		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		// Default Windows path
		return filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.conf")
	}

	// Unix/Linux/macOS paths
	paths := []string{
		"/etc/krb5.conf",
		"/usr/local/etc/krb5.conf",
		"/opt/homebrew/etc/krb5.conf",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "/etc/krb5.conf"
}

// Authenticate adds the Proxy-Authorization header with SPNEGO token
func (s *SPNEGOProxyAuth) Authenticate(req *http.Request) error {
	if s.kerberosClient == nil {
		return fmt.Errorf("Kerberos client not initialized")
	}

	// Extract proxy hostname from URL
	var proxyHost string
	if parsedURL, err := url.Parse(s.proxyURL); err == nil && parsedURL.Host != "" {
		if host := parsedURL.Hostname(); host != "" {
			proxyHost = host
		}
	}

	// Validate proxy host was extracted
	if proxyHost == "" {
		return fmt.Errorf("failed to extract proxy hostname from URL: %s", s.proxyURL)
	}

	// Construct SPN for proxy (HTTP/proxy-hostname)
	// Note: gokrb5's SPNEGOClient adds the realm automatically from the krb5Client's realm,
	// so we only need to provide the service/hostname part without the realm
	spn := fmt.Sprintf("HTTP/%s", proxyHost)

	// Generate SPNEGO token for the proxy
	// The service principal name (SPN) for HTTP proxy is typically HTTP/<proxy-host>
	spnegoClient := spnego.SPNEGOClient(s.kerberosClient, spn)

	// Get the SPNEGO token
	token, err := spnegoClient.InitSecContext()
	if err != nil {
		return fmt.Errorf("failed to generate SPNEGO token: %w", err)
	}

	// Marshal the token to bytes and encode as base64
	tokenBytes, err := token.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal SPNEGO token: %w", err)
	}
	encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)

	// Add the Proxy-Authorization header with Negotiate scheme
	req.Header.Set("Proxy-Authorization", "Negotiate "+encodedToken)

	return nil
}
