package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"

	"github.com/hetiansu5/urlquery"
	"github.com/rotisserie/eris"
)

// ResponseType represents HTTP response type.
type ResponseType string

// Supported list of  HTTP response types.
// TODO:dsuhinin - add another type `stream`. For this type return `io.ReadCloser`.
const (
	ResponseTypeJSON   ResponseType = "json"
	ResponseTypeBuffer ResponseType = "buffer"
)

// HttpClient represents HTTP client.
type HttpClient struct {
	client       *http.Client
	baseURL      string
	basePath     string
	namespace    string
	method       string
	params       any
	headers      map[string]string
	request      any
	response     any
	responseType ResponseType
}

// NewClient creates new preconfigured HTTP client.
func NewClient(baseURL, basePath string) *HttpClient {
	return &HttpClient{
		client:   &http.Client{},
		baseURL:  baseURL,
		basePath: basePath,
		method:   http.MethodGet,
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		responseType: ResponseTypeJSON,
	}
}

// NewMlflowApiClient creates new HTTP client for the mlflow api
func NewMlflowApiClient(baseURL string) *HttpClient {
	return NewClient(baseURL, "/api/2.0/mlflow")
}

// NewAimApiClient creates new HTTP client for the aim api
func NewAimApiClient(baseURL string) *HttpClient {
	return NewClient(baseURL, "/aim/api")
}

// WithMethod sets the HTTP method.
func (c *HttpClient) WithMethod(method string) *HttpClient {
	c.method = method
	return c
}

// WithQuery adds query parameters to the HTTP request.
func (c *HttpClient) WithQuery(params any) *HttpClient {
	c.params = params
	return c
}

// WithRequest sets request object.
func (c *HttpClient) WithRequest(request any) *HttpClient {
	c.request = request
	return c
}

// WithNamespace sets the namespace path.
func (c *HttpClient) WithNamespace(namespace string) *HttpClient {
	c.namespace = namespace
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
// noling:gocyclo
func (c *HttpClient) DoRequest(uri string, values ...any) error {
	// 1. check if request object were provided. if provided then marshal it.
	var requestBody io.Reader
	if c.request != nil {
		data, err := json.Marshal(c.request)
		if err != nil {
			return eris.Wrap(err, "error marshaling request object")
		}
		requestBody = bytes.NewBuffer(data)
	}

	// 2. build path with namespace.
	path := c.basePath
	if c.namespace != "" {
		path = fmt.Sprintf("/ns/%s%s", c.namespace, c.basePath)
	}

	// 3. build actual URL.
	u, err := url.Parse(fmt.Sprintf("%s%s%s", c.baseURL, path, fmt.Sprintf(uri, values...)))
	if err != nil {
		return eris.Wrap(err, "error building url")
	}
	// 4. if params were provided then add params to actual url.
	if c.params != nil {
		switch reflect.ValueOf(c.params).Kind() {
		case reflect.Struct:
			query, err := urlquery.Marshal(c.params)
			if err != nil {
				return eris.New("error marshaling params")
			}
			u.RawQuery = string(query)
		case reflect.Map:
			query := u.Query()
			for key, value := range c.params.(map[any]any) {
				query.Set(fmt.Sprintf("%v", key), fmt.Sprintf("%v", value))
			}
			u.RawQuery = query.Encode()
		default:
			return eris.New("unsupported type of params. should be struct or map[any]any")
		}
	}

	// 5. create actual request object.
	// if HttpMethod was not provided, then by default use HttpMethodGet.
	req, err := http.NewRequestWithContext(
		context.Background(), c.method, u.String(), requestBody,
	)
	if err != nil {
		return eris.Wrap(err, "error creating request")
	}

	// 6. if headers were provided, then attach them.
	// by default attach `"Content-Type", "application/json"`
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	// 7. send request data.
	resp, err := c.client.Do(req)
	if err != nil {
		return eris.Wrap(err, "error doing request")
	}

	// 8. read and check response data.
	if c.response != nil {
		switch c.responseType {
		case ResponseTypeJSON:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return eris.Wrap(err, "error reading response data")
			}
			//nolint:errcheck
			defer resp.Body.Close()
			if err := json.Unmarshal(body, c.response); err != nil {
				return eris.Wrap(err, "error unmarshaling response data")
			}
		case ResponseTypeBuffer:
			buffer, ok := c.response.(io.Writer)
			if !ok {
				return eris.New("response object has no implementation of a io.Writer")
			}
			_, err := io.Copy(buffer, resp.Body)
			//nolint:errcheck
			defer resp.Body.Close()
			if err != nil {
				return eris.Wrap(err, "error reading streaming response")
			}
		}
	}

	return nil
}
