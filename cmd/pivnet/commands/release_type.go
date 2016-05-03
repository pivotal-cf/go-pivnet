package commands

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type ReleaseTypesCommand struct {
}

func (command *ReleaseTypesCommand) Execute([]string) error {
	client := NewClient()
	releaseTypes, err := client.ReleaseTypes.Get()
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutWriter)
		table.SetHeader([]string{"ReleaseTypes"})

		for _, r := range releaseTypes {
			table.Append([]string{r})
		}
		table.Render()
		return nil
	case PrintAsJSON:
		b, err := json.Marshal(releaseTypes)
		if err != nil {
			return err
		}

		OutWriter.Write(b)
		return nil
	case PrintAsYAML:
		b, err := yaml.Marshal(releaseTypes)
		if err != nil {
			return err
		}

		output := fmt.Sprintf("---\n%s\n", string(b))
		OutWriter.Write([]byte(output))
		return nil
	}

	return nil
}
