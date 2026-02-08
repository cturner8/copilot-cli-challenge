package github

import (
	"fmt"
	"net/http"
	"os"
)

// CreateHTTPClient creates an HTTP client with optional GitHub authentication.
// If authenticated is true, it reads the GITHUB_TOKEN environment variable
// and adds the Authorization header to all requests.
func CreateHTTPClient(authenticated bool) (*http.Client, error) {
	client := &http.Client{}

	if !authenticated {
		return client, nil
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Create a custom transport that adds the Authorization header
	transport := &authenticatedTransport{
		token:     token,
		transport: http.DefaultTransport,
	}

	client.Transport = transport
	return client, nil
}

// authenticatedTransport is an http.RoundTripper that adds GitHub authentication
type authenticatedTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.transport.RoundTrip(req)
}
