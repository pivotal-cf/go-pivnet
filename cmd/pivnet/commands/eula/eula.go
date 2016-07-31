package eula

import (
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var (
	Client       PivnetClient
	ErrorHandler errorhandler.ErrorHandler
	Format       string
	OutputWriter io.Writer
	Printer      printer.Printer
)

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	AcceptEULA(productSlug string, releaseID int) error
	EULAs() ([]pivnet.EULA, error)
	EULA(eulaSlug string) (pivnet.EULA, error)
	ReleaseForProductVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
}

type EULAsCommand struct {
}

type EULACommand struct {
	EULASlug string
}

type AcceptEULACommand struct {
	ProductSlug    string
	ReleaseVersion string
}

func (command *EULAsCommand) Execute([]string) error {
	eulas, err := Client.EULAs()
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printEULAs(eulas)
}

func printEULA(eula pivnet.EULA) error {
	switch Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		eulaAsString := []string{
			strconv.Itoa(eula.ID), eula.Slug, eula.Name,
		}
		table.Append(eulaAsString)
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(eula)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(eula)
	}

	return nil
}

func (command *EULACommand) Execute([]string) error {
	eula, err := Client.EULA(command.EULASlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printEULA(eula)
}

func printEULAs(eulas []pivnet.EULA) error {
	switch Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		for _, e := range eulas {
			eulaAsString := []string{
				strconv.Itoa(e.ID), e.Slug, e.Name,
			}
			table.Append(eulaAsString)
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(eulas)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(eulas)
	}

	return nil
}

func (command *AcceptEULACommand) Execute([]string) error {
	release, err := Client.ReleaseForProductVersion(command.ProductSlug, command.ReleaseVersion)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	err = Client.AcceptEULA(command.ProductSlug, release.ID)

	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	if Format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"eula acccepted successfully for %s/%s\n",
			command.ProductSlug,
			command.ReleaseVersion,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
