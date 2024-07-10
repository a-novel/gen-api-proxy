package v1

import (
	"context"
	"errors"
	gatewayutils "github.com/a-novel/gateway-utils"
	"net/http"
	"net/url"
)

// Implements the PingAPI interface.
type pingAPI struct {
	// The root URL for accessing the Gen-API service.
	endpoint string
}

func (api *pingAPI) Call(ctx context.Context) (int, error) {
	path, err := url.JoinPath(api.endpoint, "/ping")
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.Join(gatewayutils.ErrUnavailable, err)
	}

	// If the /ping endpoint returns a non-200 status code, it means the server is running but there is a major
	// issue, preventing it from working normally. This is a case for concern.
	if err := gatewayutils.EnsureStatus(res, http.StatusOK); err != nil {
		return res.StatusCode, err
	}

	return res.StatusCode, nil
}

// NewPingAPI returns a new instance of PingAPI.
//
// The endpoint is the root URL for accessing the Gen-API service.
func NewPingAPI(endpoint string) gatewayutils.PingAPI {
	return &pingAPI{
		endpoint: endpoint,
	}
}
