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
	client  *http.Client
	baseURL string
}

// NewHttpClient creates new HTTP client.
func NewHttpClient(baseURL string) *HttpClient {
	return &HttpClient{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

// DoPostRequest do POST request.
func (c HttpClient) DoPostRequest(uri string, request interface{}, response interface{}) error {
	// 1. create and serialize request data.
	data, err := json.Marshal(request)
	if err != nil {
		return eris.Wrap(err, "error marshaling request")
	}

	// 2. create actual request object.
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		StrReplace(
			fmt.Sprintf(
				"%s/api/2.0/mlflow%s",
				os.Getenv("SERVICE_BASE_URL"),
				uri,
			),
			[]string{},
			[]interface{}{},
		),
		bytes.NewBuffer(data),
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
	// nolint
	defer resp.Body.Close()
	if resp != nil {
		if err := json.Unmarshal(body, response); err != nil {
			return eris.Wrap(err, "error unmarshaling response data")
		}
	}

	return nil
}

// DoGetRequest do GET request.
func (c HttpClient) DoGetRequest(uri string, response interface{}) error {
	// 1. create actual request object.
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		StrReplace(
			fmt.Sprintf(
				"%s/api/2.0/mlflow%s",
				os.Getenv("SERVICE_BASE_URL"),
				uri,
			),
			[]string{},
			[]interface{}{},
		),
		nil,
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
	// nolint
	defer resp.Body.Close()
	if resp != nil {
		if err := json.Unmarshal(body, response); err != nil {
			return eris.Wrap(err, "error unmarshaling response data")
		}
	}

	return nil
}
