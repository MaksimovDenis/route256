//go:build integration
// +build integration

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	cartclient "route256/homework/e2e/clients/cart"
	lomsclient "route256/homework/e2e/clients/loms"
	model "route256/homework/e2e/clients/model/cart"
	cartcontainers "route256/homework/e2e/containers/cart"
	"route256/homework/e2e/containers/kafka"
	lomscontainers "route256/homework/e2e/containers/loms"
	productcontainers "route256/homework/e2e/containers/poduct_service"
	"route256/homework/e2e/containers/postgres"
	cartdata "route256/homework/e2e/test_data/cart"
	"route256/loms/migration"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

type cartSuite struct {
	suite.Suite
	ctx context.Context

	productCtn tc.Container
	cartCtn    tc.Container
	lomsCtn    tc.Container
	masterCtn  tc.Container
	replicaCtn tc.Container
	kafkaCtn   tc.Container
	network    *tc.DockerNetwork

	cartClient *cartclient.Client
	lomsClient *lomsclient.Client

	skuBook  uint64
	skuChef  uint64
	itemBook model.Item
	itemChef model.Item
}

func TestCartService(t *testing.T) {
	suite.RunSuite(t, new(cartSuite))
}

func (c *cartSuite) BeforeAll(t provider.T) {
	c.ctx = context.Background()
	c.skuBook = cartdata.SkuBook
	c.skuChef = cartdata.SkuChef
	c.itemBook = cartdata.ItemBook
	c.itemChef = cartdata.ItemChef

	t.WithNewStep("create network", func(sCtx provider.StepCtx) {
		net, err := network.New(c.ctx)
		t.Require().NoError(err)
		c.network = net
	})

	t.WithNewStep("create master-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.PostgresCntOpt{
			Env:           postgres.MasterENV,
			ContainerName: postgres.PostgresMasterContainerName,
			Port:          postgres.PostgresMasterPort,
			Network:       c.network,
		}

		master, err := postgres.NewContainer(c.ctx, opt)
		sCtx.Require().NoError(err, "create master postgres container")
		c.masterCtn = master
	})

	t.WithNewStep("create replica-postgres container", func(sCtx provider.StepCtx) {
		opt := postgres.PostgresCntOpt{
			Env:           postgres.ReplicaENV,
			ContainerName: postgres.PostgresReplicaContainerName,
			Port:          postgres.PostgresReplicaPort,
			Network:       c.network,
		}

		replica, err := postgres.NewContainer(c.ctx, opt)
		sCtx.Require().NoError(err, "create replica postgres container")
		c.replicaCtn = replica
	})

	t.WithNewStep("create kafka container", func(sCtx provider.StepCtx) {
		opt := kafka.KafkaCntOpt{
			Env:           kafka.KafkaENV,
			ContainerName: kafka.ContainerName,
			Port:          kafka.Port,
			Network:       c.network,
		}

		kafka, err := kafka.NewContainer(c.ctx, opt)
		sCtx.Require().NoError(err, "create kafka container")
		c.kafkaCtn = kafka
	})

	t.WithNewStep("start product service", func(sCtx provider.StepCtx) {
		product, err := productcontainers.NewContainer(c.ctx, c.network)
		sCtx.Require().NoError(err, "create product container")
		c.productCtn = product
	})

	t.WithNewStep("run loms migrations", func(sCtx provider.StepCtx) {
		err := migration.RunMigrations(postgres.MasterDSN)
		sCtx.Require().NoError(err, "failed to up migration for loms")
	})

	t.WithNewStep("start loms service", func(sCtx provider.StepCtx) {
		loms, err := lomscontainers.NewContainer(c.ctx, c.network)
		sCtx.Require().NoError(err, "create loms container")
		c.lomsCtn = loms
	})

	t.WithNewStep("create http loms client", func(sCtx provider.StepCtx) {
		address := fmt.Sprintf("%s:%s", "localhost", lomscontainers.LomsSvcHTTPPort)
		c.lomsClient = lomsclient.New("http://" + address)
	})

	t.WithNewStep("health check loms service", func(sCtx provider.StepCtx) {
		resp, err := c.lomsClient.HealthCheck(c.ctx)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(resp.Data.Message, "OK")

	})

	t.WithNewStep("start cart service", func(sCtx provider.StepCtx) {
		product, err := cartcontainers.NewContainer(c.ctx, c.network)
		sCtx.Require().NoError(err, "create cart container")
		c.cartCtn = product
	})

	t.WithNewStep("create http cart client", func(sCtx provider.StepCtx) {
		address := fmt.Sprintf("%s:%s", "localhost", cartcontainers.CartSvcPort)
		c.cartClient = cartclient.New("http://"+address, cartdata.Token)
	})
}

func (c *cartSuite) AfterAll(t provider.T) {
	t.WithNewStep("stop loms service", func(sCtx provider.StepCtx) {
		err := c.lomsCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate loms container")
	})

	t.WithNewStep("stop product service", func(sCtx provider.StepCtx) {
		err := c.productCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate product container")
	})

	t.WithNewStep("stop cart service", func(sCtx provider.StepCtx) {
		err := c.cartCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate cart container")
	})

	t.WithNewStep("stop replica-postgres container", func(sCtx provider.StepCtx) {
		err := c.replicaCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate replica-postgres container")
	})

	t.WithNewStep("stop master-postgres container", func(sCtx provider.StepCtx) {
		err := c.masterCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate master-postgres container")
	})

	t.WithNewStep("stop kafka container", func(sCtx provider.StepCtx) {
		err := c.kafkaCtn.Terminate(c.ctx)
		sCtx.Require().NoError(err, "terminate kafka container")
	})

	t.WithNewStep("remove network", func(sCtx provider.StepCtx) {
		err := c.network.Remove(c.ctx)
		sCtx.Require().NoError(err, "network removing")
	})
}

func (c *cartSuite) TestCartHandlers(t provider.T) {
	userID := uint64(time.Now().UnixNano())

	t.WithParameters(
		&allure.Parameter{
			Name:  "user_id",
			Value: userID,
		},
		&allure.Parameter{
			Name:  "sku_book",
			Value: cartdata.SkuBook,
		},
		&allure.Parameter{
			Name:  "sku_chef",
			Value: cartdata.SkuChef,
		},
	)

	t.WithNewStep("add items to cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.AddItem(c.ctx, userID, c.itemBook)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.StatusCode)

		resp, err = c.cartClient.AddItem(c.ctx, userID, c.itemChef)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	t.WithNewStep("get items from cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.GetItemsByUserID(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)

		want := cartclient.GetCartResp{
			Items: []cartclient.Item{
				{Sku: 1076963, Name: "Теория нравственных чувств | Смит Адам", Count: 2, Price: 3379},
				{Sku: 1148162, Name: "Кулинар Гуров", Count: 1, Price: 2931},
			},
			TotalPrice: 9689,
		}

		sCtx.Require().Len(resp.Data.Items, 2)
		sCtx.Require().Equal(want, resp.Data)
	})

	t.WithNewStep("delete one item from cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.DeleteItem(c.ctx, userID, c.skuBook)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusNoContent, resp.StatusCode)
	})

	t.WithNewStep("get items from cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.GetItemsByUserID(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)

		want := cartclient.GetCartResp{
			Items: []cartclient.Item{
				{Sku: 1148162, Name: "Кулинар Гуров", Count: 1, Price: 2931},
			},
			TotalPrice: 2931,
		}

		sCtx.Require().Len(resp.Data.Items, 1)
		sCtx.Require().Equal(want, resp.Data)
	})

	t.WithNewStep("delete cart by userID", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.DeleteCartByUserID(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusNoContent, resp.StatusCode)
	})

	t.WithNewStep("get items from cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.GetItemsByUserID(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusNotFound, resp.HTTPResp.StatusCode)
	})

	t.WithNewStep("add items to cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.AddItem(c.ctx, userID, c.itemChef)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	t.WithNewStep("checkout items", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.Checkout(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusOK, resp.HTTPResp.StatusCode)
	})

	t.WithNewStep("get items from cart", func(sCtx provider.StepCtx) {
		resp, err := c.cartClient.GetItemsByUserID(c.ctx, userID)
		sCtx.Require().NoError(err)
		sCtx.Require().Equal(http.StatusNotFound, resp.HTTPResp.StatusCode)
	})
}
