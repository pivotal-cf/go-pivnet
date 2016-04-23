package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
)

func init() {
	Pivnet.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}
}
