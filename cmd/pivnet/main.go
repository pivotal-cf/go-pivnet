package main

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
)

func main() {
	if version.Version == "" {
		version.Version = "dev"
	}

	parser := flags.NewParser(&commands.Pivnet, flags.HelpFlag)

	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
}
