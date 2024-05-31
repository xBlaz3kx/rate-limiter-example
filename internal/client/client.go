package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type client interface {
	SendRequest(ctx context.Context) error
}

type httpClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	url        string
}

// newHttpClient creates a client.
func newHttpClient(id, url string) *httpClient {
	return &httpClient{
		logger:     zap.L().Named(fmt.Sprintf("client-%s", id)),
		httpClient: &http.Client{},

		// Append the client id as a query parameter
		url: fmt.Sprintf("%s?clientId=%s", url, id),
	}
}

// SendRequest sends a request to the client's URL. The client ID is appended as a query parameter.
func (c *httpClient) SendRequest(ctx context.Context) error {
	c.logger.Info("Sending request", zap.String("url", c.url))

	// Create and send the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusTooManyRequests:
		return errors.Errorf("request limit exceeded")
	default:
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
