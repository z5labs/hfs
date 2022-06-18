package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFileRetrieval(t *testing.T) {
	t.Run("FileIsPresent", func(subT *testing.T) {
		testFileContent := "Hello, world!"
		fs := afero.NewMemMapFs()
		afero.WriteFile(fs, "/example.txt", []byte(testFileContent), os.ModePerm)

		h := FileServer(fs)

		req := httptest.NewRequest(http.MethodGet, "http://example.com/example.txt", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		if !assert.Nil(subT, err) {
			return
		}

		if !assert.Equal(subT, http.StatusOK, resp.StatusCode) {
			return
		}
		if !assert.Equal(subT, testFileContent, string(body)) {
			return
		}
	})

	t.Run("FileIsNotPresent", func(subT *testing.T) {
		fs := afero.NewMemMapFs()
		h := FileServer(fs)

		req := httptest.NewRequest(http.MethodGet, "http://example.com/example.txt", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		if !assert.Equal(subT, http.StatusNotFound, resp.StatusCode) {
			return
		}
	})
}

func TestFileUpload(t *testing.T) {
	t.Run("FileIsPresent", func(subT *testing.T) {
		ogFileContent := "Hello, world!"
		fs := afero.NewMemMapFs()
		filename := "/example.txt"
		afero.WriteFile(fs, filename, []byte(ogFileContent), os.ModePerm)

		h := FileServer(fs)

		newFileContent := "Good bye, world!"
		req := httptest.NewRequest(http.MethodPost, "http://example.com/example.txt", strings.NewReader(newFileContent))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		if !assert.Nil(subT, err) {
			return
		}

		if !assert.Equal(subT, http.StatusOK, resp.StatusCode) {
			return
		}
		if !assert.Equal(subT, 0, len(body)) {
			return
		}

		updatedBody, err := afero.ReadFile(fs, filename)
		if !assert.Nil(subT, err) {
			return
		}
		if !assert.Equal(subT, newFileContent, string(updatedBody)) {
			return
		}
	})

	t.Run("FileIsNotPresent", func(subT *testing.T) {
		fs := afero.NewMemMapFs()
		h := FileServer(fs)

		req := httptest.NewRequest(http.MethodPost, "http://example.com/example.txt", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		if !assert.Equal(subT, http.StatusCreated, resp.StatusCode) {
			return
		}
	})
}

func TestFileDeletion(t *testing.T) {
	t.Run("FileIsPresent", func(subT *testing.T) {
		testFileContent := "Hello, world!"
		fs := afero.NewMemMapFs()
		afero.WriteFile(fs, "/example.txt", []byte(testFileContent), os.ModePerm)

		h := FileServer(fs)

		req := httptest.NewRequest(http.MethodDelete, "http://example.com/example.txt", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		if !assert.Equal(subT, http.StatusNoContent, resp.StatusCode) {
			return
		}
	})

	t.Run("FileIsNotPresent", func(subT *testing.T) {
		fs := afero.NewMemMapFs()
		h := FileServer(fs)

		req := httptest.NewRequest(http.MethodDelete, "http://example.com/example.txt", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		if !assert.Equal(subT, http.StatusNotFound, resp.StatusCode) {
			return
		}
	})
}

func TestUnsupportedMethod(t *testing.T) {
	fs := afero.NewMemMapFs()
	h := FileServer(fs)

	req := httptest.NewRequest("BREW", "http://example.com/example.txt", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	resp := w.Result()
	if !assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode) {
		return
	}
}
