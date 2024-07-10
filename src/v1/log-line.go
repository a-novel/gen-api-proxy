package v1

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	gatewayutils "github.com/a-novel/gateway-utils"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
)

var (
	ErrInvalidLogLine = errors.New("invalid log line")
)

var mocks struct {
	Create map[string]struct {
		Result string `yaml:"result,omitempty"`
		Status int    `yaml:"status,omitempty"`
		Err    error  `yaml:"err,omitempty"`
	} `yaml:"create,omitempty"`
	Validate map[string]struct {
		Status int   `yaml:"status,omitempty"`
		Err    error `yaml:"err,omitempty"`
	} `yaml:"validate,omitempty"`
}

//go:embed log-line-mocks.yaml
var mocksFile []byte

// Load mocked data.
func init() {
	if err := yaml.Unmarshal(mocksFile, &mocks); err != nil {
		panic(err)
	}
}

// CreateLogLineAPI sends a request to create a new log line from instructions.
type CreateLogLineAPI interface {
	// Call executes the request. It returns the generated log line, along with the status of the response and error,
	// if any.
	//
	// In case the API returns a non-200 status, a utils.StatusError will be thrown.
	Call(ctx context.Context, instruction string, remix []string) (string, int, error)
	// Mock returns a mocked response, based on the chosen scenario.
	Mock(ctx context.Context, useCase string) (string, int, error)
}

// Implements the CreateLogLineAPI interface.
type createLogLineAPI struct {
	// The root URL for accessing the Gen-API service.
	endpoint string
}

func (api *createLogLineAPI) Call(ctx context.Context, instruction string, remix []string) (string, int, error) {
	path, err := url.JoinPath(api.endpoint, "/api/v1/log-lines")
	if err != nil {
		return "", 0, err
	}

	jsonBody, err := json.Marshal(map[string]interface{}{
		"instruction": instruction,
		"remix":       remix,
	})
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, path, bytes.NewReader(jsonBody))
	if err != nil {
		return "", 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}

	if err := gatewayutils.EnsureStatus(res, http.StatusOK); err != nil {
		return "", res.StatusCode, errors.Join(err, gatewayutils.GetResponseError(res))
	}

	responseBody := new(struct{ logLine string })
	if err := gatewayutils.ExtractJSONResponse(res, responseBody); err != nil {
		return "", res.StatusCode, err
	}

	return responseBody.logLine, res.StatusCode, nil
}

func (api *createLogLineAPI) Mock(_ context.Context, useCase string) (string, int, error) {
	if useCase == "" {
		useCase = "success"
	}

	mocked, ok := mocks.Create[useCase]

	if !ok {
		return "", 0, fmt.Errorf("unknown use case: %s", useCase)
	}

	return mocked.Result, mocked.Status, mocked.Err
}

// NewCreateLogLineAPI returns a new instance of CreateLogLineAPI.
//
// The endpoint is the root URL for accessing the Gen-API service.
func NewCreateLogLineAPI(endpoint string) CreateLogLineAPI {
	return &createLogLineAPI{endpoint: endpoint}
}

// ValidateLogLineAPI sends a request to check if a given input is a valid log line.
type ValidateLogLineAPI interface {
	// Call executes the request.
	//
	// If the log line is valid, a 204 status is returned.
	//
	// Otherwise, a 422 status will be returned to indicate the
	// input does not match the requirements for a valid log line. This will result in the ErrInvalidLogLine error
	// being thrown along.
	//
	// Any other status should be interpreted as an unexpected error.
	Call(ctx context.Context, logLine string) (int, error)
	// Mock returns a mocked response, based on the chosen scenario.
	Mock(ctx context.Context, useCase string) (int, error)
}

// Implements the ValidateLogLineAPI interface.
type validateLogLineAPI struct {
	// The root URL for accessing the Gen-API service.
	endpoint string
}

func (api *validateLogLineAPI) Call(ctx context.Context, logLine string) (int, error) {
	path, err := url.JoinPath(api.endpoint, "/api/v1/log-lines")
	if err != nil {
		return 0, err
	}

	jsonBody, err := json.Marshal(map[string]interface{}{
		"logLine": logLine,
	})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	// Special error for an expected use case.
	if res.StatusCode == http.StatusUnprocessableEntity {
		return res.StatusCode, ErrInvalidLogLine
	}

	if err := gatewayutils.EnsureStatus(res, http.StatusNoContent); err != nil {
		return res.StatusCode, errors.Join(err, gatewayutils.GetResponseError(res))
	}

	return res.StatusCode, nil
}

func (api *validateLogLineAPI) Mock(_ context.Context, useCase string) (int, error) {
	if useCase == "" {
		useCase = "success"
	}

	mocked, ok := mocks.Validate[useCase]

	if !ok {
		return 0, fmt.Errorf("unknown use case: %s", useCase)
	}

	return mocked.Status, mocked.Err
}

// NewValidateLogLineAPI returns a new instance of ValidateLogLineAPI.
//
// The endpoint is the root URL for accessing the Gen-API service.
func NewValidateLogLineAPI(endpoint string) ValidateLogLineAPI {
	return &validateLogLineAPI{endpoint: endpoint}
}
