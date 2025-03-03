package databasetest

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewPostgres(t *testing.T) *sqlx.DB {
	t.Helper()

	ctx := context.Background()

	var err error
	var container testcontainers.Container
	skipIfNoDocker(t, func() {
		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:12",
				ExposedPorts: []string{"5432/tcp"},
				Env: map[string]string{
					"POSTGRES_USER":     "postgres",
					"POSTGRES_PASSWORD": "postgres",
					"POSTGRES_DB":       "postgres",
				},
				WaitingFor: wait.ForListeningPort("5432/tcp"),
			},
			Started: true,
		})
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	testcontainers.CleanupContainer(t, container)

	url, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", url))
	if err != nil {
		t.Fatal(err)
	}

	// run migrations
	// for now migration path is not configurable
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto CASCADE;"); err != nil {
		t.Fatal(err)
	}
	_, path, _, _ := runtime.Caller(0)
	migrationPath := strings.ReplaceAll(filepath.Dir(path), "db2/database/databasetest", "db/migrations/postgres")

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db.DB, migrationPath); err != nil {
		t.Fatal(err)
	}

	return db
}
