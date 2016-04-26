package pivnet

import (
	"fmt"
	"net/http"
)

type ReleaseDependenciesService struct {
	client Client
}

type ReleaseDependenciesResponse struct {
	ReleaseDependencies []ReleaseDependency `json:"dependencies,omitempty"`
}

type ReleaseDependency struct {
	Release DependentRelease `json:"release,omitempty"`
}

type DependentRelease struct {
	ID      int     `json:"id,omitempty"`
	Version string  `json:"version,omitempty"`
	Product Product `json:"product,omitempty"`
}

func (r ReleaseDependenciesService) Get(productID int, releaseID int) ([]ReleaseDependency, error) {
	url := fmt.Sprintf(
		"/products/%d/releases/%d/dependencies",
		productID,
		releaseID,
	)

	var response ReleaseDependenciesResponse
	err := r.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.ReleaseDependencies, nil
}
