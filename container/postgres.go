package pkgcontainer

import (
	"context"
	"fmt"

	pkgpostgres "github.com/amanbolat/pkg/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/ksuid"
)

// PostgresContainer is Postgres container.
type PostgresContainer struct {
	dsn           string
	containerName string
	container     *docker.Container
}

// PostgresContainerConfig is PostgresContainer configuration.
type PostgresContainerConfig struct {
	User     string
	Password string
	DbName   string
	ImageTag string
}

// DefaultPostgresContainerConfig returns a default PostgresContainerConfig.
func DefaultPostgresContainerConfig() *PostgresContainerConfig {
	return &PostgresContainerConfig{
		User:     "user",
		Password: "pass",
		DbName:   "postgres",
		ImageTag: "15.1-alpine",
	}
}

// NewPostgresContainer creates a new PostgresContainer.
func NewPostgresContainer(ctx context.Context, cp *ContainerPool) (*PostgresContainer, error) {
	return NewPostgresContainerWithCfg(ctx, cp, DefaultPostgresContainerConfig())
}

// NewPostgresContainerWithCfg creates a new PostgresContainer with config.
func NewPostgresContainerWithCfg(ctx context.Context, cp *ContainerPool, cfg *PostgresContainerConfig) (*PostgresContainer, error) {
	res, err := cp.pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerResourcePrefix + ksuid.New().String(),
		Repository: "postgres",
		Tag:        cfg.ImageTag,
		NetworkID:  cp.networkID,
		Env: []string{
			"POSTGRES_USER=" + cfg.User,
			"POSTGRES_PASSWORD=" + cfg.Password,
			"POSTGRES_DB=" + cfg.DbName,
		},
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
		return nil, fmt.Errorf("failed to create postgres container: %w", err)
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@127.0.0.1:%s/%s?sslmode=disable", cfg.User, cfg.Password, res.GetPort("5432/tcp"), cfg.DbName)

	pgConn, err := pkgpostgres.NewConn(ctx, pkgpostgres.Config{
		DSN: dsn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connecto to postgres: %w", err)
	}

	// Close the connection as we use it only
	// to ensure that Postgres instance is ready.
	pgConn.Close()

	cp.addResource(res)

	ctr := &PostgresContainer{
		dsn:       dsn,
		container: res.Container,
	}

	return ctr, nil
}

// DSN returns a Postgres connection string.
func (c *PostgresContainer) DSN() string {
	return c.dsn
}
