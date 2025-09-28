package product_service

import (
	"context"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	productServiceImage  = "gitlab-registry.ozon.dev/go/classroom-18/students/base/products:latest"
	productContainerName = "product-service"
	productSvcPort       = "8082"
)

func NewContainer(ctx context.Context, network *tc.DockerNetwork) (tc.Container, error) {
	port, err := nat.NewPort("tcp", productSvcPort)
	if err != nil {
		return nil, err
	}

	req := tc.ContainerRequest{
		Image:        productServiceImage,
		Name:         productContainerName,
		ExposedPorts: []string{port.Port()},
		WaitingFor:   wait.ForListeningPort(port).WithStartupTimeout(10 * time.Second),
		Networks:     []string{network.Name},
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}
