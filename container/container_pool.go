package pkgcontainer

import (
	"fmt"
	"log"
	"sync"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/ksuid"
)

const containerResourcePrefix = "container-test_"

// ContainerPool is pool of container resources.
// It is used to create and manage a set of
// container resources.
type ContainerPool struct {
	pool      *dockertest.Pool
	resources []*dockertest.Resource
	networkID string
	mut       sync.RWMutex
	logDriver string
}

// ContainerPoolConfig is a ContainerPool configuration.
type ContainerPoolConfig struct {
	LogDriver string
}

// DefaultContainerPoolConfig returns a default ContainerPoolConfig.
func DefaultContainerPoolConfig() *ContainerPoolConfig {
	return &ContainerPoolConfig{
		LogDriver: "",
	}
}

// NewContainerPoolWithCfg creates a new ContainerPool with configuration.
func NewContainerPoolWithCfg(cfg *ContainerPoolConfig) (*ContainerPool, error) {
	dcPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create container pool: %w", err)
	}

	n, err := dcPool.Client.CreateNetwork(docker.CreateNetworkOptions{
		Name: containerResourcePrefix + ksuid.New().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container pool network: %w", err)
	}

	cp := &ContainerPool{
		pool:      dcPool,
		networkID: n.ID,
		logDriver: cfg.LogDriver,
	}

	return cp, nil
}

// NewContainerPool creates a new ContainerPool.
func NewContainerPool() (*ContainerPool, error) {
	return NewContainerPoolWithCfg(DefaultContainerPoolConfig())
}

func (dp *ContainerPool) addResource(r *dockertest.Resource) {
	dp.mut.Lock()
	defer dp.mut.Unlock()
	dp.resources = append(dp.resources, r)
}

// Stop tries to stop and remove all the ContainerPool resources.
func (dp *ContainerPool) Stop() {
	for _, r := range dp.resources {
		err := dp.pool.Purge(r)
		if err != nil {
			log.Printf("failed to purge container pool resource: %v", err)
		}
	}
	err := dp.pool.Client.RemoveNetwork(dp.networkID)
	if err != nil {
		log.Printf("failed to remove container pool network: %v", err)
	}
}
