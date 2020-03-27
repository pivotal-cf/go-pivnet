package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ImageReferencesService struct {
	client Client
}

type CreateImageReferenceConfig struct {
	ProductSlug        string
	Description        string
	DocsURL            string
	Digest             string
	Name               string
	ImagePath          string
	SystemRequirements []string
}

type ImageReferencesResponse struct {
	ImageReferences []ImageReference `json:"image_references,omitempty"`
}

type ImageReferenceResponse struct {
	ImageReference ImageReference `json:"image_reference,omitempty"`
}

type ReplicationStatus string

const (
	InProgress        ReplicationStatus = "in_progress"
	Complete          ReplicationStatus = "complete"
	FailedToReplicate ReplicationStatus = "failed_to_replicate"
)

type ImageReference struct {
	ID                 int               `json:"id,omitempty" yaml:"id,omitempty"`
	ImagePath          string            `json:"image_path,omitempty" yaml:"image_path,omitempty"`
	Description        string            `json:"description,omitempty" yaml:"description,omitempty"`
	Digest             string            `json:"digest,omitempty" yaml:"digest,omitempty"`
	DocsURL            string            `json:"docs_url,omitempty" yaml:"docs_url,omitempty"`
	Name               string            `json:"name,omitempty" yaml:"name,omitempty"`
	SystemRequirements []string          `json:"system_requirements,omitempty" yaml:"system_requirements,omitempty"`
	ReleaseVersions    []string          `json:"release_versions,omitempty" yaml:"release_versions,omitempty"`
	ReplicationStatus  ReplicationStatus `json:"replication_status,omitempty" yaml:"replication_status,omitempty"`
}

type createUpdateImageReferenceBody struct {
	ImageReference ImageReference `json:"image_reference"`
}

func (p ImageReferencesService) List(productSlug string) ([]ImageReference, error) {
	url := fmt.Sprintf("/products/%s/image_references", productSlug)

	var response ImageReferencesResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return []ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []ImageReference{}, err
	}

	return response.ImageReferences, nil
}

func (p ImageReferencesService) ListForDigest(productSlug string, digest string) ([]ImageReference, error) {
	url := fmt.Sprintf("/products/%s/image_references", productSlug)
	params := []QueryParameter{
		{"digest", digest},
	}

	var response ImageReferencesResponse
	resp, err := p.client.MakeRequestWithParams(
		"GET",
		url,
		http.StatusOK,
		params,
		nil,
	)
	if err != nil {
		return []ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []ImageReference{}, err
	}

	return response.ImageReferences, nil
}

func (p ImageReferencesService) ListForRelease(productSlug string, releaseID int) ([]ImageReference, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/image_references",
		productSlug,
		releaseID,
	)

	var response ImageReferencesResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return []ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []ImageReference{}, err
	}

	return response.ImageReferences, nil
}

func (p ImageReferencesService) Get(productSlug string, imageReferenceID int) (ImageReference, error) {
	url := fmt.Sprintf(
		"/products/%s/image_references/%d",
		productSlug,
		imageReferenceID,
	)

	var response ImageReferenceResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ImageReference{}, err
	}

	return response.ImageReference, nil
}

func (p ImageReferencesService) Update(productSlug string, imageReference ImageReference) (ImageReference, error) {
	url := fmt.Sprintf("/products/%s/image_references/%d", productSlug, imageReference.ID)

	body := createUpdateImageReferenceBody{
		ImageReference: ImageReference{
			Description:        imageReference.Description,
			Name:               imageReference.Name,
			DocsURL:            imageReference.DocsURL,
			SystemRequirements: imageReference.SystemRequirements,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return ImageReference{}, err
	}

	var response ImageReferenceResponse
	resp, err := p.client.MakeRequest(
		"PATCH",
		url,
		http.StatusOK,
		bytes.NewReader(b),
	)
	if err != nil {
		return ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ImageReference{}, err
	}

	return response.ImageReference, nil
}

func (p ImageReferencesService) GetForRelease(productSlug string, releaseID int, imageReferenceID int) (ImageReference, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/image_references/%d",
		productSlug,
		releaseID,
		imageReferenceID,
	)

	var response ImageReferenceResponse
	resp, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ImageReference{}, err
	}

	return response.ImageReference, nil
}

func (p ImageReferencesService) Create(config CreateImageReferenceConfig) (ImageReference, error) {
	url := fmt.Sprintf("/products/%s/image_references", config.ProductSlug)

	body := createUpdateImageReferenceBody{
		ImageReference: ImageReference{
			ImagePath:          config.ImagePath,
			Description:        config.Description,
			Digest:             config.Digest,
			DocsURL:            config.DocsURL,
			Name:               config.Name,
			SystemRequirements: config.SystemRequirements,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		// Untested as we cannot force an error because we are marshalling
		// a known-good body
		return ImageReference{}, err
	}

	var response ImageReferenceResponse
	resp, err := p.client.MakeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
	)
	if err != nil {
		_, ok := err.(ErrTooManyRequests)
		if ok {
			return ImageReference{}, fmt.Errorf("You have hit the image reference creation limit. Please wait before creating more image references. Contact pivnet-eng@pivotal.io with additional questions.")
		} else {
			return ImageReference{}, err
		}
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ImageReference{}, err
	}

	return response.ImageReference, nil
}

func (p ImageReferencesService) Delete(productSlug string, id int) (ImageReference, error) {
	url := fmt.Sprintf(
		"/products/%s/image_references/%d",
		productSlug,
		id,
	)

	var response ImageReferenceResponse
	resp, err := p.client.MakeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
	)
	if err != nil {
		return ImageReference{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ImageReference{}, err
	}

	return response.ImageReference, nil
}

func (p ImageReferencesService) AddToRelease(
	productSlug string,
	releaseID int,
	imageReferenceID int,
) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/add_image_reference",
		productSlug,
		releaseID,
	)

	body := createUpdateImageReferenceBody{
		ImageReference: ImageReference{
			ID: imageReferenceID,
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

func (p ImageReferencesService) RemoveFromRelease(
	productSlug string,
	releaseID int,
	imageReferenceID int,
) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/remove_image_reference",
		productSlug,
		releaseID,
	)

	body := createUpdateImageReferenceBody{
		ImageReference: ImageReference{
			ID: imageReferenceID,
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
