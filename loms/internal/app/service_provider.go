package app

import (
	"context"
	syncproducer "route256/loms/internal/adapter/kafka/sync_producer"
	orderrepository "route256/loms/internal/adapter/repository/postgtres/order"
	outboxrepository "route256/loms/internal/adapter/repository/postgtres/outbox"
	stockrepository "route256/loms/internal/adapter/repository/postgtres/stock"
	api "route256/loms/internal/api/grpc/orders/handler"
	orderevent "route256/loms/internal/business/cron/order_event"
	orderservice "route256/loms/internal/business/service/order"
	stockservice "route256/loms/internal/business/service/stock"
	"route256/loms/internal/infra/closer"
	"route256/loms/internal/infra/config"
	daemon "route256/loms/internal/infra/daemon"
	logger "route256/loms/internal/infra/logger"
	pgpool "route256/loms/internal/infra/postgres"
	txmanager "route256/loms/internal/infra/tx_manager"
	"time"
)

type serviceProvider struct {
	config config.Config

	masterDSN  string
	replicaDSN string

	orderRepository  *orderrepository.Repository
	stockRepository  *stockrepository.Repository
	outboxRepository *outboxrepository.Repository
	connPools        *pgpool.Pools

	txManagerMaster  *txmanager.TxManager
	txManagerReplica *txmanager.TxManager

	stockService       *stockservice.Service
	eventCronProcessor *orderevent.CronProcessor
	daemon             *daemon.Daemon
	orderService       *orderservice.Service

	appServer *api.Implementation

	kafkaProducer *syncproducer.Producer
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (srv *serviceProvider) PostgresPools(ctx context.Context) *pgpool.Pools {
	if srv.connPools != nil {
		return srv.connPools
	}

	pools, err := pgpool.New(ctx, srv.masterDSN, srv.replicaDSN)
	if err != nil {
		logger.Fatalf(ctx, "pgpool.NewPools: failed to connect to db %v", err)
	}

	closer.Add(func() error {
		pools.Master.Close()
		return nil
	})

	closer.Add(func() error {
		pools.Replica.Close()
		return nil
	})

	srv.connPools = pools
	return pools
}

func (srv *serviceProvider) AppKafkaProducer(ctx context.Context) *syncproducer.Producer {
	if srv.kafkaProducer == nil {
		var err error
		srv.kafkaProducer, err = syncproducer.New(ctx, srv.config)

		if err != nil {
			logger.Fatalf(ctx, "syncproducer.New: failed to run producer %v", err)
		}

		closer.Add(func() error {
			srv.kafkaProducer.Close(ctx)
			return nil
		})
	}

	return srv.kafkaProducer
}

func (srv *serviceProvider) OutboxRepository(ctx context.Context) *outboxrepository.Repository {
	if srv.outboxRepository == nil {
		srv.outboxRepository = outboxrepository.New(
			srv.PostgresPools(ctx),
		)
	}

	return srv.outboxRepository
}

func (srv *serviceProvider) OrderRepository(ctx context.Context) *orderrepository.Repository {
	if srv.orderRepository == nil {
		srv.orderRepository = orderrepository.New(
			srv.PostgresPools(ctx),
			srv.OutboxRepository(ctx),
		)
	}

	return srv.orderRepository
}

func (srv *serviceProvider) StockRepository(ctx context.Context) *stockrepository.Repository {
	if srv.stockRepository == nil {
		srv.stockRepository = stockrepository.New(
			srv.PostgresPools(ctx),
		)
	}

	return srv.stockRepository
}

func (srv *serviceProvider) TxManagerMaster(_ context.Context) *txmanager.TxManager {
	if srv.txManagerMaster == nil {
		srv.txManagerMaster = txmanager.New(srv.connPools.Master)
	}

	return srv.txManagerMaster
}

func (srv *serviceProvider) TxManagerReplica(_ context.Context) *txmanager.TxManager {
	if srv.txManagerReplica == nil {
		srv.txManagerReplica = txmanager.New(srv.connPools.Replica)
	}

	return srv.txManagerReplica
}

func (srv *serviceProvider) AppStockService(ctx context.Context) *stockservice.Service {
	if srv.stockService == nil {
		srv.stockService = stockservice.New(
			srv.StockRepository(ctx),
		)
	}

	return srv.stockService
}

func (srv *serviceProvider) EventCronProcessor(ctx context.Context) *orderevent.CronProcessor {
	if srv.eventCronProcessor == nil {
		srv.eventCronProcessor = orderevent.New(
			srv.outboxRepository,
			srv.AppKafkaProducer(ctx),
			srv.txManagerMaster,
			srv.config.Service.LimitOutboxMsg,
		)
	}

	return srv.eventCronProcessor
}

func (srv *serviceProvider) Daemon(ctx context.Context) *daemon.Daemon {
	if srv.daemon == nil {
		srv.daemon = daemon.New(
			srv.EventCronProcessor(ctx),
			time.Duration(srv.config.Service.HandlePeriod)*time.Second,
		)
	}

	return srv.daemon
}

func (srv *serviceProvider) AppOrderService(ctx context.Context) *orderservice.Service {
	if srv.orderService == nil {
		srv.orderService = orderservice.New(
			srv.config.Kafka.OrderTopic,
			srv.OrderRepository(ctx),
			srv.OutboxRepository(ctx),
			srv.AppStockService(ctx),
			srv.TxManagerMaster(ctx),
		)
	}

	return srv.orderService
}

func (srv *serviceProvider) AppHandler(ctx context.Context) *api.Implementation {
	if srv.appServer == nil {
		srv.appServer = api.NewImplementation(
			srv.AppOrderService(ctx),
			srv.AppStockService(ctx),
		)
	}

	return srv.appServer
}
