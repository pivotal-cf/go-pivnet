package main

import (
	"fmt"
	"log"
	"os"

	pivnet "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logshim"
)

func main() {
	config := pivnet.ClientConfig{
		Host:      pivnet.DefaultHost,
		UserAgent: "pivnet-cli-example",
		SkipSSLValidation: true,
	}

	accessTokenService := pivnet.NewAccessTokenOrLegacyToken("token-from-pivnet", config.Host, config.SkipSSLValidation)

	stdoutLogger := log.New(os.Stdout, "", log.LstdFlags)
	stderrLogger := log.New(os.Stderr, "", log.LstdFlags)

	verbose := false
	logger := logshim.NewLogShim(stdoutLogger, stderrLogger, verbose)

	client := pivnet.NewClient(accessTokenService, config, logger)

	products, err := client.Products.List()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("products: %v", products)
}
