package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"runtime"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"golang.org/x/sync/errgroup"
)

var ClickHouseNativeWriter ch.Conn

func MustInitClickhouseNative(writer *types.DatabaseConfig) ch.Conn {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}
	log.Infof("initializing clickhouse native writer db connection to %v:%v/%v with %v/%v conn limit", writer.Host, writer.Port, writer.Name, writer.MaxIdleConns, writer.MaxOpenConns)
	dbWriter, err := ch.Open(&ch.Options{
		MaxOpenConns: writer.MaxOpenConns,
		MaxIdleConns: writer.MaxIdleConns,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr: []string{fmt.Sprintf("%s:%s", writer.Host, writer.Port)},
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
	})
	if err != nil {
		log.Fatal(err, "Error connecting to clickhouse native writer", 0)
	}
	// verify connection
	ClickHouseTestConnection(&dbWriter, writer.Name)

	return dbWriter
}

func ClickHouseTestConnection(db *ch.Conn, dataBaseName string) {
	v, err := (*db).ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping clickhouse database %s: %w", dataBaseName, err), "", 0)
	}
	log.Debugf("connected to clickhouse database %s with version %s", dataBaseName, v)
}

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
