package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HelmChartReferencesService struct {
	client Client
}

type CreateHelmChartReferenceConfig struct {
	ProductSlug        string
	Description        string
	DocsURL            string
	Name               string
	Version            string
	SystemRequirements []string
}

type HelmChartReferencesResponse struct {
	HelmChartReferences []HelmChartReference `json:"helm_chart_references,omitempty"`
}

type HelmChartReferenceResponse struct {
	HelmChartReference HelmChartReference `json:"helm_chart_reference,omitempty"`
}

type HelmChartReference struct {
	ID                 int               `json:"id,omitempty" yaml:"id,omitempty"`
	Description        string            `json:"description,omitempty" yaml:"description,omitempty"`
	DocsURL            string            `json:"docs_url,omitempty" yaml:"docs_url,omitempty"`
	Name               string            `json:"name,omitempty" yaml:"name,omitempty"`
	Version            string            `json:"version,omitempty" yaml:"version,omitempty"`
	SystemRequirements []string          `json:"system_requirements,omitempty" yaml:"system_requirements,omitempty"`
	ReplicationStatus  ReplicationStatus `json:"replication_status,omitempty" yaml:"replication_status,omitempty"`
}

type createUpdateHelmChartReferenceBody struct {
	HelmChartReference HelmChartReference `json:"helm_chart_reference"`
}

func (p HelmChartReferencesService) List(productSlug string) ([]HelmChartReference, error) {
	url := fmt.Sprintf("/products/%s/helm_chart_references", productSlug)

	var response HelmChartReferencesResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return []HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []HelmChartReference{}, err
	}

	return response.HelmChartReferences, nil
}

func (p HelmChartReferencesService) ListForRelease(productSlug string, releaseID int) ([]HelmChartReference, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/helm_chart_references",
		productSlug,
		releaseID,
	)

	var response HelmChartReferencesResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return []HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []HelmChartReference{}, err
	}

	return response.HelmChartReferences, nil
}

func (p HelmChartReferencesService) Get(productSlug string, helmChartReferenceID int) (HelmChartReference, error) {
	url := fmt.Sprintf(
		"/products/%s/helm_chart_references/%d",
		productSlug,
		helmChartReferenceID,
	)

	var response HelmChartReferenceResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return HelmChartReference{}, err
	}

	return response.HelmChartReference, nil
}

func (p HelmChartReferencesService) Update(productSlug string, helmChartReference HelmChartReference) (HelmChartReference, error) {
	url := fmt.Sprintf("/products/%s/helm_chart_references/%d", productSlug, helmChartReference.ID)

	body := createUpdateHelmChartReferenceBody{
		HelmChartReference: HelmChartReference{
			Description:        helmChartReference.Description,
			DocsURL:            helmChartReference.DocsURL,
			SystemRequirements: helmChartReference.SystemRequirements,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return HelmChartReference{}, err
	}

	var response HelmChartReferenceResponse
	resp, err := p.client.MakeRequest(
		"PATCH",
		url,
		http.StatusOK,
		bytes.NewReader(b),
	)
	if err != nil {
		return HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return HelmChartReference{}, err
	}

	return response.HelmChartReference, nil
}

func (p HelmChartReferencesService) GetForRelease(productSlug string, releaseID int, helmChartReferenceID int) (HelmChartReference, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/helm_chart_references/%d",
		productSlug,
		releaseID,
		helmChartReferenceID,
	)

	var response HelmChartReferenceResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return HelmChartReference{}, err
	}

	return response.HelmChartReference, nil
}

func (p HelmChartReferencesService) Create(config CreateHelmChartReferenceConfig) (HelmChartReference, error) {
	url := fmt.Sprintf("/products/%s/helm_chart_references", config.ProductSlug)

	body := createUpdateHelmChartReferenceBody{
		HelmChartReference: HelmChartReference{
			Description:        config.Description,
			DocsURL:            config.DocsURL,
			Name:               config.Name,
			SystemRequirements: config.SystemRequirements,
			Version:            config.Version,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		// Untested as we cannot force an error because we are marshalling
		// a known-good body
		return HelmChartReference{}, err
	}

	var response HelmChartReferenceResponse
	resp, err := p.client.MakeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
	)
	if err != nil {
		_, ok := err.(ErrTooManyRequests)
		if ok {
			return HelmChartReference{}, fmt.Errorf("You have hit the helm chart reference creation limit. Please wait before creating more helm chart references. Contact pivnet-eng@pivotal.io with additional questions.")
		} else {
			return HelmChartReference{}, err
		}
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return HelmChartReference{}, err
	}

	return response.HelmChartReference, nil
}

func (p HelmChartReferencesService) Delete(productSlug string, id int) (HelmChartReference, error) {
	url := fmt.Sprintf(
		"/products/%s/helm_chart_references/%d",
		productSlug,
		id,
	)

	var response HelmChartReferenceResponse
	resp, err := p.client.MakeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return HelmChartReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return HelmChartReference{}, err
	}

	return response.HelmChartReference, nil
}

func (p HelmChartReferencesService) AddToRelease(
	productSlug string,
	releaseID int,
	helmChartReferenceID int,
) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/add_helm_chart_reference",
		productSlug,
		releaseID,
	)

	body := createUpdateHelmChartReferenceBody{
		HelmChartReference: HelmChartReference{
			ID: helmChartReferenceID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		// Untested as we cannot force an error because we are marshalling
		// a known-good body
		return err
	}

	resp, err := p.client.MakeRequest(
		"PATCH",
		url,
		http.StatusNoContent,
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p HelmChartReferencesService) RemoveFromRelease(
	productSlug string,
	releaseID int,
	helmChartReferenceID int,
) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/remove_helm_chart_reference",
		productSlug,
		releaseID,
	)

	body := createUpdateHelmChartReferenceBody{
		HelmChartReference: HelmChartReference{
			ID: helmChartReferenceID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		// Untested as we cannot force an error because we are marshalling
		// a known-good body
		return err
	}

	resp, err := p.client.MakeRequest(
		"PATCH",
		url,
		http.StatusNoContent,
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
