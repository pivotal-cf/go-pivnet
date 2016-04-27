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

type ReleaseCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

type DeleteReleaseCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleasesCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.Releases.GetByProductSlug(command.ProductSlug)
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

func (command *ReleaseCommand) Execute([]string) error {
	client := NewClient()
	release, err := client.GetRelease(command.ProductSlug, command.ReleaseVersion)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Version", "Description"})

		table.Append([]string{
			strconv.Itoa(release.ID), release.Version, release.Description,
		})
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(release)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(release)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}

func (command *DeleteReleaseCommand) Execute([]string) error {
	client := NewClient()
	release, err := client.GetRelease(command.ProductSlug, command.ReleaseVersion)
	if err != nil {
		return err
	}

	err = client.DeleteRelease(release, command.ProductSlug)
	if err != nil {
		return err
	}

	return nil
}
