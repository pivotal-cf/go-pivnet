package commands

import "github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/eula"

type EULAsCommand struct {
}

type EULACommand struct {
	EULASlug string `long:"eula-slug" description:"EULA slug e.g. pivotal_software_eula" required:"true"`
}

type AcceptEULACommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func initPackage() {
	NewClient()
	eula.Client = NewPivnetClient()
	eula.ErrorHandler = ErrorHandler
	eula.Format = Pivnet.Format
	eula.OutputWriter = OutputWriter
	eula.Printer = Printer
}

func (command *EULAsCommand) Execute(args []string) error {
	initPackage()
	c := &eula.EULAsCommand{}
	return c.Execute(args)
}

func (command *EULACommand) Execute(args []string) error {
	initPackage()
	c := &eula.EULACommand{EULASlug: command.EULASlug}
	return c.Execute(args)
}

func (command *AcceptEULACommand) Execute(args []string) error {
	initPackage()
	c := &eula.AcceptEULACommand{
		ProductSlug:    command.ProductSlug,
		ReleaseVersion: command.ReleaseVersion,
	}
	return c.Execute(args)
}
