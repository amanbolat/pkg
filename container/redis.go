package pkgcontainer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/ksuid"
)

// RedisContainer is Redis container.
type RedisContainer struct {
	dsn           string
	containerName string
}

// RedisContainerConfig is RedisContainer configuration.
type RedisContainerConfig struct {
	imageRepo string
	imageTag  string
}

// DefaultRedisContainerConfig returns a default RedisContainerConfig.
func DefaultRedisContainerConfig() *RedisContainerConfig {
	return &RedisContainerConfig{
		imageRepo: "redis",
		imageTag:  "7.0.9-alpine3.17",
	}
}

// NewRedisContainer creates a new RedisContainer.
func NewRedisContainer(ctx context.Context, cp *ContainerPool) (*RedisContainer, error) {
	return NewRedisContainerWithCfg(ctx, cp, DefaultRedisContainerConfig())
}

// NewRedisContainerWithCfg creates a new RedisContainer with config.
func NewRedisContainerWithCfg(ctx context.Context, cp *ContainerPool, cfg *RedisContainerConfig) (*RedisContainer, error) {
	container, err := cp.pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerResourcePrefix + ksuid.New().String(),
		Repository: cfg.imageRepo,
		Tag:        cfg.imageTag,
		NetworkID:  cp.networkID,
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
		return nil, fmt.Errorf("failed to create redis container: %w", err)
	}

	dsn := fmt.Sprintf("redis://127.0.0.1:%s/0", container.GetPort("6379/tcp"))

	err = retry.Do(func() error {
		options, err := redis.ParseURL(dsn)
		if err != nil {
			return err
		}
		redisConn := redis.NewClient(options)

		ping := redisConn.Ping(ctx)
		err = ping.Err()
		if err != nil {
			return fmt.Errorf("failed to ping redis: %w", err)
		}

		return redisConn.Close()
	}, retry.Attempts(10), retry.Delay(time.Second), retry.DelayType(retry.FixedDelay), retry.Context(ctx))

	if err != nil {
		return nil, err
	}

	cp.addResource(container)

	name := strings.TrimPrefix(container.Container.Name, "/")

	ctr := &RedisContainer{
		dsn:           dsn,
		containerName: name,
	}

	return ctr, nil
}

// DSN returns a Redis connection string.
func (c *RedisContainer) DSN() string {
	return c.dsn
}
