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

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	AcceptEULA(productSlug string, releaseID int) error
	EULAs() ([]pivnet.EULA, error)
	EULA(eulaSlug string) (pivnet.EULA, error)
	ReleaseForProductVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
}

type EULAs struct {
	Client       PivnetClient
	ErrorHandler errorhandler.ErrorHandler
	Format       string
	OutputWriter io.Writer
	Printer      printer.Printer
}

func (c *EULAs) List([]string) error {
	eulas, err := c.Client.EULAs()
	if err != nil {
		return c.ErrorHandler.HandleError(err)
	}

	return c.printEULAs(eulas)
}

func (c *EULAs) printEULA(eula pivnet.EULA) error {
	switch c.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		eulaAsString := []string{
			strconv.Itoa(eula.ID), eula.Slug, eula.Name,
		}
		table.Append(eulaAsString)
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return c.Printer.PrintJSON(eula)
	case printer.PrintAsYAML:
		return c.Printer.PrintYAML(eula)
	}

	return nil
}

func (c *EULAs) Get(eulaSlug string) error {
	eula, err := c.Client.EULA(eulaSlug)
	if err != nil {
		return c.ErrorHandler.HandleError(err)
	}

	return c.printEULA(eula)
}

func (c *EULAs) printEULAs(eulas []pivnet.EULA) error {
	switch c.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.OutputWriter)
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
		return c.Printer.PrintJSON(eulas)
	case printer.PrintAsYAML:
		return c.Printer.PrintYAML(eulas)
	}

	return nil
}

func (c *EULAs) AcceptEULA(productSlug string, releaseVersion string) error {
	release, err := c.Client.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.ErrorHandler.HandleError(err)
	}

	err = c.Client.AcceptEULA(productSlug, release.ID)

	if err != nil {
		return c.ErrorHandler.HandleError(err)
	}

	if c.Format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			c.OutputWriter,
			"eula acccepted successfully for %s/%s\n",
			productSlug,
			releaseVersion,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
