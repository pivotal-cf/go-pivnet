package commands

import "fmt"

type ProductCommand struct {
	Slug string `short:"s" long:"slug" description:"Product slug e.g. p-mysql" required:"true"`
}

func (command *ProductCommand) Execute([]string) error {
	client := NewClient()
	product, err := client.FindProductForSlug(command.Slug)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", product)

	return nil
}
