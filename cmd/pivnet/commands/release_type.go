package commands

import "github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasetype"

type ReleaseTypesCommand struct {
}

func initReleasetypePackage() {
	Init()
	releasetype.Client = NewPivnetClient()
	releasetype.ErrorHandler = ErrorHandler
	releasetype.Format = Pivnet.Format
	releasetype.OutputWriter = OutputWriter
	releasetype.Printer = Printer
}

func (command *ReleaseTypesCommand) Execute(args []string) error {
	initReleasetypePackage()

	c := &releasetype.ReleaseTypesCommand{}
	return c.Execute(args)
}
