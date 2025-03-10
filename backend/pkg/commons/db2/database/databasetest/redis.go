package databasetest

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewRedis(t testing.TB) *redis.Client {
	t.Helper()

	ctx := context.Background()

	var container testcontainers.Container
	var err error
	skipIfNoDocker(t, func() {
		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:latest",
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForLog("Ready to accept connections"),
			},
			Started: true,
		})
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	testcontainers.CleanupContainer(t, container)

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	return redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
}
