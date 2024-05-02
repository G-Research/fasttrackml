package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"

	"github.com/PuerkitoBio/goquery"
	"github.com/hetiansu5/urlquery"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/server"
)

// ResponseType represents HTTP response type.
type ResponseType string

// Supported list of  HTTP response types.
// TODO:dsuhinin - add another type `stream`. For this type return `io.ReadCloser`.
const (
	ResponseTypeJSON   ResponseType = "json"
	ResponseTypeBuffer ResponseType = "buffer"
	ResponseTypeHTML   ResponseType = "html"
)

// HttpClient represents HTTP client.
type HttpClient struct {
	server       server.Server
	basePath     string
	namespace    string
	method       string
	params       any
	headers      map[string]string
	request      any
	response     any
	responseType ResponseType
	statusCode   int
}

// NewClient creates new preconfigured HTTP client.
func NewClient(server server.Server, basePath string) *HttpClient {
	return &HttpClient{
		server:   server,
		basePath: basePath,
		method:   http.MethodGet,
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		responseType: ResponseTypeJSON,
	}
}

// NewMlflowApiClient creates new HTTP client for the mlflow api
func NewMlflowApiClient(server server.Server) *HttpClient {
	return NewClient(server, "/api/2.0/mlflow")
}

// NewAimApiClient creates new HTTP client for the aim api
func NewAimApiClient(server server.Server) *HttpClient {
	return NewClient(server, "/aim/api")
}

// NewAdminApiClient creates new HTTP client for the admin api
func NewAdminApiClient(server server.Server) *HttpClient {
	return NewClient(server, "/admin")
}

// NewChooserApiClient creates new HTTP client for the chooser api
func NewChooserApiClient(server server.Server) *HttpClient {
	return NewClient(server, "/chooser")
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

// GetStatusCode returns HTTP status code of the last response, if available.
func (c *HttpClient) GetStatusCode() int {
	return c.statusCode
}

// DoRequest do actual HTTP request based on provided parameters.
// nolint:gocyclo
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
	u, err := url.Parse(fmt.Sprintf("%s%s", path, fmt.Sprintf(uri, values...)))
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
	req := httptest.NewRequest(
		c.method, u.String(), requestBody,
	)

	// 6. if headers were provided, then attach them.
	// by default attach `"Content-Type", "application/json"`
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	// 7. send request data and handle possible redirects.
	//nolint:bodyclose
	resp, err := c.server.Test(req, 60000)
	if err != nil {
		return eris.Wrap(err, "error doing request")
	}
	if resp.StatusCode == http.StatusMovedPermanently {
		req.RequestURI = resp.Header.Get("location")
		//nolint:bodyclose
		resp, err = c.server.Test(req, 60000)
		if err != nil {
			return eris.Wrap(err, "error doing request")
		}
	}
	//nolint:errcheck
	defer resp.Body.Close()

	c.statusCode = resp.StatusCode

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
			// if ErrorResponse object provided, then populate StatusCode also.
			if _, ok := c.response.(*api.ErrorResponse); ok {
				c.response.(*api.ErrorResponse).StatusCode = resp.StatusCode
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
		case ResponseTypeHTML:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return eris.Wrap(err, "error reading response data")
			}
			//nolint:errcheck
			defer resp.Body.Close()
			response, ok := c.response.(*goquery.Document)
			if !ok {
				return eris.New("response object is not a *goquery.Document")
			}
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				return eris.Wrap(err, "error creating goquery document")
			}

			*response = *doc
		}
	}

	return nil
}
