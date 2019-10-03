package pivnet

import (
	"encoding/json"
	"net/http"
)

type CompanyGroupsService struct {
	client Client
}

type CompanyGroupsResponse struct {
	CompanyGroups []CompanyGroup `json:"company_groups,omitempty"`
}

type CompanyGroup struct {
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
