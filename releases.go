package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pivotal-golang/lager"
)

type ReleasesService struct {
	client Client
}

type createReleaseBody struct {
	Release Release `json:"release"`
}

type ReleasesResponse struct {
	Releases []Release `json:"releases,omitempty"`
}

type CreateReleaseResponse struct {
	Release Release `json:"release,omitempty"`
}

type Release struct {
	ID                    int    `json:"id,omitempty"`
	Availability          string `json:"availability,omitempty"`
	EULA                  *EULA  `json:"eula,omitempty"`
	OSSCompliant          string `json:"oss_compliant,omitempty"`
	ReleaseDate           string `json:"release_date,omitempty"`
	ReleaseType           string `json:"release_type,omitempty"`
	Version               string `json:"version,omitempty"`
	Links                 *Links `json:"_links,omitempty"`
	Description           string `json:"description,omitempty"`
	ReleaseNotesURL       string `json:"release_notes_url,omitempty"`
	Controlled            bool   `json:"controlled,omitempty"`
	ECCN                  string `json:"eccn,omitempty"`
	LicenseException      string `json:"license_exception,omitempty"`
	EndOfSupportDate      string `json:"end_of_support_date,omitempty"`
	EndOfGuidanceDate     string `json:"end_of_guidance_date,omitempty"`
	EndOfAvailabilityDate string `json:"end_of_availability_date,omitempty"`
}

type CreateReleaseConfig struct {
	ProductSlug           string
	ProductVersion        string
	ReleaseType           string
	ReleaseDate           string
	EULASlug              string
	Description           string
	ReleaseNotesURL       string
	Controlled            bool
	ECCN                  string
	LicenseException      string
	EndOfSupportDate      string
	EndOfGuidanceDate     string
	EndOfAvailabilityDate string
}

func (r ReleasesService) List(productSlug string) ([]Release, error) {
	url := fmt.Sprintf("/products/%s/releases", productSlug)

	var response ReleasesResponse
	err := r.client.makeRequest("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Releases, nil
}

func (r ReleasesService) Create(config CreateReleaseConfig) (Release, error) {
	url := fmt.Sprintf("/products/%s/releases", config.ProductSlug)

	body := createReleaseBody{
		Release: Release{
			Availability: "Admins Only",
			EULA: &EULA{
				Slug: config.EULASlug,
			},
			OSSCompliant:          "confirm",
			ReleaseDate:           config.ReleaseDate,
			ReleaseType:           config.ReleaseType,
			Version:               config.ProductVersion,
			Description:           config.Description,
			ReleaseNotesURL:       config.ReleaseNotesURL,
			Controlled:            config.Controlled,
			ECCN:                  config.ECCN,
			LicenseException:      config.LicenseException,
			EndOfSupportDate:      config.EndOfSupportDate,
			EndOfGuidanceDate:     config.EndOfGuidanceDate,
			EndOfAvailabilityDate: config.EndOfAvailabilityDate,
		},
	}

	if config.ReleaseDate == "" {
		body.Release.ReleaseDate = time.Now().Format("2006-01-02")
		r.client.logger.Debug("No release date found - defaulting to", lager.Data{"release date": body.Release.ReleaseDate})
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response CreateReleaseResponse
	err = r.client.makeRequest("POST", url, http.StatusCreated, bytes.NewReader(b), &response)
	if err != nil {
		return Release{}, err
	}

	return response.Release, nil
}

func (r ReleasesService) Update(productSlug string, release Release) (Release, error) {
	url := fmt.Sprintf("/products/%s/releases/%d", productSlug, release.ID)

	release.OSSCompliant = "confirm"

	var updatedRelease = createReleaseBody{
		Release: release,
	}

	body, err := json.Marshal(updatedRelease)
	if err != nil {
		panic(err)
	}

	var response CreateReleaseResponse
	err = r.client.makeRequest("PATCH", url, http.StatusOK, bytes.NewReader(body), &response)
	if err != nil {
		return Release{}, err
	}

	return response.Release, nil
}

func (r ReleasesService) Delete(release Release, productSlug string) error {
	url := fmt.Sprintf("/products/%s/releases/%d", productSlug, release.ID)

	err := r.client.makeRequest(
		"DELETE",
		url,
		http.StatusNoContent,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
