package commands

type PivnetCommand struct {
	Version func() `short:"v" long:"version" description:"Print the version of Pivnet and exit"`
}

var Pivnet PivnetCommand
