package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type UserGroupsCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1"`
}

type CreateUserGroupCommand struct {
	Name        string   `long:"name" description:"Name e.g. all_users" required:"true"`
	Description string   `long:"description" description:"Description e.g. 'All users in the world'" required:"true"`
	Members     []string `long:"member" description:"Email addresses of members to be added"`
}

type DeleteUserGroupCommand struct {
	UserGroupID int `long:"user-group-id" description:"User group ID e.g. 1234" required:"true"`
}

func (command *UserGroupsCommand) Execute([]string) error {
	client := NewClient()

	if command.ProductSlug == "" && command.ReleaseVersion == "" {
		var err error
		userGroups, err := client.UserGroups.List()
		if err != nil {
			return err
		}

		return printUserGroups(userGroups)
	}

	if command.ProductSlug == "" || command.ReleaseVersion == "" {
		return fmt.Errorf("Both or neither of product slug and release version must be provided")
	}

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	userGroups, err := client.UserGroups.ListForRelease(command.ProductSlug, release.ID)
	if err != nil {
		return err
	}

	return printUserGroups(userGroups)
}

func printUserGroups(userGroups []pivnet.UserGroup) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutWriter)
		table.SetHeader([]string{"ID", "Name", "Description"})

		for _, u := range userGroups {
			table.Append([]string{
				strconv.Itoa(u.ID),
				u.Name,
				u.Description,
			})
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(userGroups)
	case PrintAsYAML:
		return printYAML(userGroups)
	}

	return nil
}

func (command *CreateUserGroupCommand) Execute([]string) error {
	client := NewClient()

	userGroup, err := client.UserGroups.Create(command.Name, command.Description, command.Members)
	if err != nil {
		return err
	}

	return printUserGroup(userGroup)
}

func printUserGroup(userGroup pivnet.UserGroup) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutWriter)
		table.SetHeader([]string{"ID", "Name", "Description", "Members"})

		table.Append([]string{
			strconv.Itoa(userGroup.ID),
			userGroup.Name,
			userGroup.Description,
			strings.Join(userGroup.Members, ""),
		})

		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(userGroup)
	case PrintAsYAML:
		return printYAML(userGroup)
	}

	return nil
}

func (command *DeleteUserGroupCommand) Execute([]string) error {
	client := NewClient()

	err := client.UserGroups.Delete(command.UserGroupID)
	if err != nil {
		return err
	}

	return nil
}
