package http

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type fileHandler struct {
	fs afero.Fs
}

// FileServer
func FileServer(root afero.Fs) http.Handler {
	return &fileHandler{
		fs: root,
	}
}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	zap.L().Info("received request", zap.String("method", req.Method), zap.String("path", req.URL.Path))
	defer zap.L().Info("response sent", zap.String("method", req.Method), zap.String("path", req.URL.Path))

	switch req.Method {
	case http.MethodGet:
		getFile(w, h.fs, req.URL.Path)
	case http.MethodPost:
		upsertFile(w, h.fs, req.URL.Path, req.Body)
	case http.MethodDelete:
		deleteFile(w, h.fs, req.URL.Path)
	default:
		zap.L().Error("received request with unsupported method", zap.String("method", req.Method))
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
	}
}

func getFile(w http.ResponseWriter, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	if err != nil {
		zap.L().Error("unexpected error when checking if file exists", zap.String("path", path), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		zap.L().Warn("file not found", zap.String("path", path))
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	f, err := fs.Open(path)
	if err != nil {
		zap.L().Error("unexpected error when opening file", zap.String("path", path), zap.Error(err))
		http.Error(w, "failed to open file", http.StatusInternalServerError)
		return
	}

	// TODO: Respect Content-Length headers for only retrieving parts of a file
	n, err := copy(w, f)
	if err != nil {
		zap.L().Error("unexpected error when writing file to response", zap.Error(err))
		return
	}
	zap.L().Info("wrote file to response", zap.String("path", path), zap.Int64("total_bytes", n))
}

func upsertFile(w http.ResponseWriter, fs afero.Fs, path string, body io.Reader) {
	f, existed, err := openFile(fs, path)
	if err != nil {
		zap.L().Error("unexpected error when getting file", zap.String("path", path), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = f.Truncate(0)
	if err != nil {
		zap.L().Error("failed to truncate file", zap.String("path", path), zap.Error(err))
		http.Error(w, "failed to truncate file", http.StatusInternalServerError)
		return
	}

	if !existed {
		w.WriteHeader(http.StatusCreated)
	}
	n, err := copy(f, body)
	if err != nil {
		zap.L().Error("unexpected error while copying contents", zap.String("path", path), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = f.Sync()
	if err != nil {
		zap.L().Error("failed to sync file", zap.String("path", path), zap.Error(err))
		http.Error(w, "failed to sync file", http.StatusInternalServerError)
		return
	}
	zap.L().Info("successfully saved file", zap.String("path", path), zap.Int64("total_bytes", n))
}

func deleteFile(w http.ResponseWriter, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	if err != nil {
		zap.L().Error("unexpected error when checking if file exists", zap.String("path", path), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		zap.L().Warn("file not found", zap.String("path", path))
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	err = fs.Remove(path)
	if err != nil {
		zap.L().Error("unexpected error when removing file", zap.String("path", path), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	zap.L().Info("successfully deleted file", zap.String("path", path))
}

func openFile(fs afero.Fs, path string) (afero.File, bool, error) {
	exists, err := afero.Exists(fs, path)
	if err != nil {
		zap.L().Error("unexpected error when checking if file exists", zap.String("path", path), zap.Error(err))
		return nil, false, err
	}
	if exists {
		f, err := fs.OpenFile(path, os.O_RDWR, os.ModePerm)
		return f, true, err
	}

	err = fs.MkdirAll(filepath.Dir(path), os.ModeDir)
	if err != nil {
		zap.L().Error("unexpected error when creating parent directories in path", zap.String("path", path), zap.Error(err))
		return nil, true, err
	}
	f, err := fs.Create(path)
	return f, false, err
}

func copyAndClose(w io.Writer, rc io.ReadCloser) (int64, error) {
	defer rc.Close()

	return io.Copy(w, rc)
}

func copy(w io.Writer, r io.Reader) (int64, error) {
	if rc, ok := r.(io.ReadCloser); ok {
		return copyAndClose(w, rc)
	}
	return io.Copy(w, r)
}
