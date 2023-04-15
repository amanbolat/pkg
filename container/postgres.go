package pkgcontainer

import (
	"context"
	"fmt"
	"strings"

	pkgpostgres "github.com/amanbolat/pkg/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/ksuid"
)

// PostgresContainer is Postgres container.
type PostgresContainer struct {
	dsn           string
	containerName string
}

// PostgresContainerConfig is PostgresContainer configuration.
type PostgresContainerConfig struct {
	User      string
	Password  string
	DbName    string
	imageRepo string
	imageTag  string
}

// DefaultPostgresContainerConfig returns a default PostgresContainerConfig.
func DefaultPostgresContainerConfig() *PostgresContainerConfig {
	return &PostgresContainerConfig{
		User:      "user",
		Password:  "pass",
		DbName:    "postgres",
		imageRepo: "postgres",
		imageTag:  "15.1-alpine",
	}
}

// NewPostgresContainer creates a new PostgresContainer.
func NewPostgresContainer(ctx context.Context, cp *ContainerPool) (*PostgresContainer, error) {
	return NewPostgresContainerWithCfg(ctx, cp, DefaultPostgresContainerConfig())
}

// NewPostgresContainerWithCfg creates a new PostgresContainer with config.
func NewPostgresContainerWithCfg(ctx context.Context, cp *ContainerPool, cfg *PostgresContainerConfig) (*PostgresContainer, error) {
	container, err := cp.pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerResourcePrefix + ksuid.New().String(),
		Repository: cfg.imageRepo,
		Tag:        cfg.imageTag,
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

	dsn := fmt.Sprintf("postgresql://%s:%s@127.0.0.1:%s/%s?sslmode=disable", cfg.User, cfg.Password, container.GetPort("5432/tcp"), cfg.DbName)

	pgConn, err := pkgpostgres.NewConn(ctx, pkgpostgres.Config{
		DSN: dsn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connecto to postgres: %w", err)
	}

	// Close the connection as we use it only
	// to ensure that Postgres instance is ready.
	pgConn.Close()

	cp.addResource(container)

	name := strings.TrimPrefix(container.Container.Name, "/")

	ctr := &PostgresContainer{
		dsn:           dsn,
		containerName: name,
	}

	return ctr, nil
}

// DSN returns a Postgres connection string.
func (c *PostgresContainer) DSN() string {
	return c.dsn
}
