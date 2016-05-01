package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type FileGroupsCommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
}

func (command *FileGroupsCommand) Execute([]string) error {
	client := NewClient()

	fileGroups, err := client.FileGroups.List(
		command.ProductSlug,
	)
	if err != nil {
		return err
	}

	return printFileGroups(fileGroups)
}

func printFileGroups(fileGroups []pivnet.FileGroup) error {
	switch Pivnet.Format {

	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"ID",
			"Name",
			"Product Name",
			"Product File Names",
		})

		for _, fileGroup := range fileGroups {
			var productFileNames []string

			for _, productFile := range fileGroup.ProductFiles {
				productFileNames = append(productFileNames, productFile.Name)
			}

			fileGroupAsString := []string{
				strconv.Itoa(fileGroup.ID),
				fileGroup.Name,
				fileGroup.Product.Name,
				strings.Join(productFileNames, " "),
			}
			table.Append(fileGroupAsString)
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(fileGroups)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(fileGroups)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
