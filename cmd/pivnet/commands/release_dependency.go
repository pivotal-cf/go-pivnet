package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type ReleaseDependenciesCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleaseDependenciesCommand) Execute([]string) error {
	client := NewClient()

	product, err := client.FindProductForSlug(command.ProductSlug)
	if err != nil {
		return err
	}

	release, err := client.GetRelease(command.ProductSlug, command.ReleaseVersion)
	if err != nil {
		return err
	}

	releaseDependencies, err := client.ReleaseDependencies(product.ID, release.ID)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"ID",
			"Version",
			"Description",
			"Product ID",
			"Product Slug",
		})

		for _, r := range releaseDependencies {
			table.Append([]string{
				strconv.Itoa(r.Release.ID),
				r.Release.Version,
				strconv.Itoa(r.Release.Product.ID),
				r.Release.Product.Slug,
			})
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(releaseDependencies)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(releaseDependencies)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
