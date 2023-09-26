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

// HttpMethod http method type.
type HttpMethod string

// Support list of http methods + one custom - "STREAM
const (
	HttpMethodPut    HttpMethod = "PUT"
	HttpMethodGet    HttpMethod = "GET"
	HttpMethodPost   HttpMethod = "POST"
	HttpMethodDelete HttpMethod = "DELETE"
	HttpMethodStream HttpMethod = "STREAM"
)

// HttpRequest represents object to wrap all the Http parameters.
type HttpRequest struct {
	URI      string
	Params   map[any]any
	Method   HttpMethod
	Request  interface{}
	Headers  map[string]string
	Response interface{}
}

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

// DoRequest do actual HTTP request based on provided parameters.
func (c HttpClient) DoRequest(request *HttpRequest) error {
	// 1. check if request object were provided. if provided then marshal it.
	var requestBody io.Reader
	if request.Request != nil {
		data, err := json.Marshal(request)
		if err != nil {
			return eris.Wrap(err, "error marshaling request object")
		}
		requestBody = bytes.NewBuffer(data)
	}

	// 2. build actual URL.
	u, err := url.Parse(fmt.Sprintf("%s%s%s", c.baseURL, c.basePath, request.URI))
	if err != nil {
		return eris.Wrap(err, "error building url")
	}
	// 3. if params were provided then add params to actual url.
	if request.Params != nil {
		query := u.Query()
		for key, value := range request.Params {
			query.Set(fmt.Sprintf("%s", key), fmt.Sprintf("%s", value))
		}
		u.RawQuery = query.Encode()
	}

	// 4. create actual request object.
	// if HttpMethod was not provided, then by default use HttpMethodGet.
	if request.Method == "" {
		request.Method = HttpMethodGet
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		string(request.Method),
		StrReplace(
			fmt.Sprintf(
				"%s%s%s",
				c.baseURL,
				c.basePath,
				request.URI,
			),
			[]string{},
			[]interface{}{},
		),
		requestBody,
	)
	if err != nil {
		return eris.Wrap(err, "error creating request")
	}

	// 5. if headers were provided, then attach them.
	// by default attach `"Content-Type", "application/json"`
	if request.Headers != nil {
		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	// 6. send request data.
	resp, err := c.client.Do(req)
	if err != nil {
		return eris.Wrap(err, "error doing request")
	}

	// 7. read and check response data. if requested method is HttpMethodStream then just read
	// data as is and return it back, otherwise Unmarshal it to provided request.Response object.
	if request.Method == HttpMethodStream {
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return eris.Wrap(err, "error reading streaming response")
		}
		request.Response = body
	} else if request.Response != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return eris.Wrap(err, "error reading response data")
		}
		defer resp.Body.Close()
		if err := json.Unmarshal(body, request.Response); err != nil {
			return eris.Wrap(err, "error unmarshaling response data")
		}
	}

	return nil
}

// DoPostRequest do POST request.
func (c HttpClient) DoPostRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.doRequest(http.MethodPost, uri, response, bytes.NewBuffer(data))
}

// DoPutRequest do PUT request.
func (c HttpClient) DoPutRequest(uri string, request interface{}, response interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}
	return c.doRequest(http.MethodPut, uri, response, bytes.NewBuffer(data))
}

// DoGetRequest do GET request.
func (c HttpClient) DoGetRequest(uri string, response interface{}) error {
	return c.doRequest(http.MethodGet, uri, response, nil)
}

// DoDeleteRequest do DELETE request.
func (c HttpClient) DoDeleteRequest(uri string, response interface{}) error {
	return c.doRequest(http.MethodDelete, uri, response, nil)
}

// DoStreamRequest do stream request.
func (c HttpClient) DoStreamRequest(method, uri string, request interface{}) ([]byte, error) {
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
func (c HttpClient) doRequest(httpMethod string, uri string, response interface{}, requestBody io.Reader) error {
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
