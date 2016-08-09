package commands

import "github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/productfile"

type ProductFilesCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1"`
}

type ProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type AddProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type RemoveProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type DeleteProductFileCommand struct {
	ProductSlug   string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ProductFileID int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

//go:generate counterfeiter . ProductFileClient
type ProductFileClient interface {
	List(productSlug string, releaseVersion string) error
	Get(productSlug string, releaseVersion string, productFileID int) error
	AddToRelease(productSlug string, releaseVersion string, productFileID int) error
	RemoveFromRelease(productSlug string, releaseVersion string, productFileID int) error
	Delete(productSlug string, productFileID int) error
}

var NewProductFileClient = func() ProductFileClient {
	return productfile.NewProductFileClient(
		NewPivnetClient(),
		ErrorHandler,
		Pivnet.Format,
		OutputWriter,
		Printer,
	)
}

func (command *ProductFilesCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().List(command.ProductSlug, command.ReleaseVersion)
}

func (command *ProductFileCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().Get(command.ProductSlug, command.ReleaseVersion, command.ProductFileID)
}

func (command *AddProductFileCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().AddToRelease(command.ProductSlug, command.ReleaseVersion, command.ProductFileID)
}

func (command *RemoveProductFileCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().RemoveFromRelease(command.ProductSlug, command.ReleaseVersion, command.ProductFileID)
}

func (command *DeleteProductFileCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().Delete(command.ProductSlug, command.ProductFileID)
}
