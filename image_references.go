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

type ImageReferenceResponse struct {
	ImageReference ImageReference `json:"image_reference,omitempty"`
}

type ImageReference struct {
	ID                 int      `json:"id,omitempty" yaml:"id,omitempty"`
	ImagePath          string   `json:"image_path,omitempty" yaml:"image_path,omitempty"`
	Description        string   `json:"description,omitempty" yaml:"description,omitempty"`
	Digest             string   `json:"digest,omitempty" yaml:"digest,omitempty"`
	DocsURL            string   `json:"docs_url,omitempty" yaml:"docs_url,omitempty"`
	Name               string   `json:"name,omitempty" yaml:"name,omitempty"`
	SystemRequirements []string `json:"system_requirements,omitempty" yaml:"system_requirements,omitempty"`
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

type createUpdateImageReferenceBody struct {
	ImageReference ImageReference `json:"image_reference"`
}
