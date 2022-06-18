package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logLevel zapcore.Level

func (l logLevel) String() string {
	return (zapcore.Level)(l).String()
}

func (l *logLevel) Set(s string) error {
	return (*zapcore.Level)(l).Set(s)
}

func (l logLevel) Type() string {
	return "Level"
}

var rootCmd = &cobra.Command{
	Use:   "hfs",
	Short: "HTTP File Server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var lvl zapcore.Level
		lvlStr := cmd.Flags().Lookup("log-level").Value.String()
		err := lvl.UnmarshalText([]byte(lvlStr))
		if err != nil {
			panic(err)
		}

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.OutputPaths = []string{viper.GetString("log-file")}
		l, err := cfg.Build(zap.IncreaseLevel(lvl))
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	},
}

func init() {
	// Persistent flags
	lvl := logLevel(zapcore.InfoLevel)
	rootCmd.PersistentFlags().Var(&lvl, "log-level", "Specify log level")
	rootCmd.PersistentFlags().String("log-file", "stderr", "Specify log file")

	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
}
