package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

var ErrFileNotFound = errors.New("http: file not found")

// FileServerClient
type FileServerClient struct {
	httpClient *http.Client
}

// NewClient
func NewClient() *FileServerClient {
	return &FileServerClient{
		httpClient: http.DefaultClient,
	}
}

// RemoveRequest
type RemoveRequest struct {
	// Path is the full path to the file you wish to remove.
	Path string
}

// Remove
func (c *FileServerClient) Remove(ctx context.Context, req *RemoveRequest) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, req.Path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("http: unexpected http status code received - %d", resp.StatusCode)
	}
	return nil
}
