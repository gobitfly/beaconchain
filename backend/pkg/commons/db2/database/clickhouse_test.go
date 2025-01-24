package database

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type chContainer struct {
	t         *testing.T
	ctx       context.Context
	container testcontainers.Container
	config    types.DatabaseConfig
}

func (c *chContainer) SetupClickHouseContainer() {
	var err error
	c.ctx = context.Background()

	c.container, err = testcontainers.GenericContainer(c.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "clickhouse/clickhouse-server:latest",
			Name:         "clickhouse-test-container",
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForListeningPort(nat.Port("9000")),
		},
		Started: true,
	})
	if err != nil {
		c.t.Fatalf("failed to start ClickHouse container: %v", err)
	}

	host, err := c.container.Host(c.ctx)
	if err != nil {
		c.t.Fatalf("failed to get ClickHouse container host: %v", err)
	}

	port, err := c.container.MappedPort(c.ctx, "9000")
	if err != nil {
		c.t.Fatalf("failed to get ClickHouse container port: %v", err)
	}

	c.config = types.DatabaseConfig{
		Username: "default",
		Password: "",
		Name:     "default",
		Host:     host,
		Port:     strconv.Itoa(port.Int()),
		SSL:      false,
	}

}

func (c *chContainer) TerminateClickHouseContainer() {
	if c.container != nil {
		err := c.container.Terminate(c.ctx)
		if err != nil {
			c.t.Fatalf("failed to stop ClickHouse container: %v", err)
		}
	}
}

func (c *chContainer) createTestTables(testTable string, client *ClickHouseClient) {
	err := client.NativeWriter.Exec(c.ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			chain_id LowCardinality(String),
			tx_index UInt64,
			block_number UInt64,
			tx_hash FixedString(32),
			from_address FixedString(20),
			to_address FixedString(20),
			timestamp DateTime,
			inserted_at DateTime MATERIALIZED now()
		) ENGINE = ReplacingMergeTree(inserted_at)
		PARTITION BY (toStartOfQuarter(timestamp), chain_id)
		ORDER BY (timestamp, block_number, tx_index)
		SETTINGS index_granularity = 8192
		`, testTable))

	if err != nil {
		c.t.Fatalf("failed to create test table: %v", err)
	}

	c.t.Logf("CREATED TABLE %s", testTable)

}

func TestClickHouseClient_Add(t *testing.T) {
	container := &chContainer{t: t}
	container.SetupClickHouseContainer()
	defer container.TerminateClickHouseContainer()

	client, err := NewClickHouseClient(&container.config, &container.config)
	if err != nil {
		t.Fatalf("failed to create ClickHouse client: %v", err)
	}

	testTable := "transactions_test"
	container.createTestTables(testTable, client)
	// err = client.CheckIfTablesExist()
	// if err != nil {
	// 	t.Fatalf("failed to check if tables exist: %v", err)
	// }

	convertString := func(input string) string {
		bytes, err := convertToBytes(input)
		if err != nil {
			t.Fatalf("failed to convert to bytes: %v", err)
		}
		return string(bytes)
	}

	txHash := []string{
		convertString("0x5f3b03d742a967acbdcf013fa3fd63669f4b89a5f0e77dbd6db3877b07c16bb0"),
		convertString("0x869e4720c6f00378d97b10e2fdb3a493c8a561907533d4e5cf41fd74bb560268"),
		convertString("0xabcde4720c6f00378d97b10e2fdb3a493c8a561907533d4e5cf41fd74bb56026"),
	}

	fromAddress := []string{
		convertString("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"),
		convertString("0x0e5dda855eb1de2a212cd1f62b2a3ee49d20c444"),
	}

	toAddress := []string{
		convertString("0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"),
		convertString("0x5c0ab2d9b5a7ed9f470386e82bb36a3613cdd4b5"),
	}

	t.Run("AddSingleRow", func(t *testing.T) {
		row := []Transaction{
			{
				ChainID:     "1",
				TxIndex:     1,
				BlockNumber: 1,
				TxHash:      txHash[0],
				FromAddress: fromAddress[0],
				ToAddress:   toAddress[0],
				Timestamp:   time.Now(),
			},
		}

		if err := client.Add(testTable, row); err != nil {
			t.Fatalf("AddSingleRow failed: %v", err)
		}
	})

	t.Run("AddSingleRowWithEmptyToAddress", func(t *testing.T) {
		row := []Transaction{
			{
				ChainID:     "1",
				TxIndex:     1,
				BlockNumber: 1,
				TxHash:      txHash[0],
				FromAddress: fromAddress[0],
				Timestamp:   time.Now(),
			},
		}

		if err := client.Add(testTable, row); err != nil {
			t.Fatalf("AddSingleRowWithEmptyToAddress failed: %v", err)
		}
	})

	t.Run("AddBatchRows", func(t *testing.T) {
		rows := []Transaction{
			{
				ChainID:     "1",
				TxIndex:     2,
				BlockNumber: 2,
				TxHash:      txHash[1],
				FromAddress: fromAddress[0],
				ToAddress:   toAddress[0],
				Timestamp:   time.Now(),
			},
			{
				ChainID:     "1",
				TxIndex:     3,
				BlockNumber: 3,
				TxHash:      txHash[2],
				FromAddress: fromAddress[1],
				ToAddress:   toAddress[1],
				Timestamp:   time.Now(),
			},
		}

		if err := client.Add(testTable, rows); err != nil {
			t.Fatalf("AddBatchRows failed: %v", err)
		}
	})

	t.Run("AddInvalidRowWithEmptyFromAddress", func(t *testing.T) {
		row := []Transaction{
			{
				ChainID:     "1",
				TxIndex:     1,
				BlockNumber: 1,
				TxHash:      txHash[1],
				ToAddress:   toAddress[1],
				Timestamp:   time.Now(),
			},
		}

		err := client.Add(testTable, row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidRowWithEmptyTxHash", func(t *testing.T) {
		row := []Transaction{
			{
				ChainID:     "1",
				TxIndex:     1,
				BlockNumber: 1,
				FromAddress: fromAddress[1],
				ToAddress:   toAddress[1],
				Timestamp:   time.Now(),
			},
		}

		err := client.Add(testTable, row)
		if err == nil {
			t.Fatalf("expected error for missing tx_hash field, but got none")
		}
	})

}

func convertToBytes(s string) ([]byte, error) {
	bytes, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
