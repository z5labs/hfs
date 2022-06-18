package cmd

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap"
)

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		zap.L().Fatal("unexpected error", zap.Error(err))
	}
}
