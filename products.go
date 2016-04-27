package pivnet

import (
	"fmt"
	"net/http"
)

type ProductsService struct {
	client Client
}

type Product struct {
	ID   int    `json:"id,omitempty"`
	Slug string `json:"slug,omitempty"`
	Name string `json:"name,omitempty"`
}

type ProductsResponse struct {
	Products []Product `json:"products,omitempty"`
}

func (p ProductsService) List() ([]Product, error) {
	url := "/products"

	var response ProductsResponse
	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return []Product{}, err
	}

	return response.Products, nil
}

func (p ProductsService) Get(slug string) (Product, error) {
	url := fmt.Sprintf("/products/%s", slug)

	var response Product
	err := p.client.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return Product{}, err
	}

	return response, nil
}
