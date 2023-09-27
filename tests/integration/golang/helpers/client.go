package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rotisserie/eris"
)

// ResponseType represents HTTP response type.
type ResponseType string

// Supported list of  HTTP response types.
const (
	ResponseTypeJSON   ResponseType = "json"
	ResponseTypeStream ResponseType = "stream"
)

// HttpClient represents HTTP client.
type HttpClient struct {
	client       *http.Client
	baseURL      string
	basePath     string
	method       string
	params       map[any]any
	headers      map[string]string
	request      any
	response     any
	responseType ResponseType
}

// NewMlflowApiClient creates new HTTP client for the mlflow api
func NewMlflowApiClient(baseURL string) *HttpClient {
	return &HttpClient{
		client:   &http.Client{},
		baseURL:  baseURL,
		basePath: "/api/2.0/mlflow",
		method:   http.MethodGet,
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		responseType: ResponseTypeJSON,
	}
}

// NewAimApiClient creates new HTTP client for the aim api
func NewAimApiClient(baseURL string) *HttpClient {
	return &HttpClient{
		client:   &http.Client{},
		baseURL:  baseURL,
		basePath: "/aim/api",
		method:   http.MethodGet,
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		responseType: ResponseTypeJSON,
	}
}

// WithMethod sets the HTTP method to use.
func (c *HttpClient) WithMethod(method string) *HttpClient {
	c.method = method
	return c
}

// WithParams adds query parameters to the HTTP request.
func (c *HttpClient) WithParams(params map[any]any) *HttpClient {
	c.params = params
	return c
}

// WithRequest sets request object.
func (c *HttpClient) WithRequest(request any) *HttpClient {
	c.request = request
	return c
}

// WithHeaders adds headers to the HTTP request.
func (c *HttpClient) WithHeaders(headers map[string]string) *HttpClient {
	c.headers = headers
	return c
}

// WithResponse sets the response object where HTTP response will be deserialized.
func (c *HttpClient) WithResponse(response any) *HttpClient {
	c.response = response
	return c
}

// WithResponseType sets the response object type.
func (c *HttpClient) WithResponseType(responseType ResponseType) *HttpClient {
	c.responseType = responseType
	return c
}

// DoRequest do actual HTTP request based on provided parameters.
func (c *HttpClient) DoRequest(uri string) error {
	// 1. check if request object were provided. if provided then marshal it.
	var requestBody io.Reader
	if c.request != nil {
		data, err := json.Marshal(c.request)
		if err != nil {
			return eris.Wrap(err, "error marshaling request object")
		}
		requestBody = bytes.NewBuffer(data)
	}

	// 2. build actual URL.
	u, err := url.Parse(fmt.Sprintf("%s%s%s", c.baseURL, c.basePath, uri))
	if err != nil {
		return eris.Wrap(err, "error building url")
	}
	// 3. if params were provided then add params to actual url.
	if c.params != nil {
		query := u.Query()
		for key, value := range c.params {
			query.Set(fmt.Sprintf("%v", key), fmt.Sprintf("%v", value))
		}
		u.RawQuery = query.Encode()
	}

	// 4. create actual request object.
	// if HttpMethod was not provided, then by default use HttpMethodGet.
	req, err := http.NewRequestWithContext(
		context.Background(), string(c.method), u.String(), requestBody,
	)
	if err != nil {
		return eris.Wrap(err, "error creating request")
	}

	// 5. if headers were provided, then attach them.
	// by default attach `"Content-Type", "application/json"`
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	// 6. send request data.
	resp, err := c.client.Do(req)
	if err != nil {
		return eris.Wrap(err, "error doing request")
	}

	// 7. read and check response data.
	if c.response != nil {
		switch c.responseType {
		case ResponseTypeJSON:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return eris.Wrap(err, "error reading response data")
			}
			defer resp.Body.Close()
			if err := json.Unmarshal(body, c.response); err != nil {
				return eris.Wrap(err, "error unmarshaling response data")
			}
		case ResponseTypeStream:
			body, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return eris.Wrap(err, "error reading streaming response")
			}
			c.response = body
		}
	}

	return nil
}

// DoPostRequest do POST request.
func (c *HttpClient) DoPostRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.doRequest(http.MethodPost, uri, response, bytes.NewBuffer(data))
}

// DoPutRequest do PUT request.
func (c *HttpClient) DoPutRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.doRequest(http.MethodPut, uri, response, bytes.NewBuffer(data))
}

// DoGetRequest do GET request.
func (c *HttpClient) DoGetRequest(uri string, response interface{}) error {
	return c.doRequest(http.MethodGet, uri, response, nil)
}

// DoDeleteRequest do DELETE request.
func (c *HttpClient) DoDeleteRequest(uri string, response interface{}) error {
	return c.doRequest(http.MethodDelete, uri, response, nil)
}

// DoStreamRequest do stream request.
func (c *HttpClient) DoStreamRequest(method, uri string, request interface{}) ([]byte, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, eris.Wrap(err, "error marshaling request")
	}

	// 1. create actual request object.
	req, err := http.NewRequestWithContext(
		context.Background(),
		method,
		StrReplace(
			fmt.Sprintf(
				"%s%s%s",
				c.baseURL,
				c.basePath,
				uri,
			),
			[]string{},
			[]interface{}{},
		),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, eris.Wrap(err, "error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	// 2. send request data.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, eris.Wrap(err, "error doing request")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, eris.Wrap(err, "error reading streaming response")
	}

	return body, nil
}

// doRequest do request of any http method
func (c *HttpClient) doRequest(httpMethod string, uri string, response interface{}, requestBody io.Reader) error {
	// 1. create actual request object.
	req, err := http.NewRequestWithContext(
		context.Background(),
		httpMethod,
		StrReplace(
			fmt.Sprintf(
				"%s%s%s",
				c.baseURL,
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
