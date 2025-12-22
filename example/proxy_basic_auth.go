package main

import (
	"fmt"
	"log"
	"os"

	pivnet "github.com/pivotal-cf/go-pivnet/v9"
	"github.com/pivotal-cf/go-pivnet/v9/logshim"
)

// Example demonstrating HTTP Basic proxy authentication
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
		log.Fatal("PROXY_USERNAME environment variable is required")
	}

	proxyPassword := os.Getenv("PROXY_PASSWORD")
	if proxyPassword == "" {
		log.Fatal("PROXY_PASSWORD environment variable is required")
	}

	// Create a logger
	logger := logshim.NewLogShim(
		log.New(os.Stdout, "", log.LstdFlags),
		log.New(os.Stderr, "", log.LstdFlags),
		true,
	)

	// Configure the client with Basic proxy authentication
	config := pivnet.ClientConfig{
		Host:              pivnet.DefaultHost,
		UserAgent:         "pivnet-example-basic-proxy/1.0",
		SkipSSLValidation: false,
		ProxyAuthConfig: pivnet.ProxyAuthConfig{
			ProxyURL: proxyURL,
			AuthType: pivnet.ProxyAuthTypeBasic,
			Username: proxyUsername,
			Password: proxyPassword,
		},
	}

	// Create access token with proxy auth config
	token := pivnet.NewAccessTokenOrLegacyTokenWithProxy(apiToken, config.Host, config.SkipSSLValidation, config.ProxyAuthConfig)

	// Create the client with proxy support
	client, err := pivnet.NewClientWithProxy(token, config, logger)
	if err != nil {
		log.Fatalf("Failed to create client with proxy: %v", err)
	}

	// Example: List all products
	fmt.Println("Fetching products through authenticated proxy...")
	products, err := client.Products.List()
	if err != nil {
		log.Fatalf("Failed to list products: %v", err)
	}

	fmt.Printf("\nSuccessfully retrieved %d products via Basic proxy authentication!\n\n", len(products))

	// Display first 5 products
	fmt.Println("First 5 products:")
	for i, product := range products {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s (slug: %s)\n", i+1, product.Name, product.Slug)
	}

	// Example: Get details for a specific product
	if len(products) > 0 {
		fmt.Printf("\nFetching details for product: %s\n", products[0].Slug)
		product, err := client.Products.Get(products[0].Slug)
		if err != nil {
			log.Fatalf("Failed to get product details: %v", err)
		}
		fmt.Printf("Product Name: %s\n", product.Name)
		fmt.Printf("Product ID: %d\n", product.ID)
	}

	fmt.Println("\n✅ Basic proxy authentication example completed successfully!")
}
