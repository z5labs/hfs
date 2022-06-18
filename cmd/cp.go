package cmd

import (
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

  # Copy from an HFS server to stdout.
  hfs cp http://example.org/hello.txt -
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		zap.L().Info("hello")
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
