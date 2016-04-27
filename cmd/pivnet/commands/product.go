package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type ProductCommand struct {
	Slug string `short:"s" long:"slug" description:"Product slug e.g. p-mysql" required:"true"`
}

type ProductsCommand struct {
}

func (command *ProductCommand) Execute([]string) error {
	client := NewClient()
	product, err := client.Products.Get(command.Slug)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		productAsString := []string{
			strconv.Itoa(product.ID), product.Slug, product.Name,
		}
		table.Append(productAsString)
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(product)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(product)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}

func (command *ProductsCommand) Execute([]string) error {
	client := NewClient()
	products, err := client.Products.List()
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		for _, product := range products {
			productAsString := []string{
				strconv.Itoa(product.ID), product.Slug, product.Name,
			}
			table.Append(productAsString)
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(products)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(products)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
