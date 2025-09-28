package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"route256/notifier/internal/infra/closer"
	"route256/notifier/internal/infra/config"
	"route256/notifier/internal/infra/logger"
	"sync"
	"syscall"

	"go.uber.org/zap/zapcore"
)

type App struct {
	config *config.Config

	serviceProvider *serviceProvider
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	c, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	app := &App{
		config: c,
	}

	err = app.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Config loaded successfully")

	return app, nil
}

func (app *App) initDeps(ctx context.Context) error {
	inits := []func(ctx context.Context) error{
		app.initLogger,
		app.initServiceProvider,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) Run(ctx context.Context) error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	ctx, cancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := app.runOrderConsumer(ctx)
		if err != nil {
			logger.Errorf(ctx, "failed to run consumer: %s", err.Error())
		}
	}()

	gracefulShutdown(ctx, cancel, wg)

	return nil
}

func (app *App) initServiceProvider(_ context.Context) error {
	app.serviceProvider = newServiceProvider()
	app.serviceProvider.config = *app.config

	return nil
}

func (app *App) initLogger(_ context.Context) error {
	var level zapcore.Level

	switch {
	case app.config.Server.LogLevel == "debug":
		level = zapcore.DebugLevel
	case app.config.Server.LogLevel == "info":
		level = zapcore.InfoLevel
	case app.config.Server.LogLevel == "error":
		level = zapcore.ErrorLevel
	case app.config.Server.LogLevel == "warn":
		level = zapcore.ErrorLevel
	case app.config.Server.LogLevel == "fatal":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.ErrorLevel
	}

	err := logger.Init(level)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) runOrderConsumer(ctx context.Context) error {
	c := app.serviceProvider.OrderConsumer(ctx)

	err := c.Consume(ctx)
	if err != nil {
		return err
	}

	return nil
}

func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Infof(ctx, "terminating: context cancelled")
	case <-waitSignal():
		logger.Infof(ctx, "terminating: via signal")
	}

	cancel()
	wg.Wait()
}

func waitSignal() chan os.Signal {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	return sigterm
}
