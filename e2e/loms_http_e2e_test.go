//go:build integration
// +build integration

package e2e_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"route256/homework/e2e/containers/kafka"
	lomscontainers "route256/homework/e2e/containers/loms"
	"route256/homework/e2e/containers/postgres"
	lomsdata "route256/homework/e2e/test_data/loms"
	"route256/loms/migration"
	"strconv"
	"testing"
	"time"

	lomsclient "route256/homework/e2e/clients/loms"
	model "route256/homework/e2e/clients/model/loms"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

type httpLomsSuite struct {
	suite.Suite
	ctx context.Context

	lomsCtn    tc.Container
	masterCtn  tc.Container
	replicaCtn tc.Container
	kafkaCtn   tc.Container
	network    *tc.DockerNetwork

	httpClient lomsclient.Client
}

func TestLomsHTTPService(t *testing.T) {
	suite.RunSuite(t, new(httpLomsSuite))
}

func (l *httpLomsSuite) BeforeAll(t provider.T) {
	l.ctx = context.Background()

	t.WithNewStep("create network", func(sCtx provider.StepCtx) {
		net, err := network.New(l.ctx)
		t.Require().NoError(err)
		l.network = net
	})

	t.WithNewStep("create master-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.PostgresCntOpt{
			Env:           postgres.MasterENV,
			ContainerName: postgres.PostgresMasterContainerName,
			Port:          postgres.PostgresMasterPort,
			Network:       l.network,
		}

		master, err := postgres.NewContainer(l.ctx, opt)
		sCtx.Require().NoError(err, "create master postgres container")
		l.masterCtn = master
	})

	t.WithNewStep("create replica-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.PostgresCntOpt{
			Env:           postgres.ReplicaENV,
			ContainerName: postgres.PostgresReplicaContainerName,
			Port:          postgres.PostgresReplicaPort,
			Network:       l.network,
		}

		replica, err := postgres.NewContainer(l.ctx, opt)
		sCtx.Require().NoError(err, "create replica postgres container")
		l.replicaCtn = replica
	})

	t.WithNewStep("run loms migrations", func(sCtx provider.StepCtx) {
		err := migration.RunMigrations(postgres.MasterDSN)
		sCtx.Require().NoError(err, "failed to up migration for loms")
	})

	t.WithNewStep("create kafka container", func(sCtx provider.StepCtx) {
		opt := kafka.KafkaCntOpt{
			Env:           kafka.KafkaENV,
			ContainerName: kafka.ContainerName,
			Port:          kafka.Port,
			Network:       l.network,
		}

		kafka, err := kafka.NewContainer(l.ctx, opt)
		sCtx.Require().NoError(err, "create kafka container")
		l.kafkaCtn = kafka
	})

	t.WithNewStep("start loms service", func(sCtx provider.StepCtx) {
		loms, err := lomscontainers.NewContainer(l.ctx, l.network)
		sCtx.Require().NoError(err, "create loms container")
		l.lomsCtn = loms
	})

	t.WithNewStep("create http loms client", func(sCtx provider.StepCtx) {
		address := fmt.Sprintf("%s:%s", "localhost", lomscontainers.LomsSvcHTTPPort)
		l.httpClient = *lomsclient.New("http://" + address)
	})

	t.WithNewStep("health check loms service", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.HealthCheck(l.ctx)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(resp.Data.Message, "OK")
	})
}

func (l *httpLomsSuite) AfterAll(t provider.T) {
	t.WithNewStep("stop loms container", func(sCtx provider.StepCtx) {
		err := l.lomsCtn.Terminate(l.ctx)
		sCtx.Require().NoError(err, "terminate loms-svc container")
	})

	t.WithNewStep("stop replica-postgres container", func(sCtx provider.StepCtx) {
		err := l.replicaCtn.Terminate(l.ctx)
		sCtx.Require().NoError(err, "terminate replica-postgres container")
	})

	t.WithNewStep("stop master-postgres container", func(sCtx provider.StepCtx) {
		err := l.masterCtn.Terminate(l.ctx)
		sCtx.Require().NoError(err, "terminate master-postgres container")
	})

	t.WithNewStep("stop kafka container", func(sCtx provider.StepCtx) {
		err := l.kafkaCtn.Terminate(l.ctx)
		sCtx.Require().NoError(err, "terminate kafka container")
	})

	t.WithNewStep("remove network", func(sCtx provider.StepCtx) {
		err := l.network.Remove(l.ctx)
		sCtx.Require().NoError(err, "network removing")
	})
}

func (l *httpLomsSuite) TestLomsHTTPHandlers(t provider.T) {
	t.Parallel()

	userID := rand.Int64()

	testOrder := model.Order{
		UserID: userID,
		Items:  lomsdata.TestOrder.Items,
	}

	testOrder2 := model.Order{
		UserID: userID,
		Items:  lomsdata.TestOrder2.Items,
	}

	t.WithParameters(
		&allure.Parameter{Name: "user_id", Value: userID},
		&allure.Parameter{Name: "test_order", Value: testOrder},
		&allure.Parameter{Name: "test_order_2", Value: testOrder2},
	)

	t.Run("OrderCreateAndPay", func(t provider.T) {
		l.testOrderCreateAndPay(t, testOrder, userID)
	})

	t.Run("OrderCreateAndCancel", func(t provider.T) {
		l.testOrderCreateAndCacnel(t, testOrder, userID)
	})

	t.Run("StockInfo", func(t provider.T) {
		l.testStocksInfoCheck(t, testOrder.Items[0].Sku)
	})
}

func (l *httpLomsSuite) testOrderCreateAndPay(t provider.T, order model.Order, userID int64) {
	var orderID string

	t.WithNewStep("create a new order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.CreateOrder(l.ctx, order)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(resp.Data.OrderID, "1")

		orderID = resp.Data.OrderID
	})

	t.WithNewStep("get info about created order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderInfo(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(strconv.FormatInt(userID, 10), resp.Data.UserID)
		sCtx.Require().Equal(model.OrderStatusAwaitingPayment, resp.Data.Status)
		for i, item := range resp.Data.Items {
			sCtx.Require().Equal(strconv.FormatInt(int64(order.Items[i].Sku), 10), item.Sku)
			sCtx.Require().Equal(order.Items[i].Count, item.Count)
		}
	})

	t.WithNewStep("pay order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderPay(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	t.WithNewStep("check kafka message after payment", func(sCtx provider.StepCtx) {
		t.Parallel()
		groupID := fmt.Sprintf("test-group-%d", rand.IntN(10000))

		messages, err := kafka.ConsumeKafkaMessages(kafka.Broker, kafka.Topic, groupID, 30*time.Second)
		sCtx.Require().NoError(err, "failed to consume kafka messages")
		sCtx.Require().NotEmpty(messages, "no messages received from kafka")

		var found bool
		for _, msg := range messages {
			if string(msg.Value) != "" && string(msg.Key) == orderID {
				found = true
				break
			}
		}

		sCtx.Require().True(found, fmt.Sprintf("Kafka message with status=%s not found", model.OrderStatusPayed))
	})

	t.WithNewStep("get info about payed order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderInfo(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(strconv.FormatInt(userID, 10), resp.Data.UserID)
		sCtx.Require().Equal(model.OrderStatusPayed, resp.Data.Status)
		for i, item := range resp.Data.Items {
			sCtx.Require().Equal(strconv.FormatInt(int64(order.Items[i].Sku), 10), item.Sku)
			sCtx.Require().Equal(order.Items[i].Count, item.Count)
		}
	})

	t.WithNewStep("try to cancel payed order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderCancel(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (l *httpLomsSuite) testOrderCreateAndCacnel(t provider.T, order model.Order, userID int64) {
	var orderID string

	t.WithNewStep("create a new order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.CreateOrder(l.ctx, order)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(resp.Data.OrderID, "2")

		orderID = resp.Data.OrderID
	})

	t.WithNewStep("cancel order", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderCancel(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	t.WithNewStep("check cancelled order info", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.OrderInfo(l.ctx, orderID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(strconv.FormatInt(userID, 10), resp.Data.UserID)
		sCtx.Require().Equal(model.OrderStatusCancelled, resp.Data.Status)
		for i, item := range resp.Data.Items {
			sCtx.Require().Equal(strconv.FormatInt(int64(order.Items[i].Sku), 10), item.Sku)
			sCtx.Require().Equal(order.Items[i].Count, item.Count)
		}
	})
}

func (l *httpLomsSuite) testStocksInfoCheck(t provider.T, sku model.Sku) {
	t.WithNewStep("get info about stocks", func(sCtx provider.StepCtx) {
		resp, err := l.httpClient.StockInfo(l.ctx, sku)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
		sCtx.Require().Equal(int64(62), resp.Data.Count)
	})
}
