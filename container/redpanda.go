package pkgcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/amanbolat/pkg/net"
	"github.com/avast/retry-go"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/ksuid"
)

// RedpandaContainer is redpanda container, which has
// a fully Kafka compatible API.
type RedpandaContainer struct {
	brokers []string
}

// Brokers returns redpanda broker addresses.
func (c *RedpandaContainer) Brokers() []string {
	arr := make([]string, len(c.brokers))
	copy(arr, c.brokers)

	return arr
}

// RedpandaContainerConfig is used to configure
// redpanda container.
type RedpandaContainerConfig struct {
	imageRepo string
	imageTag  string
}

// DefaultRedpandaContainerConfig returns a default redpanda configuration.
func DefaultRedpandaContainerConfig() *RedpandaContainerConfig {
	return &RedpandaContainerConfig{
		imageRepo: "vectorized/redpanda",
		imageTag:  "v22.3.10",
	}
}

// NewRedpandaContainer returns a new RedpandaContainer.
func NewRedpandaContainer(ctx context.Context, cp *ContainerPool) (*RedpandaContainer, error) {
	return NewRedpandaContainerWithConfig(ctx, cp, DefaultRedpandaContainerConfig())
}

// NewRedpandaContainerWithConfig configures and returns a new RedpandaContainer.
func NewRedpandaContainerWithConfig(ctx context.Context, cp *ContainerPool, cfg *RedpandaContainerConfig) (*RedpandaContainer, error) {
	kafkaPort, err := pkgnet.RandomTCPPort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free tcp port: %w", err)
	}

	containerName := containerResourcePrefix + ksuid.New().String()
	kafkaTCPPort := fmt.Sprintf("%d/tcp", kafkaPort)
	container, err := cp.pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: cfg.imageRepo,
		Tag:        cfg.imageTag,
		NetworkID:  cp.networkID,
		PortBindings: map[docker.Port][]docker.PortBinding{
			docker.Port(kafkaTCPPort): {{HostIP: "localhost", HostPort: kafkaTCPPort}},
		},
		Cmd: []string{
			"redpanda",
			"start",
			"--smp",
			"1",
			"--reserve-memory",
			"0M",
			"--overprovisioned",
			"--set", "redpanda.empty_seed_starts_cluster=false",
			"--seeds", fmt.Sprintf("%s:33145", containerName),
			"--kafka-addr",
			fmt.Sprintf("PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:%d", kafkaPort),
			"--advertise-kafka-addr",
			fmt.Sprintf("PLAINTEXT://redpanda:29092,OUTSIDE://localhost:%d", kafkaPort),
			"--pandaproxy-addr",
			"PLAINTEXT://0.0.0.0:28082,OUTSIDE://0.0.0.0:8082",
			"--advertise-pandaproxy-addr",
			"PLAINTEXT://redpanda:28082,OUTSIDE://localhost:8082",
			"--advertise-rpc-addr", fmt.Sprintf("%s:33145", containerName),
		},
		ExposedPorts: []string{kafkaTCPPort},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		if cp.logDriver != "" {
			config.LogConfig = docker.LogConfig{
				Type: cp.logDriver,
			}
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create redpanda container: %w", err)
	}

	brokerAddr := fmt.Sprintf("localhost:%s", container.GetPort(kafkaTCPPort))

	err = retry.Do(func() error {
		conn, connErr := kafka.DialContext(ctx, "tcp", brokerAddr)
		if connErr != nil {
			return fmt.Errorf("failed to dial redpanda container: %w", connErr)
		}

		_, err = conn.Brokers()
		if err != nil {
			return fmt.Errorf("failed to get redpanda api versions: %w", err)
		}
		defer conn.Close()

		return nil
	}, retry.Attempts(10), retry.Delay(time.Second), retry.DelayType(retry.FixedDelay), retry.Context(ctx))

	if err != nil {
		return nil, err
	}

	cp.addResource(container)
	kafkaCtr := &RedpandaContainer{brokers: []string{brokerAddr}}

	return kafkaCtr, nil
}
