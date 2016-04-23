package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
	"github.com/pivotal-golang/lager"
)

type PivnetCommand struct {
	Version func() `short:"v" long:"version" description:"Print the version of Pivnet and exit"`

	APIToken string `long:"api-token" description:"Pivnet API token"`
	Endpoint string `long:"endpoint" description:"Pivnet API Endpoint"`

	Products ProductCommand `command:"product" description:"List products"`
}

var Pivnet PivnetCommand

func init() {
	Pivnet.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if Pivnet.Endpoint == "" {
		Pivnet.Endpoint = pivnet.Endpoint
	}
}

func NewClient() pivnet.Client {
	useragent := fmt.Sprintf(
		"go-pivnet/%s",
		version.Version,
	)

	pivnetClient := pivnet.NewClient(
		pivnet.ClientConfig{
			Token:     Pivnet.APIToken,
			Endpoint:  Pivnet.Endpoint,
			UserAgent: useragent,
		},
		lager.NewLogger("pivnet CLI"),
	)

	return pivnetClient
}
