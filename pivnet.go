package pivnet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/pivotal-golang/lager"
)

const (
	Endpoint = "https://network.pivotal.io"
	path     = "/api/v2"
)

type Client struct {
	url       string
	token     string
	userAgent string
	logger    lager.Logger
}

type ClientConfig struct {
	Endpoint  string
	Token     string
	UserAgent string
}

func NewClient(config ClientConfig, logger lager.Logger) Client {
	url := fmt.Sprintf("%s%s", config.Endpoint, path)

	return Client{
		url:       url,
		token:     config.Token,
		userAgent: config.UserAgent,
		logger:    logger,
	}
}

func (c Client) makeRequestWithHTTPResponse(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) (*http.Response, error) {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Add("User-Agent", c.userAgent)

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		c.logger.Debug("Error dumping request", lager.Data{"error": err})
		return nil, err
	}

	c.logger.Debug("Making request", lager.Data{"request": string(reqBytes)})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Debug("Error making request", lager.Data{"error": err})
		return nil, err
	}
	defer resp.Body.Close()

	c.logger.Debug("Response status code", lager.Data{"status code": resp.StatusCode})
	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf(
			"Pivnet returned status code: %d for the request - expected %d",
			resp.StatusCode,
			expectedStatusCode,
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(b) > 0 {
		c.logger.Debug("Response body", lager.Data{"response body": string(b)})
		err = json.Unmarshal(b, data)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c Client) makeRequest(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) error {
	_, err := c.makeRequestWithHTTPResponse(
		requestType,
		url,
		expectedStatusCode,
		body,
		data,
	)
	return err
}
