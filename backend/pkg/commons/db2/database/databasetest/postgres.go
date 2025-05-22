package databasetest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var oncePostgres sync.Once

func callerPackage() string {
	_, filename, _, _ := runtime.Caller(4)
	return filepath.Base(filepath.Dir(filename))
}

func NewPostgres(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()

	var err error
	var container testcontainers.Container
	// increase the container life, this way it can be reused
	_ = os.Setenv("RYUK_RECONNECTION_TIMEOUT", "30s")

	skipIfNoDocker(t, func() {
		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				// we run one postgres per package
				// packages are tested independently in go
				// we cannot easily retrieve a container used by another package
				Name:         "postgres_" + callerPackage(),
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
			Reuse:   true,
		})
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	url, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	postgresURL := fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", url)

	db, err := sqlx.Open("postgres", postgresURL)
	if err != nil {
		t.Fatal(err)
	}

	// ping database to check if it's ready
	if err := checkIfPostgresIsReady(db, 10, 1*time.Second); err != nil {
		t.Fatal(err)
	}

	// only run migration once
	oncePostgres.Do(func() {
		_, path, _, _ := runtime.Caller(0)
		// for now migration path is not configurable
		migrationPath := strings.ReplaceAll(filepath.Dir(path), "db2/database/databasetest", "db/migrations/postgres")
		if err := runMigrations(db, migrationPath); err != nil {
			if !strings.Contains(err.Error(), "no next version found") {
				t.Fatal(err)
			}
		}
	})
	t.Cleanup(func() {
		if err := truncate(db); err != nil {
			t.Fatal(err)
		}
	})
	return db
}

func runMigrations(db *sqlx.DB, path string) error {
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto CASCADE;"); err != nil {
		return err
	}

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db.DB, path)
}

func truncate(db *sqlx.DB) error {
	var tables []string
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';`
	err := db.Select(&tables, query)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", strings.Join(tables, ", ")))
	if err != nil {
		return err
	}
	return nil
}

func checkIfPostgresIsReady(db *sqlx.DB, retry int, delay time.Duration) error {
	var err error
	for i := 0; i < retry; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		fmt.Printf("waiting for postgres to be ready... attempt %d/%d\n", i+1, retry)
		time.Sleep(delay)
	}

	return fmt.Errorf("postgres is not ready after %d attempts: %v", retry, err)
}
