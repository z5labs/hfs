package cmd

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"time"

	hfshttp "github.com/z5labs/hfs/http"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve ROOT_DIR",
	Short: "Start an HTTP File Server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		var err error
		root, err = filepath.Abs(root)
		if err != nil {
			zap.L().Fatal("invalid root filepath", zap.Error(err))
		}

		// Cache heavily sought after files in memory
		fs := afero.NewBasePathFs(afero.NewOsFs(), root)
		h := hfshttp.FileServer(fs)
		s := &http.Server{ // TODO: add more config
			Handler: h,
		}

		// start up http server and begin serving on addr
		errCh := make(chan error, 1)
		go func() {
			defer close(errCh)

			addr := viper.GetString("addr")
			ls, err := net.Listen("tcp", addr)
			if err != nil {
				errCh <- err
				return
			}
			zap.L().Info("hfs is now ready for requests", zap.String("addr", ls.Addr().String()), zap.String("root", root))

			err = s.Serve(ls)
			if err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()

		// block until shutdown or unexpected error
		select {
		case <-cmd.Context().Done():
			timeout := 5 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			zap.L().Info("shutting down", zap.Duration("timeout", timeout))
			err := s.Shutdown(ctx)
			if err != nil {
				zap.L().Fatal("unexpected error while shutting down", zap.Error(err))
			}
			zap.L().Info("successfully shutdown")
		case err := <-errCh:
			zap.L().Fatal("unexpected error", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Local flags
	serveCmd.Flags().String("addr", ":8080", "Address for HFS to listen on")

	// Viper
	viper.BindPFlag("addr", serveCmd.Flags().Lookup("addr"))
	viper.BindPFlag("root", serveCmd.Flags().Lookup("root"))
}
