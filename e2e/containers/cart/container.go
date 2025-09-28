package cart

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	cartContainerName = "cart"
	CartSvcPort       = "8080"
)

func NewContainer(ctx context.Context, network *tc.DockerNetwork) (tc.Container, error) {
	port, err := nat.NewPort("tcp", CartSvcPort)
	if err != nil {
		return nil, err
	}

	req := tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context: "../cart",
		},
		Name:         cartContainerName,
		ExposedPorts: []string{port.Port()},
		WaitingFor:   wait.ForListeningPort(port).WithStartupTimeout(10 * time.Second),
		Networks:     []string{network.Name},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				port: []nat.PortBinding{{HostPort: CartSvcPort}},
			}
		},
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
