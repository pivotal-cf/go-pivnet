package pivnet

import (
	"fmt"
	"net/http"
)

type FileGroupsService struct {
	client Client
}

type FileGroup struct {
	ID           int              `json:"id,omitempty"`
	Name         string           `json:"name,omitempty"`
	Product      FileGroupProduct `json:"product,omitempty"`
	ProductFiles []ProductFile    `json:"product_files,omitempty"`
}

type FileGroupProduct struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type FileGroupsResponse struct {
	FileGroups []FileGroup `json:"file_groups,omitempty"`
}

func (e FileGroupsService) List(productSlug string) ([]FileGroup, error) {
	url := fmt.Sprintf("/products/%s/file_groups", productSlug)

	var response FileGroupsResponse
	err := e.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.FileGroups, nil
}

func (p FileGroupsService) Get(productSlug string, releaseID int, fileGroupID int) (FileGroup, error) {
	url := fmt.Sprintf("/products/%s/releases/%d/file_groups/%d",
		productSlug,
		releaseID,
		fileGroupID,
	)
	response := FileGroup{}

	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return FileGroup{}, err
	}

	return response, nil
}

func (p FileGroupsService) Delete(productSlug string, id int) (FileGroup, error) {
	url := fmt.Sprintf(
		"/products/%s/file_groups/%d",
		productSlug,
		id,
	)

	var response FileGroup
	err := p.client.makeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return FileGroup{}, err
	}

	return response, nil
}

func (p FileGroupsService) ListForRelease(productSlug string, releaseID int) ([]FileGroup, error) {
	url := fmt.Sprintf("/products/%s/releases/%d/file_groups",
		productSlug,
		releaseID,
	)
	response := FileGroupsResponse{}

	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return []FileGroup{}, err
	}

	return response.FileGroups, nil
}
