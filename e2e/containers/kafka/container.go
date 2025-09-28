package kafka

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type KafkaCntOpt struct {
	Env           map[string]string
	ContainerName string
	Port          string
	Network       *tc.DockerNetwork
}

var (
	KafkaENV = map[string]string{
		"KAFKA_NODE_ID":                                  "1",
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
		"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092",
		"KAFKA_LISTENERS":                                "PLAINTEXT://kafka:29092,CONTROLLER://kafka:29093,PLAINTEXT_HOST://:9092",
		"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
		"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
		"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
		"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@kafka:29093",
		"KAFKA_PROCESS_ROLES":                            "broker,controller",
		"KAFKA_LOG_DIRS":                                 "/tmp/kraft-combined-logs",
		"CLUSTER_ID":                                     "MkU3OEVBNTcwNTJENDM2Qk",
	}
)

const (
	Image       = "confluentinc/cp-kafka:7.7.1"
	exposedPort = "9092/tcp"

	ContainerName = "kafka"
	Port          = "9092"

	Broker = "localhost:9092"
	Topic  = "loms.order-events"
	Group  = "test-group"
)

func NewContainer(ctx context.Context, opt KafkaCntOpt) (tc.Container, error) {
	req := tc.ContainerRequest{
		Name:         opt.ContainerName,
		Image:        Image,
		ExposedPorts: []string{exposedPort},
		Env:          opt.Env,
		WaitingFor:   wait.ForListeningPort(nat.Port(exposedPort)).WithStartupTimeout(15 * time.Second),
		Networks:     []string{opt.Network.Name},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				nat.Port(exposedPort): []nat.PortBinding{{
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
