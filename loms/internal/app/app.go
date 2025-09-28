package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"route256/loms/internal/infra/closer"
	config "route256/loms/internal/infra/config"
	logger "route256/loms/internal/infra/logger"
	"route256/loms/internal/infra/metrics"
	serve "route256/loms/internal/infra/swagger"
	"route256/loms/internal/infra/tracing"
	"route256/loms/internal/middleware"
	desc "route256/loms/internal/pb/loms/v1"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rakyll/statik/fs"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	config         *config.Config
	httpAddress    string
	grpcAddress    string
	swaggerAddress string

	masterDSN  string
	replicaDSN string

	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
	swaggerServer   *http.Server
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	c, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	app := &App{
		config:         c,
		grpcAddress:    fmt.Sprintf("%v:%v", c.Service.Host, c.Service.GRPCPort),
		httpAddress:    fmt.Sprintf("%v:%v", c.Service.Host, c.Service.HTTPPort),
		swaggerAddress: fmt.Sprintf("%v:%v", c.Service.Host, c.Service.SwaggerPort),
		masterDSN:      buildPGDSN(c.MasterDB),
		replicaDSN:     buildPGDSN(c.ReplicDB),
	}

	err = app.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
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
		if err := app.runGRPCServer(ctx); err != nil {
			logger.Errorf(ctx, "grpc server error: %v", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.runHTTPServer(ctx); err != nil {
			logger.Errorf(ctx, "http server error: %v", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.runSwaggerServer(ctx); err != nil {
			logger.Errorf(ctx, "swagger server error: %v", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		daemon := app.serviceProvider.Daemon(ctx)
		daemon.Start(ctx)
		<-ctx.Done()
	}()

	gracefulShutdown(ctx, cancel, wg)

	return nil
}

func (app *App) initDeps(ctx context.Context) error {
	inits := []func(ctx context.Context) error{
		app.initMetrics,
		app.initTracing,
		app.initLogger,
		app.initServiceProvider,
		app.initGRPCServer,
		app.initSwaggerServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) initServiceProvider(_ context.Context) error {
	app.serviceProvider = newServiceProvider()
	app.serviceProvider.config = *app.config
	app.serviceProvider.masterDSN = app.masterDSN
	app.serviceProvider.replicaDSN = app.replicaDSN
	return nil
}

func (app *App) initMetrics(ctx context.Context) error {
	err := metrics.Init(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) initLogger(_ context.Context) error {
	var level zapcore.Level

	switch {
	case app.config.Service.LogLevel == "debug":
		level = zapcore.DebugLevel
	case app.config.Service.LogLevel == "info":
		level = zapcore.InfoLevel
	case app.config.Service.LogLevel == "error":
		level = zapcore.ErrorLevel
	case app.config.Service.LogLevel == "warn":
		level = zapcore.ErrorLevel
	case app.config.Service.LogLevel == "fatal":
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

func (app *App) initTracing(_ context.Context) error {
	address := fmt.Sprintf("%v:%v", app.config.Jaeger.Host, app.config.Jaeger.Port)

	err := tracing.Init(address)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) initGRPCServer(ctx context.Context) error {
	app.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.ChainUnaryInterceptor(
			middleware.ServerTracingInterceptor,
			middleware.MetricsInterceptor,
			middleware.Validate,
		),
	)

	reflection.Register(app.grpcServer)

	desc.RegisterOrdersServer(app.grpcServer, app.serviceProvider.AppHandler(ctx))
	desc.RegisterStocksServer(app.grpcServer, app.serviceProvider.AppHandler(ctx))
	desc.RegisterHealthServer(app.grpcServer, app.serviceProvider.AppHandler(ctx))

	return nil
}

func (app *App) initSwaggerServer(ctx context.Context) error {
	statikFs, err := fs.New()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(statikFs)))
	mux.HandleFunc("/orders.swagger.json", serve.SwaggerFile(ctx, "/orders.swagger.json"))

	app.swaggerServer = &http.Server{
		Addr:        app.swaggerAddress,
		Handler:     mux,
		ReadTimeout: time.Duration(app.config.Service.Timeout) * time.Second,
	}

	return nil
}

func (app *App) runGRPCServer(ctx context.Context) error {
	logger.Infof(ctx, "GRPC server is running on %s", app.grpcAddress)

	list, err := net.Listen("tcp", app.grpcAddress)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.grpcServer.Serve(list)
	}()

	select {
	case <-ctx.Done():
		logger.Infof(ctx, "GRPC server shutdown initiated")
		app.grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

func (app *App) runHTTPServer(ctx context.Context) error {
	grpcMux := runtime.NewServeMux()

	if err := desc.RegisterOrdersHandlerFromEndpoint(ctx, grpcMux, app.grpcAddress, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		return fmt.Errorf("failed to register orders gateway: %w", err)
	}

	if err := desc.RegisterStocksHandlerFromEndpoint(ctx, grpcMux, app.grpcAddress, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		return fmt.Errorf("failed to register stocks gateway: %w", err)
	}

	if err := desc.RegisterHealthHandlerFromEndpoint(ctx, grpcMux, app.grpcAddress, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		return fmt.Errorf("failed to register health gateway: %w", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/", middleware.HTTPMetrics(grpcMux))

	httpMux.HandleFunc("/debug/pprof/", pprof.Index)
	httpMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	httpMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	httpMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	httpMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	tracedHandler := middleware.TracingMiddleware(httpMux)

	server := &http.Server{
		Addr:        app.httpAddress,
		Handler:     middleware.CorsMiddleware().Handler(tracedHandler),
		ReadTimeout: time.Duration(app.config.Service.Timeout) * time.Second,
	}

	logger.Infof(ctx, "HTTP (grpc-gateway) server is running on %s", app.httpAddress)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Infof(ctx, "HTTP server shutdown initiated")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (app *App) runSwaggerServer(ctx context.Context) error {
	logger.Infof(ctx, "Swagger server is running on %s", app.swaggerAddress)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.swaggerServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Infof(ctx, "Swagger server shutdown initiated")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return app.swaggerServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func buildPGDSN(cfg config.DBConfig) string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.DBname,
	)
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
