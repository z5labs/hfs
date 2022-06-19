package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/z5labs/hfs/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cpCmd = &cobra.Command{
	Use:   "cp SOURCE DEST",
	Short: "Copy files to and from an HFS server.",
	Long:  ``,
	Example: `  # Copy from an HFS server to current working directory.
  hfs cp http://example.org/hello.txt .

  # Copy from current working directory to an HFS server.
  hfs cp hello.txt http://example.org/
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := http.NewClient()
		fromHFS := strings.Contains(args[0], "http")
		toHFS := strings.Contains(args[1], "http")

		// Remote to remote
		if fromHFS && toHFS {
			src, err := download(cmd.Context(), client, args[0])
			if err != nil {
				zap.L().Fatal("failed to download file", zap.String("src", args[0]), zap.Error(err))
			}

			err = upload(cmd.Context(), client, args[1], src)
			if err != nil {
				zap.L().Fatal("failed to upload file", zap.String("dst", args[1]), zap.Error(err))
			}
			return
		}

		// Remote to local
		if fromHFS {
			dst, err := openFile(args[1])
			if err != nil {
				zap.L().Fatal("failed to destination file", zap.String("dst", args[1]), zap.Error(err))
			}

			src, err := download(cmd.Context(), client, args[0])
			if err != nil {
				zap.L().Fatal("failed to download file", zap.String("src", args[0]), zap.Error(err))
			}

			_, err = io.Copy(dst, src)
			if err != nil {
				zap.L().Fatal(
					"failed to copy from source to destination",
					zap.String("src", args[0]),
					zap.String("dst", args[1]),
					zap.Error(err),
				)
			}
			err = dst.Sync()
			if err != nil {
				zap.L().Fatal("failed to sync destination file", zap.String("dst", args[1]), zap.Error(err))
			}
			return
		}

		// Local to remote
		src, err := openFile(args[0])
		if err != nil {
			zap.L().Fatal("failed to source file", zap.String("src", args[0]), zap.Error(err))
		}

		err = upload(cmd.Context(), client, args[1], src)
		if err != nil {
			zap.L().Fatal("failed to upload file", zap.String("dst", args[1]), zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}

func download(ctx context.Context, c *http.FileServerClient, path string) (io.Reader, error) {
	resp, err := c.Download(ctx, &http.DownloadRequest{
		Path: path,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func upload(ctx context.Context, c *http.FileServerClient, path string, src io.Reader) error {
	return c.Upload(ctx, &http.UploadRequest{
		Path:    path,
		Content: src,
	})
}

func openFile(path string) (*os.File, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
}
