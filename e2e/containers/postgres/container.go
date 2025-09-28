package postgres

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresCntOpt struct {
	Env           map[string]string
	ContainerName string
	Port          string
	Network       *tc.DockerNetwork
}

var (
	MasterENV = map[string]string{
		"POSTGRESQL_DATABASE":                 "loms_db",
		"POSTGRESQL_USERNAME":                 "loms-user",
		"POSTGRESQL_PASSWORD":                 "loms-password",
		"POSTGRESQL_REPLICATION_MODE":         "master",
		"POSTGRESQL_REPLICATION_USER":         "repl_user",
		"POSTGRESQL_REPLICATION_PASSWORD":     "repl_password",
		"POSTGRESQL_SYNCHRONOUS_COMMIT_MODE":  "on",
		"POSTGRESQL_NUM_SYNCHRONOUS_REPLICAS": "1",
	}

	ReplicaENV = map[string]string{
		"POSTGRESQL_DATABASE":                 "loms_db",
		"POSTGRESQL_USERNAME":                 "loms-user",
		"POSTGRESQL_PASSWORD":                 "loms-password",
		"POSTGRESQL_REPLICATION_MODE":         "slave",
		"POSTGRESQL_REPLICATION_USER":         "repl_user",
		"POSTGRESQL_REPLICATION_PASSWORD":     "repl_password",
		"POSTGRESQL_MASTER_HOST":              PostgresMasterContainerName,
		"POSTGRESQL_MASTER_PORT_NUMBER":       "5432",
		"POSTGRESQL_SYNCHRONOUS_COMMIT_MODE":  "on",
		"POSTGRESQL_NUM_SYNCHRONOUS_REPLICAS": "1",
	}
)

const (
	postgresImage = "gitlab-registry.ozon.dev/go/classroom-18/students/base/postgres:16"

	PostgresMasterContainerName = "postgres-master"
	PostgresMasterPort          = "5433"

	PostgresReplicaContainerName = "postgres-replica"
	PostgresReplicaPort          = "5434"

	dbPort = "5432/tcp"

	MasterDSN  = "postgresql://loms-user:loms-password@localhost:5433/loms_db?sslmode=disable"
	ReplicaDSN = "postgresql://loms-user:loms-password@localhost:5434/loms_db?sslmode=disable"
)

func NewContainer(ctx context.Context, opt PostgresCntOpt) (tc.Container, error) {
	req := tc.ContainerRequest{
		Name:         opt.ContainerName,
		Image:        postgresImage,
		ExposedPorts: []string{dbPort},
		Env:          opt.Env,
		WaitingFor:   wait.ForListeningPort(nat.Port(dbPort)).WithStartupTimeout(15 * time.Second),
		Networks:     []string{opt.Network.Name},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				nat.Port(dbPort): []nat.PortBinding{{
					HostPort: opt.Port,
				}},
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
