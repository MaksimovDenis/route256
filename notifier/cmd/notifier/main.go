package main

import (
	"context"
	"os"
	"route256/notifier/internal/app"
	"route256/notifier/internal/infra/logger"
)

func main() {
	ctx := context.Background()

	notifierApp, err := app.NewApp(ctx, os.Getenv("CONFIG_FILE"))
	if err != nil {
		logger.Fatalf(ctx, "failed to init app: %s", err)
	}

	err = notifierApp.Run(ctx)
	if err != nil {
		logger.Fatalf(ctx, "failed to run app: %s", err)
	}
}
