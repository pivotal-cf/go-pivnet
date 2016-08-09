package productfile

import (
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	ReleaseForProductVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	GetProductFiles(productSlug string) ([]pivnet.ProductFile, error)
	GetProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	GetProductFile(productSlug string, productFileID int) (pivnet.ProductFile, error)
	GetProductFileForRelease(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error)
	AddProductFile(productSlug string, releaseID int, productFileID int) error
	RemoveProductFile(productSlug string, releaseID int, productFileID int) error
	DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error)
}

type ProductFileClient struct {
	pivnetClient PivnetClient
	eh           errorhandler.ErrorHandler
	format       string
	outputWriter io.Writer
	printer      printer.Printer
}

func NewProductFileClient(
	pivnetClient PivnetClient,
	eh errorhandler.ErrorHandler,
	format string,
	outputWriter io.Writer,
	printer printer.Printer,
) *ProductFileClient {
	return &ProductFileClient{
		pivnetClient: pivnetClient,
		eh:           eh,
		format:       format,
		outputWriter: outputWriter,
		printer:      printer,
	}
}

func (c *ProductFileClient) List(productSlug string, releaseVersion string) error {
	if releaseVersion == "" {
		productFiles, err := c.pivnetClient.GetProductFiles(productSlug)
		if err != nil {
			return c.eh.HandleError(err)
		}

		return c.printProductFiles(productFiles)
	}

	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	productFiles, err := c.pivnetClient.GetProductFilesForRelease(
		productSlug,
		release.ID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	return c.printProductFiles(productFiles)
}

func (c *ProductFileClient) printProductFiles(productFiles []pivnet.ProductFile) error {
	switch c.format {

	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.outputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"File Version",
			"AWS Object Key",
		})

		for _, productFile := range productFiles {
			productFileAsString := []string{
				strconv.Itoa(productFile.ID),
				productFile.Name,
				productFile.FileVersion,
				productFile.AWSObjectKey,
			}
			table.Append(productFileAsString)
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return c.printer.PrintJSON(productFiles)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(productFiles)
	}

	return nil
}

func (c *ProductFileClient) printProductFile(productFile pivnet.ProductFile) error {
	switch c.format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.outputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"File Version",
			"File Type",
			"Description",
			"MD5",
			"AWS Object Key",
			"Size (Bytes)",
		})

		productFileAsString := []string{
			strconv.Itoa(productFile.ID),
			productFile.Name,
			productFile.FileVersion,
			productFile.FileType,
			productFile.Description,
			productFile.MD5,
			productFile.AWSObjectKey,
			fmt.Sprintf("%d", productFile.Size),
		}
		table.Append(productFileAsString)
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return c.printer.PrintJSON(productFile)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(productFile)
	}

	return nil
}

func (c *ProductFileClient) Get(
	productSlug string,
	releaseVersion string,
	productFileID int,
) error {
	if releaseVersion == "" {
		productFile, err := c.pivnetClient.GetProductFile(
			productSlug,
			productFileID,
		)
		if err != nil {
			return c.eh.HandleError(err)
		}
		return c.printProductFile(productFile)
	}

	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	productFile, err := c.pivnetClient.GetProductFileForRelease(
		productSlug,
		release.ID,
		productFileID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	return c.printProductFile(productFile)
}

func (c *ProductFileClient) AddToRelease(
	productSlug string,
	releaseVersion string,
	productFileID int,
) error {
	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	err = c.pivnetClient.AddProductFile(
		productSlug,
		release.ID,
		productFileID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	if c.format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			c.outputWriter,
			"product file %d added successfully to %s/%s\n",
			productFileID,
			productSlug,
			releaseVersion,
		)
	}

	return nil
}

func (c *ProductFileClient) RemoveFromRelease(
	productSlug string,
	releaseVersion string,
	productFileID int,
) error {
	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	err = c.pivnetClient.RemoveProductFile(
		productSlug,
		release.ID,
		productFileID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	if c.format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			c.outputWriter,
			"product file %d removed successfully from %s/%s\n",
			productFileID,
			productSlug,
			releaseVersion,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ProductFileClient) Delete(productSlug string, productFileID int) error {
	productFile, err := c.pivnetClient.DeleteProductFile(
		productSlug,
		productFileID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	if c.format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			c.outputWriter,
			"product file %d deleted successfully for %s\n",
			productFileID,
			productSlug,
		)
	}

	return c.printProductFile(productFile)
}
