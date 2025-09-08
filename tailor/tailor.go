package tailor

import (
	"errors"
	"net/http"
	"os"

	"buf.build/gen/go/tailor-inc/tailor/connectrpc/go/tailor/v1/tailorv1connect"
	"github.com/tailor-platform/patterner/config"
)

type Client struct {
	client tailorv1connect.OperatorServiceClient
	cfg    *config.Config
}

func New(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if cfg.WorkspaceID == "" {
		return nil, errors.New("workspace ID is required")
	}

	baseURL := "https://api.tailor.tech"
	if platformURL := os.Getenv("PLATFORM_URL"); platformURL != "" {
		baseURL = platformURL
	}

	// Create HTTP client with Bearer token authorization
	httpClient := &http.Client{}
	if token := os.Getenv("TAILOR_TOKEN"); token != "" {
		httpClient = &http.Client{
			Transport: &bearerTokenTransport{
				token: token,
				base:  http.DefaultTransport,
			},
		}
	}

	return &Client{
		client: tailorv1connect.NewOperatorServiceClient(httpClient, baseURL),
		cfg:    cfg,
	}, nil
}

// bearerTokenTransport implements http.RoundTripper to add Bearer token to requests.
type bearerTokenTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}
