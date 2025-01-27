package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var timeoutClickHouse = 3 * time.Second

type ClickHouse struct {
	db driver.Conn
}

func NewClickHouseWithClient(client driver.Conn, schema string) (*ClickHouse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutClickHouse)
	defer cancel()

	if err := client.Exec(ctx, schema); err != nil {
		return nil, err
	}
	return &ClickHouse{
		db: client,
	}, nil
}

func (client *ClickHouse) Add(table string, rows []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutClickHouse)
	defer cancel()

	batch, err := client.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", table))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}
	// TODO check if this is necessary
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

func (client *ClickHouse) clearTable(table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := client.db.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DELETE WHERE 1=1", table)); err != nil {
		return fmt.Errorf("failed to delete table entries: %w", err)
	}
	return nil
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
