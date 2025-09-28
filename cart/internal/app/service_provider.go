package app

import (
	"context"
	"fmt"
	"net/http"
	lomsclient "route256/cart/internal/adapter/client/loms"
	productclient "route256/cart/internal/adapter/client/product_service"
	cartrepository "route256/cart/internal/adapter/repository/cart"
	api "route256/cart/internal/api/http/handler"
	cartcron "route256/cart/internal/business/cron/cart"
	cartservice "route256/cart/internal/business/service/cart"
	config "route256/cart/internal/infra/config"
	daemon "route256/cart/internal/infra/daemon"
	"time"

	"github.com/go-playground/validator"
)

type serviceProvider struct {
	config config.Config

	appRepositroy *cartrepository.Repository

	appService        *cartservice.Service
	cartCronProcessor *cartcron.CronProcessor
	daemon            *daemon.Daemon

	productClient     *productclient.Client
	httpProductClient *http.Client

	lomsClient *lomsclient.Client

	appServer *api.Server
	validator *validator.Validate
}

func newServiceProvider() *serviceProvider {
	srv := &serviceProvider{}

	srv.validator = validator.New()

	return srv
}

func (srv *serviceProvider) AppRepository(_ context.Context) *cartrepository.Repository {
	if srv.appRepositroy == nil {
		srv.appRepositroy = cartrepository.New(srv.config.Server.CartCap)
	}

	return srv.appRepositroy
}

func (srv *serviceProvider) AppProductClient(_ context.Context) *productclient.Client {
	address := fmt.Sprintf("%s:%s", srv.config.ProductService.Host, srv.config.ProductService.Port)

	if srv.productClient == nil {
		srv.productClient = productclient.New(
			*srv.httpProductClient,
			srv.config.ProductService.Token,
			address,
		)
	}

	return srv.productClient
}

func (srv *serviceProvider) AppService(ctx context.Context) *cartservice.Service {
	if srv.appService == nil {
		srv.appService = cartservice.New(
			srv.AppRepository(ctx),
			srv.AppProductClient(ctx),
			srv.lomsClient,
			srv.config.Server.Workers,
		)
	}
	return srv.appService
}

func (srv *serviceProvider) AppHandler(ctx context.Context) *api.Server {
	if srv.appServer == nil {
		srv.appServer = api.New(
			srv.AppService(ctx),
			srv.validator,
		)
	}
	return srv.appServer
}

func (srv *serviceProvider) CartCronProcessor(ctx context.Context) *cartcron.CronProcessor {
	if srv.cartCronProcessor == nil {
		srv.cartCronProcessor = cartcron.New(
			srv.AppRepository(ctx),
		)
	}

	return srv.cartCronProcessor
}

func (srv *serviceProvider) Daemon(ctx context.Context) *daemon.Daemon {
	if srv.daemon == nil {
		srv.daemon = daemon.New(
			srv.CartCronProcessor(ctx),
			time.Duration(srv.config.Server.CheckStorageInterval)*time.Second,
		)
	}

	return srv.daemon
}
