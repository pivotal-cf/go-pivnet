package pivnet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type CompanyGroupsService struct {
	client Client
}

type CompanyGroupsResponse struct {
	CompanyGroups []CompanyGroup `json:"company_groups,omitempty"`
}

type CompanyGroupMember struct {
	ID      int    `json:"id,omitempty" yaml:"id,omitempty"`
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Email   string `json:"email,omitempty" yaml:"email,omitempty"`
	IsAdmin bool   `json:"admin" yaml:"admin"`
}

type CompanyGroupMemberNoAdmin struct {
	ID    int    `json:"id,omitempty" yaml:"id,omitempty"`
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type CompanyGroupMemberEmail struct {
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type companyGroupMemberToAdd struct {
	Member CompanyGroupMember `json:"member,omitempty"`
}

type companyGroupMemberNoAdminToAdd struct {
	Member CompanyGroupMemberNoAdmin `json:"member,omitempty"`
}

type companyGroupMemberToRemove struct {
	Member CompanyGroupMemberEmail `json:"member"`
}

type CompanyGroup struct {
	ID                 int                       `json:"id,omitempty" yaml:"id,omitempty"`
	Name               string                    `json:"name,omitempty" yaml:"name,omitempty"`
	Members            []CompanyGroupMember      `json:"members,omitempty" yaml:"members,omitempty"`
	PendingInvitations []string                  `json:"pending_invitations,omitempty" yaml:"pending_invitations,omitempty"`
	Entitlements       []CompanyGroupEntitlement `json:"entitlements,omitempty" yaml:"entitlements,omitempty"`
}

type CompanyGroupEntitlement struct {
	ID   int    `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (c CompanyGroupsService) List() ([]CompanyGroup, error) {
	url := "/company_groups"

	var response CompanyGroupsResponse
	resp, err := c.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.CompanyGroups, nil
}

func (c CompanyGroupsService) Get(companyGroupID int) (CompanyGroup, error) {
	url := fmt.Sprintf("/company_groups/%d", companyGroupID)

	var response CompanyGroup
	resp, err := c.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return CompanyGroup{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return CompanyGroup{}, err
	}

	return response, nil
}

func (c CompanyGroupsService) AddMember(
	companyGroupID int,
	memberEmailAddress string,
	isAdmin string,
) (CompanyGroup, error) {
	url := fmt.Sprintf("/company_groups/%d/add_member", companyGroupID)

	var b []byte
	var err error

	if len(strings.TrimSpace(isAdmin)) == 0 {
		addCompanyGroupMemberBody := companyGroupMemberNoAdminToAdd{
			CompanyGroupMemberNoAdmin{
				Email: memberEmailAddress,
			},
		}

		b, err = json.Marshal(addCompanyGroupMemberBody)
		if err != nil {
			return CompanyGroup{}, err
		}
	} else {
		isAdmin, err := strconv.ParseBool(isAdmin)
		if err != nil {
			return CompanyGroup{}, errors.New("parameter admin should be true or false")
		}

		addCompanyGroupMemberBody := companyGroupMemberToAdd{
			CompanyGroupMember{
				Email:   memberEmailAddress,
				IsAdmin: isAdmin,
			},
		}

		b, err = json.Marshal(addCompanyGroupMemberBody)
		if err != nil {
			return CompanyGroup{}, err
		}
	}

	body := bytes.NewReader(b)

	var response CompanyGroup
	resp, err := c.client.MakeRequest(
		"PATCH",
		url,
		http.StatusOK,
		body,
	)
	if err != nil {
		return CompanyGroup{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return CompanyGroup{}, err
	}

	return response, nil
}

func (c CompanyGroupsService) RemoveMember(
	companyGroupID int,
	memberEmailAddress string,
) (CompanyGroup, error) {
	url := fmt.Sprintf("/company_groups/%d/remove_member", companyGroupID)

	addCompanyGroupMemberBody := companyGroupMemberToRemove{
		CompanyGroupMemberEmail{
			Email: memberEmailAddress,
		},
	}

	b, err := json.Marshal(addCompanyGroupMemberBody)
	if err != nil {
		return CompanyGroup{}, err
	}

	body := bytes.NewReader(b)

	var response CompanyGroup
	resp, err := c.client.MakeRequest(
		"PATCH",
		url,
		http.StatusOK,
		body,
	)
	if err != nil {
		return CompanyGroup{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return CompanyGroup{}, err
	}

	return response, nil
}
