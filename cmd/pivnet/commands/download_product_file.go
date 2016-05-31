package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
)

type DownloadProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
	Filepath       string `long:"filepath" description:"Local filepath to download file to e.g. /tmp/my-file" required:"true"`
}

func (command *DownloadProductFileCommand) Execute([]string) error {
	client := NewClient()

	if command.ReleaseVersion == "" {
		productFiles, err := client.ProductFiles.List(
			command.ProductSlug,
		)
		if err != nil {
			return ErrorHandler.HandleError(err)
		}

		return printProductFiles(productFiles)
	}

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	extendedClient := extension.NewExtendedClient(client, Pivnet.Logger)

	downloadLink := fmt.Sprintf(
		"/products/%s/releases/%d/product_files/%d/download",
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)

	err = extendedClient.DownloadFile(command.Filepath, downloadLink)
	if err != nil {
		return err
		// return ErrorHandler.HandleError(err)
	}

	return nil
}
