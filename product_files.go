package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ProductFilesService struct {
	client Client
}

type CreateProductFileConfig struct {
	ProductSlug  string
	FileVersion  string
	AWSObjectKey string
	Name         string
	MD5          string
	Description  string
}

type ProductFilesResponse struct {
	ProductFiles []ProductFile `json:"product_files,omitempty"`
}

type ProductFileResponse struct {
	ProductFile ProductFile `json:"product_file,omitempty"`
}

type ProductFile struct {
	ID           int    `json:"id,omitempty"`
	AWSObjectKey string `json:"aws_object_key,omitempty"`
	Links        *Links `json:"_links,omitempty"`
	FileType     string `json:"file_type,omitempty"`
	FileVersion  string `json:"file_version,omitempty"`
	Name         string `json:"name,omitempty"`
	MD5          string `json:"md5,omitempty"`
	Description  string `json:"description,omitempty"`
}

func (p ProductFilesService) ListForRelease(productSlug string, releaseID int) ([]ProductFile, error) {
	url := fmt.Sprintf("/products/%s/releases/%d/product_files",
		productSlug,
		releaseID,
	)
	response := ProductFilesResponse{}

	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return []ProductFile{}, err
	}

	return response.ProductFiles, nil
}

func (p ProductFilesService) Get(productSlug string, releaseID int, productFileID int) (ProductFile, error) {
	url := fmt.Sprintf("/products/%s/releases/%d/product_files/%d",
		productSlug,
		releaseID,
		productFileID,
	)
	response := ProductFileResponse{}

	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}

func (p ProductFilesService) Create(config CreateProductFileConfig) (ProductFile, error) {
	if config.AWSObjectKey == "" {
		return ProductFile{}, fmt.Errorf("AWS object key must not be empty")
	}

	url := "/products/" + config.ProductSlug + "/product_files"

	body := createProductFileBody{
		ProductFile: ProductFile{
			MD5:          config.MD5,
			FileType:     "Software",
			FileVersion:  config.FileVersion,
			AWSObjectKey: config.AWSObjectKey,
			Name:         config.Name,
			Description:  config.Description,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response ProductFileResponse
	err = p.client.makeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}

type createProductFileBody struct {
	ProductFile ProductFile `json:"product_file"`
}

func (p ProductFilesService) Delete(productSlug string, id int) (ProductFile, error) {
	url := fmt.Sprintf(
		"/products/%s/product_files/%d",
		productSlug,
		id,
	)

	var response ProductFileResponse
	err := p.client.makeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}

func (p ProductFilesService) AddToRelease(
	productSlug string,
	releaseID int,
	productFileID int,
) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/add_product_file",
		productSlug,
		releaseID,
	)

	body := createProductFileBody{
		ProductFile: ProductFile{
			ID: productFileID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	err = p.client.makeRequest(
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
