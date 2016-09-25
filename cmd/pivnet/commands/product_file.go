package commands

import (
	"errors"

	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/productfile"
)

type ProductFilesCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1"`
}

type ProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1"`
	ProductFileID  int    `long:"product-file-id" short:"i" description:"Product file ID e.g. 1234" required:"true"`
}

type AddProductFileCommand struct {
	ProductSlug    string  `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion *string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1"`
	ProductFileID  int     `long:"product-file-id" short:"i" description:"Product file ID e.g. 1234" required:"true"`
	FileGroupID    *int    `long:"file-group-id" short:"f" description:"File group ID e.g. 1234"`
}

type RemoveProductFileCommand struct {
	ProductSlug    string  `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion *string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1"`
	ProductFileID  int     `long:"product-file-id" short:"i" description:"Product file ID e.g. 1234" required:"true"`
	FileGroupID    *int    `long:"file-group-id" short:"f" description:"File group ID e.g. 1234"`
}

type DeleteProductFileCommand struct {
	ProductSlug   string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ProductFileID int    `long:"product-file-id" short:"i" description:"Product file ID e.g. 1234" required:"true"`
}

type DownloadProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" short:"i" description:"Product file ID e.g. 1234" required:"true"`
	Filepath       string `long:"filepath" description:"Local filepath to download file to e.g. /tmp/my-file" required:"true"`
	AcceptEULA     bool   `long:"accept-eula" description:"Automatically accept EULA if necessary"`
}

//go:generate counterfeiter . ProductFileClient
type ProductFileClient interface {
	List(productSlug string, releaseVersion string) error
	Get(productSlug string, releaseVersion string, productFileID int) error
	AddToRelease(productSlug string, releaseVersion string, productFileID int) error
	RemoveFromRelease(productSlug string, releaseVersion string, productFileID int) error
	AddToFileGroup(productSlug string, fileGroupID int, productFileID int) error
	RemoveFromFileGroup(productSlug string, fileGroupID int, productFileID int) error
	Delete(productSlug string, productFileID int) error
	Download(productSlug string, releaseVersion string, productFileID int, filepath string, acceptEULA bool) error
}

var NewProductFileClient = func() ProductFileClient {
	return productfile.NewProductFileClient(
		NewPivnetClient(),
		ErrorHandler,
		Pivnet.Format,
		OutputWriter,
		LogWriter,
		Printer,
		Pivnet.Logger,
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

	if command.ReleaseVersion == nil && command.FileGroupID == nil {
		return errors.New("one of release-version or file-group-id must be provided")
	}
	if command.ReleaseVersion != nil && command.FileGroupID != nil {
		return errors.New("only one of release-version or file-group-id must be provided")
	}

	if command.ReleaseVersion != nil {
		return NewProductFileClient().AddToRelease(command.ProductSlug, *command.ReleaseVersion, command.ProductFileID)
	}

	return NewProductFileClient().AddToFileGroup(command.ProductSlug, *command.FileGroupID, command.ProductFileID)
}

func (command *RemoveProductFileCommand) Execute([]string) error {
	Init()

	if command.ReleaseVersion == nil && command.FileGroupID == nil {
		return errors.New("one of release-version or file-group-id must be provided")
	}
	if command.ReleaseVersion != nil && command.FileGroupID != nil {
		return errors.New("only one of release-version or file-group-id must be provided")
	}

	if command.ReleaseVersion != nil {
		return NewProductFileClient().RemoveFromRelease(command.ProductSlug, *command.ReleaseVersion, command.ProductFileID)
	}

	return NewProductFileClient().RemoveFromFileGroup(command.ProductSlug, *command.FileGroupID, command.ProductFileID)
}

func (command *DeleteProductFileCommand) Execute([]string) error {
	Init()
	return NewProductFileClient().Delete(command.ProductSlug, command.ProductFileID)
}

func (command *DownloadProductFileCommand) Execute([]string) error {
	Init()

	return NewProductFileClient().Download(
		command.ProductSlug,
		command.ReleaseVersion,
		command.ProductFileID,
		command.Filepath,
		command.AcceptEULA,
	)
}
