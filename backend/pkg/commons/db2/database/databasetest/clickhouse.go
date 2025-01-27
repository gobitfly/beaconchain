package databasetest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewClickHouse(t *testing.T) (clickhouse.Conn, *sql.DB) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "clickhouse/clickhouse-server:latest",
			Name:         "clickhouse-test-container",
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForListeningPort("9000"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start ClickHouse container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get ClickHouse container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get ClickHouse container port: %v", err)
	}

	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/default", "default", "", host, port.Port())
	options, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}

	client, err := clickhouse.Open(options)
	if err != nil {
		t.Fatalf("failed to get open ClickHouse connection %v", err)
	}

	db := clickhouse.OpenDB(options)
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping ClickHouse: %v", err)
	}

	return client, db
}
