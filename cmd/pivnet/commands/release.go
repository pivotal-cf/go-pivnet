package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type ReleasesCommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
}

func (command *ReleasesCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.ReleasesForProductSlug(command.ProductSlug)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Version", "Description"})

		for _, r := range releases {
			table.Append([]string{
				strconv.Itoa(r.ID), r.Version, r.Description,
			})
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(releases)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(releases)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
