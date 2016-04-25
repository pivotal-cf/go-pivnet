package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type UserGroupsCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *UserGroupsCommand) Execute([]string) error {
	client := NewClient()
	release, err := client.GetRelease(command.ProductSlug, command.ReleaseVersion)
	if err != nil {
		return err
	}

	userGroups, err := client.UserGroups(command.ProductSlug, release.ID)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description"})

		for _, u := range userGroups {
			table.Append([]string{
				strconv.Itoa(u.ID), u.Name, u.Description,
			})
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(userGroups)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(userGroups)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
