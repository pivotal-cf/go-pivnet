package main

import (
	"fmt"
	"log"
	"os"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
)

// Example demonstrating Kerberos/SPNEGO proxy authentication
func main() {
	// Get configuration from environment variables
	apiToken := os.Getenv("PIVNET_API_TOKEN")
	if apiToken == "" {
		log.Fatal("PIVNET_API_TOKEN environment variable is required")
	}

	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL == "" {
		log.Fatal("PROXY_URL environment variable is required (e.g., http://proxy.example.com:8080)")
	}

	proxyUsername := os.Getenv("PROXY_USERNAME")
	if proxyUsername == "" {
		log.Fatal("PROXY_USERNAME environment variable is required (e.g., user@REALM.COM)")
	}

	proxyPassword := os.Getenv("PROXY_PASSWORD")
	if proxyPassword == "" {
		log.Fatal("PROXY_PASSWORD environment variable is required")
	}

	// Optional: Path to custom Kerberos config file
	krb5ConfigPath := os.Getenv("KRB5_CONFIG")

	// Create a logger
	logger := logshim.NewLogShim(
		log.New(os.Stdout, "", log.LstdFlags),
		log.New(os.Stderr, "", log.LstdFlags),
		true,
	)

	// Configure the client with SPNEGO proxy authentication
	config := pivnet.ClientConfig{
		Host:              pivnet.DefaultHost,
		UserAgent:         "pivnet-example-spnego-proxy/1.0",
		SkipSSLValidation: false,
		ProxyAuthConfig: pivnet.ProxyAuthConfig{
			ProxyURL:   proxyURL,
			AuthType:   pivnet.ProxyAuthTypeSPNEGO,
			Username:   proxyUsername,
			Password:   proxyPassword,
			Krb5Config: krb5ConfigPath, // Optional: empty string uses default config
		},
	}

	// Create access token with proxy auth config
	token := pivnet.NewAccessTokenOrLegacyTokenWithProxy(apiToken, config.Host, config.SkipSSLValidation, config.ProxyAuthConfig)

	// Create the client with proxy support
	fmt.Println("Initializing client with SPNEGO proxy authentication...")
	fmt.Printf("Proxy URL: %s\n", proxyURL)
	fmt.Printf("Username: %s\n", proxyUsername)
	if krb5ConfigPath != "" {
		fmt.Printf("Kerberos Config: %s\n", krb5ConfigPath)
	} else {
		fmt.Println("Kerberos Config: Using default config")
	}

	client, err := pivnet.NewClientWithProxy(token, config, logger)
	if err != nil {
		log.Fatalf("Failed to create client with proxy: %v", err)
	}

	// Example: List all products
	fmt.Println("\nFetching products through SPNEGO authenticated proxy...")
	products, err := client.Products.List()
	if err != nil {
		log.Fatalf("Failed to list products: %v", err)
	}

	fmt.Printf("\nSuccessfully retrieved %d products via SPNEGO proxy authentication!\n\n", len(products))

	// Display first 5 products
	fmt.Println("First 5 products:")
	for i, product := range products {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s (slug: %s)\n", i+1, product.Name, product.Slug)
	}

	// Example: Get releases for a product
	if len(products) > 0 {
		productSlug := products[0].Slug
		fmt.Printf("\nFetching releases for product: %s\n", productSlug)
		releases, err := client.Releases.List(productSlug)
		if err != nil {
			log.Fatalf("Failed to list releases: %v", err)
		}
		fmt.Printf("Found %d releases\n", len(releases))

		// Display first 3 releases
		if len(releases) > 0 {
			fmt.Println("\nFirst 3 releases:")
			for i, release := range releases {
				if i >= 3 {
					break
				}
				fmt.Printf("  %d. Version %s (ID: %d)\n", i+1, release.Version, release.ID)
			}
		}
	}

	fmt.Println("\n✅ SPNEGO proxy authentication example completed successfully!")
}
