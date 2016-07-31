package releasetype

import (
	"io"

	"github.com/olekukonko/tablewriter"
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
	ReleaseTypes() ([]string, error)
}

type ReleaseTypesCommand struct {
}

func (command *ReleaseTypesCommand) Execute([]string) error {
	releaseTypes, err := Client.ReleaseTypes()
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	switch Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ReleaseTypes"})

		for _, r := range releaseTypes {
			table.Append([]string{r})
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(releaseTypes)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(releaseTypes)
	}

	return nil
}
