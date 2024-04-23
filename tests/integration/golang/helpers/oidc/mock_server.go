package oidc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

// MockServer represents OIDC mock server.
type MockServer struct {
	oidcMockServer *mockoidc.MockOIDC
}

// NewMockServer creates new OIDC mock server.
func NewMockServer() (*MockServer, error) {
	oidcMockServer, err := mockoidc.Run()
	if err != nil {
		return nil, eris.Wrap(err, "error running oidc mock server")
	}
	return &MockServer{
		oidcMockServer: oidcMockServer,
	}, nil
}

// Login mimics User login action.
func (m MockServer) Login(user *mockoidc.MockUser, scopes []string) (string, error) {
	// Emulate client to IDP request.
	authorizeQuery := url.Values{}
	authorizeQuery.Set("client_id", m.oidcMockServer.ClientID)
	authorizeQuery.Set("state", helpers.GenerateRandomString(10))
	authorizeQuery.Set("nonce", helpers.GenerateRandomString(10))
	authorizeQuery.Set("scope", strings.Join(scopes, " "))
	authorizeQuery.Set("response_type", "code")
	authorizeQuery.Set("redirect_uri", "http://127.0.0.1/oauth2/callback")

	codeVerifier := "sum"
	challenge, err := mockoidc.GenerateCodeChallenge(mockoidc.CodeChallengeMethodS256, codeVerifier)
	if err != nil {
		return "", eris.Wrapf(err, "error generating code challenge")
	}
	authorizeQuery.Set("code_challenge", challenge)
	authorizeQuery.Set("code_challenge_method", mockoidc.CodeChallengeMethodS256)

	authorizeURL, err := url.Parse(m.oidcMockServer.AuthorizationEndpoint())
	if err != nil {
		return "", eris.Wrapf(err, "error parsing authorization endpoint")
	}
	authorizeURL.RawQuery = authorizeQuery.Encode()

	authorizeRequest, err := http.NewRequest(http.MethodGet, authorizeURL.String(), nil)
	if err != nil {
		return "", eris.Wrap(err, "error creating authorize request")
	}

	m.oidcMockServer.QueueUser(user)
	m.oidcMockServer.QueueCode(helpers.GenerateRandomString(10))

	// A custom client that doesn't automatically follow redirects
	var httpClient = &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	authorizeResponse, err := httpClient.Do(authorizeRequest)
	if err != nil {
		return "", eris.Wrap(err, "error making authorization request")
	}
	if authorizeResponse.StatusCode != http.StatusFound {
		body, err := io.ReadAll(authorizeResponse.Body)
		if err != nil {
			return "", eris.Wrap(err, "error reading authorization response body")
		}
		return "", eris.Errorf(
			"oidc server returns non 302 http code during authorization request, body: %s", string(body),
		)
	}

	redirectURL, err := url.Parse(authorizeResponse.Header.Get("Location"))
	if err != nil {
		return "", eris.Wrapf(err, "error getting location header from authorization response")
	}

	// emulate appRedirect handling token endpoint call.
	tokenForm := url.Values{}
	tokenForm.Set("client_id", m.oidcMockServer.ClientID)
	tokenForm.Set("client_secret", m.oidcMockServer.ClientSecret)
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", redirectURL.Query().Get("code"))
	tokenForm.Set("code_verifier", codeVerifier)

	tokenRequest, err := http.NewRequest(
		http.MethodPost, m.oidcMockServer.TokenEndpoint(), bytes.NewBufferString(tokenForm.Encode()),
	)
	if err != nil {
		return "", eris.Wrap(err, "error making token request")
	}
	tokenRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokenResponse, err := httpClient.Do(tokenRequest)
	if err != nil {
		return "", eris.Wrap(err, "error making token request")
	}
	if tokenResponse.StatusCode != http.StatusOK {
		body, err := io.ReadAll(authorizeResponse.Body)
		if err != nil {
			return "", eris.Wrap(err, "error reading token response body")
		}
		return "", eris.Errorf(
			"oidc server returns non 200 http code during token request, body: %s", string(body),
		)
	}

	defer tokenResponse.Body.Close()
	body, err := io.ReadAll(tokenResponse.Body)
	if err != nil {
		return "", eris.Wrap(err, "error reading token response body")
	}

	var token struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", eris.Wrapf(err, "error unmarshaling token information")
	}

	return token.IDToken, nil
}

// Address returns OIDC mock server address.
func (m MockServer) Address() string {
	return m.oidcMockServer.Addr() + mockoidc.IssuerBase
}

// ClientID returns OIDC mock server ClientID.
func (m MockServer) ClientID() string {
	return m.oidcMockServer.ClientID
}

// ClientSecret returns OIDC mock server ClientSecret.
func (m MockServer) ClientSecret() string {
	return m.oidcMockServer.ClientSecret
}
