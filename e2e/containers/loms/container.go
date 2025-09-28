package loms

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	lomsContainerName = "loms"
	LomsSvcGRPCPort   = "8083"
	LomsSvcHTTPPort   = "8084"
)

func NewContainer(ctx context.Context, network *tc.DockerNetwork) (tc.Container, error) {
	grpcPort, err := nat.NewPort("tcp", LomsSvcGRPCPort)
	if err != nil {
		return nil, err
	}

	httpPort, err := nat.NewPort("tcp", LomsSvcHTTPPort)
	if err != nil {
		return nil, err
	}

	req := tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context: "../loms",
		},
		Name:         lomsContainerName,
		ExposedPorts: []string{grpcPort.Port(), httpPort.Port()},
		WaitingFor:   wait.ForListeningPort(LomsSvcHTTPPort).WithStartupTimeout(15 * time.Second),
		Networks:     []string{network.Name},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				grpcPort: []nat.PortBinding{{HostPort: LomsSvcGRPCPort}},
				httpPort: []nat.PortBinding{{HostPort: LomsSvcHTTPPort}},
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
