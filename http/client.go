package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

type DownloadRequest struct {
	Path string
}

type DownloadResponse struct {
	Body io.Reader
}

func (c *FileServerClient) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, req.Path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http: unexpected http status code received - %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return &DownloadResponse{Body: &buf}, nil
}

type UploadRequest struct {
	Path    string
	Content io.Reader
}

func (c *FileServerClient) Upload(ctx context.Context, req *UploadRequest) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, req.Path, req.Content)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("http: unexpected http status code received - %d", resp.StatusCode)
	}
	return nil
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
