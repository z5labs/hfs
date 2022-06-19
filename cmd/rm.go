package cmd

import (
	"github.com/z5labs/hfs/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rmCmd = &cobra.Command{
	Use:   "rm PATH",
	Short: "Remove files from an HFS server",
	Long:  ``,
	Example: `  # Remove file from HFS.
  hfs rm http://example.org/example.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := http.NewClient()

		zap.L().Debug("removing file", zap.String("path", args[0]))
		err := client.Remove(cmd.Context(), &http.RemoveRequest{
			Path: args[0],
		})
		if err != nil {
			zap.L().Fatal("failed to remove file", zap.String("path", args[0]), zap.Error(err))
		}
		zap.L().Debug("successfully removed file", zap.String("path", args[0]))
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
