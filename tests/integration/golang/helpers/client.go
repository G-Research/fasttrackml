package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rotisserie/eris"
)

// HttpClient represents HTTP client.
type HttpClient struct {
	client   *http.Client
	baseURL  string
	basePath string
}

// NewMlflowApiClient creates new HTTP client for the mlflow api
func NewMlflowApiClient(baseURL string) *HttpClient {
	return &HttpClient{
		client:   &http.Client{},
		baseURL:  baseURL,
		basePath: "/api/2.0/mlflow",
	}
}

// NewAimApiClient creates new HTTP client for the aim api
func NewAimApiClient(baseURL string) *HttpClient {
	return &HttpClient{
		client:   &http.Client{},
		baseURL:  baseURL,
		basePath: "/aim/api",
	}
}

// DoPostRequest do POST request.
func (c HttpClient) DoPostRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.DoRequest(http.MethodPost, uri, response, bytes.NewBuffer(data))}

// DoPostRequest do PUT request.
func (c HttpClient) DoPutRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.DoRequest(http.MethodPut, uri, response, bytes.NewBuffer(data))
}

// DoGetRequest do GET request.
func (c HttpClient) DoGetRequest(uri string, response interface{}) error {
	return c.DoRequest(http.MethodGet, uri, response, nil)
}

// DoDeleteRequest do DELETE request.
func (c HttpClient) DoDeleteRequest(uri string, response interface{}) error {
	return c.DoRequest(http.MethodDelete, uri, response, nil)
}

// DoRequest do request of any http method
func (c HttpClient) DoRequest(httpMethod string, uri string, response interface{}, requestBody io.Reader) error {
	// 1. create actual request object.
	req, err := http.NewRequestWithContext(
		context.Background(),
		httpMethod,
		StrReplace(
			fmt.Sprintf(
				"%s%s%s",
				os.Getenv("SERVICE_BASE_URL"),
				c.basePath,
				uri,
			),
			[]string{},
			[]interface{}{},
		),
		requestBody,
	)
	if err != nil {
		return eris.Wrap(err, "error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	// 3. send request data.
	resp, err := c.client.Do(req)
	if err != nil {
		return eris.Wrap(err, "error doing request")
	}

	// 4. read and check response data.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return eris.Wrap(err, "error reading response data")
	}
	defer resp.Body.Close()
	if err := json.Unmarshal(body, response); err != nil {
		return eris.Wrap(err, "error unmarshaling response data")
	}

	return nil
}
