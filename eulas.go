package pivnet

import (
	"fmt"
	"net/http"
	"strings"
)

type EULAService struct {
	client Client
}

type EULA struct {
	Slug    string `json:"slug,omitempty"`
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	Links   *Links `json:"_links,omitempty"`
}

type EULAsResponse struct {
	EULAs []EULA `json:"eulas,omitempty"`
	Links *Links `json:"_links,omitempty"`
}

type EULAAcceptanceResponse struct {
	AcceptedAt string `json:"accepted_at,omitempty"`
	Links      *Links `json:"_links,omitempty"`
}

func (e EULAService) List() ([]EULA, error) {
	url := "/eulas"

	var response EULAsResponse
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

	return response.EULAs, nil
}

func (e EULAService) Accept(productSlug string, releaseID int) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/eula_acceptance",
		productSlug,
		releaseID,
	)

	var response EULAAcceptanceResponse
	err := e.client.makeRequest(
		"POST",
		url,
		http.StatusOK,
		strings.NewReader(`{}`),
		&response,
	)
	if err != nil {
		return err
	}

	return nil
}
