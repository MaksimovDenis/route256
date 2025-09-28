package main

import (
	"context"
	"os"
	application "route256/cart/internal/app"
	logger "route256/cart/internal/infra/logger"
)

func main() {
	ctx := context.Background()

	app, err := application.NewApp(ctx, os.Getenv("CONFIG_FILE"))
	if err != nil {
		logger.Fatalf(ctx, "failed to init app: %s", err)
	}

	err = app.Run(ctx)
	if err != nil {
		logger.Fatalf(ctx, "failed to run app: %s", err)
	}
}
