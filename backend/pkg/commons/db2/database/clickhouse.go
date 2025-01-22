package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"runtime"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"golang.org/x/sync/errgroup"
)

type ClickHouseClient struct {
	readerCfg *types.DatabaseConfig
	writerCfg *types.DatabaseConfig

	NativeWriter ch.Conn
	NativeReader ch.Conn
}

type ClickHouseDB interface {
	Add(table string, rows []map[string]interface{}) error
	Read(query string, args ...interface{}) ([]map[string]interface{}, error)

	Close() error
}

func NewClickHouseClient(writer, reader *types.DatabaseConfig) (*ClickHouseClient, error) {
	client := &ClickHouseClient{
		readerCfg: reader,
		writerCfg: writer,
	}

	if err := client.initClickHouseWriter(writer); err != nil {
		return nil, fmt.Errorf("failed to initialize ClickHouse writer: %v", err)
	}

	if err := client.initClickHouseReader(reader); err != nil {
		return nil, fmt.Errorf("failed to initialize ClickHouse reader: %v", err)
	}

	if err := client.checkIfTablesExist(); err != nil {
		return nil, fmt.Errorf("failed to check if tables exist in ClickHouse: %v", err)
	}

	return client, nil
}

func (client *ClickHouseClient) initClickHouseWriter(writer *types.DatabaseConfig) error {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}

	var writerHosts []string
	writerHosts = append(writerHosts, net.JoinHostPort(writer.Host, writer.Port))
	for _, f := range writer.Failovers {
		writerHosts = append(writerHosts, net.JoinHostPort(f.Host, f.Port))
	}

	log.Infof("initializing ClickHouse native writer db connection to %v/%v with %v/%v conn limit", writerHosts, writer.Name, writer.MaxIdleConns, writer.MaxOpenConns)
	dbWriter, err := ch.Open(&ch.Options{
		MaxOpenConns: writer.MaxOpenConns,
		MaxIdleConns: writer.MaxIdleConns,
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr:             writerHosts,
		ConnOpenStrategy: ch.ConnOpenInOrder,
		Auth: ch.Auth{
			Username: writer.Username,
			Password: writer.Password,
			Database: writer.Name,
		},
		Debug: false,
		TLS:   &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
		// this gets only called when debug is true
		Debugf: func(s string, p ...interface{}) {
			log.Debugf("CH NATIVE WRITER: "+s, p...)
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
		return fmt.Errorf("error connecting to ClickHouse native writer: %v", err)
	}
	client.NativeWriter = dbWriter
	return nil
}

func (client *ClickHouseClient) initClickHouseReader(reader *types.DatabaseConfig) error {
	if reader.MaxOpenConns == 0 {
		reader.MaxOpenConns = 50
	}
	if reader.MaxIdleConns == 0 {
		reader.MaxIdleConns = 10
	}
	if reader.MaxOpenConns < reader.MaxIdleConns {
		reader.MaxIdleConns = reader.MaxOpenConns
	}

	var readerHosts []string
	readerHosts = append(readerHosts, net.JoinHostPort(reader.Host, reader.Port))
	for _, f := range reader.Failovers {
		readerHosts = append(readerHosts, net.JoinHostPort(f.Host, f.Port))
	}

	log.Infof("initializing ClickHouse native reader db connection to %v/%v with %v/%v conn limit", readerHosts, reader.Name, reader.MaxIdleConns, reader.MaxOpenConns)
	dbReader, err := ch.Open(&ch.Options{
		MaxOpenConns: reader.MaxOpenConns,
		MaxIdleConns: reader.MaxIdleConns,
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr:             readerHosts,
		ConnOpenStrategy: ch.ConnOpenInOrder,
		Auth: ch.Auth{
			Username: reader.Username,
			Password: reader.Password,
			Database: reader.Name,
		},
		Debug: false,
		TLS:   &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
		// this gets only called when debug is true
		Debugf: func(s string, p ...interface{}) {
			log.Debugf("CH NATIVE READER: "+s, p...)
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
		return fmt.Errorf("error connecting to ClickHouse native reader: %v", err)
	}

	client.NativeReader = dbReader
	return nil
}

func (client *ClickHouseClient) checkIfTablesExist() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var count int
	err := client.NativeReader.QueryRow(ctx, `
		SELECT count(*) 
		FROM system.tables 
		WHERE database = ?`,
		client.writerCfg.Name,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if tables exist in ClickHouse database: %w", err)
	}

	if count == 0 {
		log.Warnf("Schema doesn't exist in ClickHouse database %s. Applying migration...", client.writerCfg.Name)
		err := initTables(client.writerCfg, "migrations/clickhouse")
		if err != nil {
			return fmt.Errorf("failed to apply migration on ClickHouse database %s: %w", client.writerCfg.Name, err)
		}
	}

	return nil
}

func initTables(writerCfg *types.DatabaseConfig, migrationPath string) error {

	sslParam := "secure=false"
	if writerCfg.SSL {
		sslParam = "secure=true"
	}

	// connect using sqlx for goose
	db, err := sqlx.Open("clickhouse", fmt.Sprintf(
		"clickhouse://%s:%s@%s/%s?%s",
		writerCfg.Username,
		writerCfg.Password,
		net.JoinHostPort(writerCfg.Host, writerCfg.Port),
		writerCfg.Name,
		sslParam,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("clickhouse"); err != nil {
		return err
	}

	log.Warnf("applying schema to %s database...", writerCfg.Name)
	if err := goose.Up(db.DB, migrationPath); err != nil {
		return err
	}

	return nil
}

func ClickHouseTestConnection(db ch.Conn, dataBaseName string) {
	v, err := db.ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping ClickHouse database %s: %w", dataBaseName, err), "", 0)
	}
	log.Debugf("connected to ClickHouse database %s with version %s", dataBaseName, v)
}

func (client *ClickHouseClient) Close() error {
	if client.NativeWriter != nil {
		if err := client.NativeWriter.Close(); err != nil {
			return fmt.Errorf("failed to close ClickHouse writer connection: %v", err)
		}
	}

	if client.NativeReader != nil {
		if err := client.NativeReader.Close(); err != nil {
			return fmt.Errorf("failed to close ClickHouse reader connection: %v", err)
		}
	}

	return nil
}

func (client *ClickHouseClient) Add(table string, rows []map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	batch, err := client.NativeWriter.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", table))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	defer func() {
		if batch.IsSent() {
			return
		}
		err := batch.Abort()
		if err != nil {
			log.Warnf("failed to abort batch: %v", err)
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

func (client *ClickHouseClient) Read(db *sqlx.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	rows, err := client.NativeReader.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.Scan(row); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, row)
	}

	return results, nil
}

var ClickHouseNativeWriter ch.Conn
var ClickHouseReader *sqlx.DB

type UltraFastClickhouseStruct interface {
	Get(string) any
	Extend(UltraFastClickhouseStruct) error
}

func UltraFastDumpToClickhouse[T UltraFastClickhouseStruct](data T, target_table string, insert_uuid string) error {
	start := time.Now()
	// add metrics
	defer func() {
		metrics.TaskDuration.WithLabelValues(fmt.Sprintf("clickhouse_dump_%s_overall", target_table)).Observe(time.Since(start).Seconds())
	}()
	now := time.Now()
	// get column order & names from clickhouse
	var columns []string
	err := ClickHouseReader.Select(&columns, "SELECT name FROM system.columns where table=$1 and database=currentDatabase() order by position;", target_table)
	if err != nil {
		return err
	}
	metrics.TaskDuration.WithLabelValues(fmt.Sprintf("clickhouse_dump_%s_get_columns", target_table)).Observe(time.Since(now).Seconds())
	now = time.Now()
	// prepare batch
	abortCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ctx := ch.Context(abortCtx, ch.WithSettings(ch.Settings{
		"insert_deduplication_token": insert_uuid, // 重复数据插入时，会根据这个字段进行去重
		"insert_deduplicate":         true,
	}), ch.WithLogs(func(l *ch.Log) {
		log.Debugf("CH NATIVE WRITER: %s", l.Text)
	}),
	)
	batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `INSERT INTO `+target_table)
	if err != nil {
		return err
	}
	metrics.TaskDuration.WithLabelValues(fmt.Sprintf("clickhouse_dump_%s_prepare_batch", target_table)).Observe(time.Since(now).Seconds())
	now = time.Now()
	defer func() {
		if batch.IsSent() {
			return
		}
		err := batch.Abort()
		if err != nil {
			log.Warnf("failed to abort batch: %v", err)
		}
	}()
	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())
	// iterate columns retrieved from clickhouse
	for i, n := range columns {
		// Capture the loop variable
		col_index := i
		col_name := n
		if col_name == "_inserted_at" {
			continue
		}
		// Start a new goroutine for each column
		g.Go(func() error {
			// get it from the struct
			column := data.Get(col_name)
			if column == nil {
				return fmt.Errorf("column %s not found in struct", col_name)
			}
			// Perform the type assertion and append operation
			err = batch.Column(col_index).Append(column)
			log.Debugf("appended column %s in %s", col_name, time.Since(now))
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	metrics.TaskDuration.WithLabelValues(fmt.Sprintf("clickhouse_dump_%s_append_columns", target_table)).Observe(time.Since(now).Seconds())
	now = time.Now()
	err = batch.Send()
	if err != nil {
		return err
	}
	metrics.TaskDuration.WithLabelValues(fmt.Sprintf("clickhouse_dump_%s_send_batch", target_table)).Observe(time.Since(now).Seconds())
	return nil
}
