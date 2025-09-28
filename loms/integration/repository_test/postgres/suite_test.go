//go:build integration
// +build integration

package repository_test

import (
	"context"
	"route256/loms/integration/containers/postgres"
	orderrepository "route256/loms/internal/adapter/repository/postgtres/order"
	outboxrepository "route256/loms/internal/adapter/repository/postgtres/outbox"
	stockrepository "route256/loms/internal/adapter/repository/postgtres/stock"
	"route256/loms/internal/domain"
	"route256/loms/internal/infra/logger"
	"route256/loms/internal/infra/metrics"
	pg "route256/loms/internal/infra/postgres"
	txmanager "route256/loms/internal/infra/tx_manager"
	"route256/loms/migration"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"go.uber.org/zap/zapcore"
)

type Suite struct {
	suite.Suite
	ctx      context.Context
	txManger *txmanager.TxManager

	orderRepo  *orderrepository.Repository
	stockRepo  *stockrepository.Repository
	outboxRepo *outboxrepository.Repository
	pools      *pg.Pools

	masterCtn  tc.Container
	replicaCtn tc.Container

	network *tc.DockerNetwork

	testData testData
}
type testData struct {
	testSku1        domain.Sku
	testReserved1   int64
	testTotalCount1 int64

	testSku2        domain.Sku
	testReserved2   int64
	testTotalCount2 int64

	testSku3        domain.Sku
	testReserved3   int64
	testTotalCount3 int64
}

func (s *Suite) BeforeAll(t provider.T) {
	s.ctx = context.Background()

	t.WithNewStep("create network", func(_ provider.StepCtx) {
		net, err := network.New(s.ctx)
		t.Require().NoError(err)
		s.network = net
	})

	t.WithNewStep("create master-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.Opt{
			Env:           postgres.MasterENV,
			ContainerName: postgres.PostgresMasterContainerName,
			Port:          postgres.PostgresMasterPort,
			Network:       s.network,
		}

		master, err := postgres.NewContainer(s.ctx, opt)
		sCtx.Require().NoError(err, "create master postgres container")
		s.masterCtn = master
	})

	t.WithNewStep("create replica-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.Opt{
			Env:           postgres.ReplicaENV,
			ContainerName: postgres.PostgresReplicaContainerName,
			Port:          postgres.PostgresReplicaPort,
			Network:       s.network,
		}

		replica, err := postgres.NewContainer(s.ctx, opt)
		sCtx.Require().NoError(err, "create replica postgres container")
		s.replicaCtn = replica
	})

	t.WithNewStep("init repository", func(sCtx provider.StepCtx) {
		err := metrics.Init(s.ctx)
		sCtx.Require().NoError(err)

		err = logger.Init(zapcore.DebugLevel)
		sCtx.Require().NoError(err)

		time.Sleep(5 * time.Second)
		pools, err := pg.New(s.ctx, postgres.MasterDSN, postgres.ReplicaDSN)
		require.NoError(t, err)

		s.pools = pools

		err = migration.RunMigrations(postgres.MasterDSN)
		require.NoError(t, err)

		s.stockRepo = stockrepository.New(s.pools)
		s.outboxRepo = outboxrepository.New(s.pools)
		s.orderRepo = orderrepository.New(s.pools, s.outboxRepo)
		s.txManger = txmanager.New(pools.Master)
	})

	t.WithNewStep("insert test stock data", func(sCtx provider.StepCtx) {
		s.testData.testSku1 = domain.Sku(42)
		s.testData.testReserved1 = int64(30)
		s.testData.testTotalCount1 = int64(100)

		s.testData.testSku2 = domain.Sku(112)
		s.testData.testReserved2 = int64(0)
		s.testData.testTotalCount2 = int64(100)

		s.testData.testSku3 = domain.Sku(29)
		s.testData.testReserved3 = int64(0)
		s.testData.testTotalCount3 = int64(150)

		const insertQuery = `
			INSERT INTO stocks (sku, total_count, reserved)
			VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9)`

		_, err := s.pools.Master.Exec(
			s.ctx, insertQuery,
			s.testData.testSku1, s.testData.testTotalCount1, s.testData.testReserved1,
			s.testData.testSku2, s.testData.testTotalCount2, s.testData.testReserved2,
			s.testData.testSku3, s.testData.testTotalCount3, s.testData.testReserved3,
		)
		sCtx.Require().NoError(err, "insert stock test data")
	})
}

func (s *Suite) AfterAll(t provider.T) {
	t.WithNewStep("stop replica-postgres container", func(sCtx provider.StepCtx) {
		err := s.replicaCtn.Terminate(s.ctx)
		sCtx.Require().NoError(err, "terminate replica-postgres container")
	})

	t.WithNewStep("stop master-postgres container", func(sCtx provider.StepCtx) {
		err := s.masterCtn.Terminate(s.ctx)
		sCtx.Require().NoError(err, "terminate master-postgres container")
	})

	t.WithNewStep("remove network", func(sCtx provider.StepCtx) {
		err := s.network.Remove(s.ctx)
		sCtx.Require().NoError(err, "network removing")
	})
}

func TestPostgresRepository(t *testing.T) {
	t.Parallel()

	suite.RunSuite(t, new(Suite))
}
