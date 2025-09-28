package main

import (
	"context"
	"os"
	"route256/loms/internal/app"
	"route256/loms/internal/infra/logger"
	_ "route256/loms/statik"
)

func main() {
	ctx := context.Background()

	lomsApp, err := app.NewApp(ctx, os.Getenv("CONFIG_FILE"))
	if err != nil {
		logger.Fatalf(ctx, "failed to init app: %s", err)
	}

	err = lomsApp.Run(ctx)
	if err != nil {
		logger.Fatalf(ctx, "failed to run app: %s", err)
	}
}
