package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserGroupsService struct {
	client Client
}

type addUserGroupBody struct {
	UserGroup UserGroup `json:"user_group"`
}

type UserGroups struct {
	UserGroups []UserGroup `json:"user_groups,omitempty"`
}

type UserGroup struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (u UserGroupsService) Get(productSlug string, releaseID int) ([]UserGroup, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/user_groups",
		productSlug,
		releaseID,
	)

	var response UserGroups
	err := u.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.UserGroups, nil
}

func (u UserGroupsService) Add(productSlug string, releaseID int, userGroupID int) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/add_user_group",
		productSlug,
		releaseID,
	)

	body := addUserGroupBody{
		UserGroup: UserGroup{
			ID: userGroupID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	err = u.client.makeRequest(
		"PATCH",
		url,
		http.StatusNoContent,
		bytes.NewReader(b),
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
