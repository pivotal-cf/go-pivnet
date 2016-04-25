package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type EULAsCommand struct {
}

type AcceptEULACommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseID   int    `long:"release-id" description:"Release ID e.g. 1234" required:"true"`
}

func (command *EULAsCommand) Execute([]string) error {
	client := NewClient()
	eulas, err := client.EULAs()
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		for _, e := range eulas {
			eulaAsString := []string{
				strconv.Itoa(e.ID), e.Slug, e.Name,
			}
			table.Append(eulaAsString)
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(eulas)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(eulas)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}

func (command *AcceptEULACommand) Execute([]string) error {
	client := NewClient()
	err := client.AcceptEULA(command.ProductSlug, command.ReleaseID)
	if err != nil {
		return err
	}

	return nil
}
