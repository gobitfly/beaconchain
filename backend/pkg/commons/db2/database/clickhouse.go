package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
)

var timeoutClickHouse = 3 * time.Second

type ClickHouse struct {
	db driver.Conn
}

func NewClickHouseWithClient(cfg *types.DatabaseConfig) (*ClickHouse, error) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.MaxOpenConns < cfg.MaxIdleConns {
		cfg.MaxIdleConns = cfg.MaxOpenConns
	}

	var hosts []string
	hosts = append(hosts, net.JoinHostPort(cfg.Host, cfg.Port))
	for _, f := range cfg.Failovers {
		hosts = append(hosts, net.JoinHostPort(f.Host, f.Port))
	}

	var tlsConfig *tls.Config
	if cfg.SSL {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}
	}

	log.Infof("initializing ClickHouse db connection to %v/%v with %v/%v conn limit", hosts, cfg.Name, cfg.MaxIdleConns, cfg.MaxOpenConns)
	client, err := ch.Open(&ch.Options{
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr:             hosts,
		ConnOpenStrategy: ch.ConnOpenInOrder,
		Auth: ch.Auth{
			Username: cfg.Username,
			Password: cfg.Password,
			Database: cfg.Name,
		},
		Debug: false,
		TLS:   tlsConfig,
		// this gets only called when debug is true
		Debugf: func(s string, p ...interface{}) {
			log.Debugf(s, p...)
		},
		Settings: ch.Settings{
			"deduplicate_blocks_in_dependent_materialized_views":                "1",
			"update_insert_deduplication_token_in_dependent_materialized_views": "1",
		},
		ClientInfo: ch.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "beaconchain-explorer", Version: version.Version},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to ClickHouse db: %v", err)
	}

	version, err := client.ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping ClickHouse database %s: %w", cfg.Name, err), "", 0)
	}
	log.Debugf("connected to ClickHouse database %s with version %s", cfg.Name, version)

	return &ClickHouse{
		db: client,
	}, nil
}

func InitSchema(client driver.Conn, schema string) (*ClickHouse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutClickHouse)
	defer cancel()

	if err := client.Exec(ctx, schema); err != nil {
		return nil, err
	}
	return &ClickHouse{
		db: client,
	}, nil
}

func (client *ClickHouse) Close() error {
	if client.db != nil {
		if err := client.db.Close(); err != nil {
			return fmt.Errorf("failed to close ClickHouse connection: %v", err)
		}
	}
	return nil
}

func (client *ClickHouse) Add(table string, rows []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutClickHouse)
	defer cancel()

	batch, err := client.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", table))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	defer func() {
		if !batch.IsSent() {
			_ = batch.Abort()
		}
	}()

	for _, row := range rows {
		if err := batch.AppendStruct(row); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	return nil
}

func (client *ClickHouse) Read(query string, scan func(driver.Rows) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	rows, err := client.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
	}
	return rows.Err()
}

func ScanArray[S ~[]E, E any](s *S) func(rows driver.Rows) error {
	return func(rows driver.Rows) error {
		var temp E
		if err := rows.ScanStruct(&temp); err != nil {
			return err
		}
		*s = append(*s, temp)
		return nil
	}
}
