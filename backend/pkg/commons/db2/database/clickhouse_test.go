package database

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
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

func createTestTables(client *ClickHouseClient, tableSchemas map[string]string, ctx context.Context) error {
	for tableName, schema := range tableSchemas {
		query := fmt.Sprintf(schema, tableName)
		if err := client.NativeWriter.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create table %s: %v", tableName, err)
		}
	}
	return nil
}

func setupTablesForTests(client *ClickHouseClient, ctx context.Context) error {
	tableSchemas := map[string]string{
		"transactions_test": `
			CREATE TABLE IF NOT EXISTS %s (
				chain_id LowCardinality(String),
				tx_index UInt64,
				block_number UInt64,
				tx_hash FixedString(32),
				from_address FixedString(20),
				to_address FixedString(20),
				timestamp DateTime
			) ENGINE = Memory;
		`,
		"internal_transactions_test": `
			CREATE TABLE IF NOT EXISTS %s (
				chain_id LowCardinality(String),
				parent_hash FixedString(32),
				block_number UInt64,
				from_address FixedString(20),
				to_address FixedString(20),
				internal_index UInt64,
				timestamp DateTime
			) ENGINE = Memory;
		`,
		"erc20_transfers_test": `
			CREATE TABLE IF NOT EXISTS %s (
				chain_id LowCardinality(String),
				parent_hash FixedString(32),
				block_number UInt64,
				from_address FixedString(20),
				to_address FixedString(20),
				token_address FixedString(20),
				log_index UInt32,
				tx_log_index UInt32,
				timestamp DateTime
			) ENGINE = Memory;
		`,
		"erc721_transfers_test": `
			CREATE TABLE IF NOT EXISTS %s (
				chain_id LowCardinality(String),
				parent_hash FixedString(32),
				block_number UInt64,
				from_address FixedString(20),
				to_address FixedString(20),
				token_address FixedString(20),
				token_id UInt256,
				log_index UInt32,
				tx_log_index UInt32,
				timestamp DateTime
			) ENGINE = Memory;
		`,
		"erc1155_transfers_test": `
			CREATE TABLE IF NOT EXISTS %s (
				chain_id LowCardinality(String),
				parent_hash FixedString(32),
				block_number UInt64,
				from_address FixedString(20),
				to_address FixedString(20),
				operator FixedString(20),
				token_address FixedString(20),
				token_id UInt256,
				log_index UInt32,
				tx_log_index UInt32,
				timestamp DateTime
			) ENGINE = Memory;
		`,
	}
	return createTestTables(client, tableSchemas, ctx)
}

func TestClickHouseClient_Add(t *testing.T) {
	container := &chContainer{t: t}
	container.SetupClickHouseContainer()
	defer container.TerminateClickHouseContainer()

	client, err := NewClickHouseClient(&container.config, &container.config)
	if err != nil {
		t.Fatalf("failed to create ClickHouse client: %v", err)
	}

	err = setupTablesForTests(client, container.ctx)
	if err != nil {
		t.Fatalf("failed to setup tables for tests: %v", err)
	}
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

	operatorAddress := []string{
		convertString("0x5f70D50ab5a2E8C658b75F4D4ea21a5e13fC5396"),
		convertString("0x2c64A40eC4E11C4b8fBcD87275b7aF3dB4C76a22"),
	}

	t.Run("AddSingleTxRow", func(t *testing.T) {
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

		if err := client.Add("transactions_test", row); err != nil {
			t.Fatalf("AddSingleTxRow failed: %v", err)
		}
	})

	t.Run("AddSingleTxRowWithEmptyToAddress", func(t *testing.T) {
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

		if err := client.Add("transactions_test", row); err != nil {
			t.Fatalf("AddSingleTxRowWithEmptyToAddress failed: %v", err)
		}
	})

	t.Run("AddBatchTxRows", func(t *testing.T) {
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

		if err := client.Add("transactions_test", rows); err != nil {
			t.Fatalf("AddBatchTxRows failed: %v", err)
		}
	})

	t.Run("AddInvalidTxRowWithEmptyChainID", func(t *testing.T) {
		row := []Transaction{
			{
				TxIndex:     1,
				BlockNumber: 1,
				TxHash:      txHash[0],
				FromAddress: fromAddress[0],
				ToAddress:   toAddress[0],
				Timestamp:   time.Now(),
			},
		}

		err := client.Add("transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing chain_id field, but got none")
		}
	})

	t.Run("AddInvalidTxRowWithEmptyFromAddress", func(t *testing.T) {
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

		err := client.Add("transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidTxRowWithEmptyTxHash", func(t *testing.T) {
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

		err := client.Add("transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing tx_hash field, but got none")
		}
	})

	t.Run("AddSingleITXRow", func(t *testing.T) {
		row := []InternalTx{
			{
				ChainID:       "1",
				ParentHash:    txHash[1],
				BlockNumber:   1,
				FromAddress:   fromAddress[1],
				ToAddress:     toAddress[1],
				InternalIndex: 0,
				Timestamp:     time.Now(),
			},
		}

		if err := client.Add("internal_transactions_test", row); err != nil {
			t.Fatalf("AddSingleITXRow failed: %v", err)
		}
	})

	t.Run("AddBatchITXRows", func(t *testing.T) {
		rows := []InternalTx{
			{
				ChainID:       "1",
				ParentHash:    txHash[1],
				BlockNumber:   1,
				FromAddress:   fromAddress[0],
				ToAddress:     toAddress[0],
				InternalIndex: 1,
				Timestamp:     time.Now(),
			},
			{
				ChainID:       "1",
				ParentHash:    txHash[1],
				BlockNumber:   1,
				FromAddress:   fromAddress[1],
				ToAddress:     toAddress[1],
				InternalIndex: 2,
				Timestamp:     time.Now(),
			},
		}

		if err := client.Add("internal_transactions_test", rows); err != nil {
			t.Fatalf("AddBatchITXRows failed: %v", err)
		}
	})

	t.Run("AddInvalidITXRowWithEmptyChainID", func(t *testing.T) {
		row := []InternalTx{
			{
				ParentHash:    txHash[1],
				BlockNumber:   1,
				FromAddress:   fromAddress[1],
				ToAddress:     toAddress[1],
				InternalIndex: 2,
				Timestamp:     time.Now(),
			},
		}

		err := client.Add("internal_transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing chain_id field, but got none")
		}
	})

	t.Run("AddInvalidITXRowWithEmptyFromAddress", func(t *testing.T) {
		row := []InternalTx{
			{
				ChainID:       "1",
				ParentHash:    txHash[1],
				BlockNumber:   1,
				ToAddress:     toAddress[1],
				InternalIndex: 0,
				Timestamp:     time.Now(),
			},
		}

		err := client.Add("internal_transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidTxRowWithEmptyParentHash", func(t *testing.T) {
		row := []InternalTx{
			{
				ChainID:       "1",
				BlockNumber:   1,
				FromAddress:   fromAddress[1],
				ToAddress:     toAddress[1],
				InternalIndex: 2,
				Timestamp:     time.Now(),
			},
		}

		err := client.Add("internal_transactions_test", row)
		if err == nil {
			t.Fatalf("expected error for missing parent_hash field, but got none")
		}
	})

	t.Run("AddSingleERC20Row", func(t *testing.T) {
		row := []ERC20{
			{
				ChainID:      "1",
				ParentHash:   txHash[0],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[0],
				TokenAddress: toAddress[0],
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc20_transfers_test", row); err != nil {
			t.Fatalf("AddSingleERC20Row failed: %v", err)
		}
	})

	t.Run("AddBatchERC20Rows", func(t *testing.T) {
		rows := []ERC20{
			{
				ChainID:      "1",
				ParentHash:   txHash[0],
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[0],
				TokenAddress: toAddress[1],
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc20_transfers_test", rows); err != nil {
			t.Fatalf("AddBatchERC20Rows failed: %v", err)
		}
	})

	t.Run("AddInvalidERC20RowWithEmptyChainID", func(t *testing.T) {
		row := []ERC20{
			{
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc20_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing chain_id field, but got none")
		}
	})

	t.Run("AddInvalidERC20RowWithEmptyFromAddress", func(t *testing.T) {
		row := []ERC20{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc20_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidERC20RowWithEmptyParentHash", func(t *testing.T) {
		row := []ERC20{
			{
				ChainID:      "1",
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc20_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing parent_hash field, but got none")
		}
	})

	t.Run("AddSingleERC721Row", func(t *testing.T) {
		row := []ERC721{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc721_transfers_test", row); err != nil {
			t.Fatalf("AddSingleERC721Row failed: %v", err)
		}
	})

	t.Run("AddBatchERC721Rows", func(t *testing.T) {
		rows := []ERC721{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc721_transfers_test", rows); err != nil {
			t.Fatalf("AddBatchERC721Rows failed: %v", err)
		}
	})

	t.Run("AddInvalidERC721RowWithEmptyChainID", func(t *testing.T) {
		row := []ERC721{
			{
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc721_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing chain_id field, but got none")
		}
	})

	t.Run("AddInvalidERC721RowWithEmptyFromAddress", func(t *testing.T) {
		row := []ERC721{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc721_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidERC721RowWithEmptyParentHash", func(t *testing.T) {
		row := []ERC721{
			{
				ChainID:      "1",
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[1],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc721_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing parent_hash field, but got none")
		}
	})

	t.Run("AddSingleERC1155Row", func(t *testing.T) {
		row := []ERC1155{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[1],
				Operator:     operatorAddress[0],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc1155_transfers_test", row); err != nil {
			t.Fatalf("AddSingleERC1155Row failed: %v", err)
		}
	})

	t.Run("AddBatchERC1155Rows", func(t *testing.T) {
		rows := []ERC1155{
			{
				ChainID:      "1",
				ParentHash:   txHash[0],
				BlockNumber:  1,
				FromAddress:  fromAddress[0],
				ToAddress:    toAddress[0],
				Operator:     operatorAddress[0],
				TokenAddress: toAddress[0],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   1,
				Timestamp:    time.Now(),
			},
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				Operator:     operatorAddress[1],
				TokenAddress: toAddress[1],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		if err := client.Add("erc1155_transfers_test", rows); err != nil {
			t.Fatalf("AddBatchERC1155Rows failed: %v", err)
		}
	})

	t.Run("AddInvalidERC1155RowWithEmptyChainID", func(t *testing.T) {
		row := []ERC1155{
			{
				ParentHash:   txHash[1],
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				Operator:     operatorAddress[1],
				TokenAddress: toAddress[1],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc1155_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing chain_id field, but got none")
		}
	})

	t.Run("AddInvalidERC1155RowWithEmptyFromAddress", func(t *testing.T) {
		row := []ERC1155{
			{
				ChainID:      "1",
				ParentHash:   txHash[1],
				BlockNumber:  1,
				ToAddress:    toAddress[1],
				Operator:     operatorAddress[1],
				TokenAddress: toAddress[1],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc1155_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing from_address field, but got none")
		}
	})

	t.Run("AddInvalidERC1155RowWithEmptyParentHash", func(t *testing.T) {
		row := []ERC1155{
			{
				ChainID:      "1",
				BlockNumber:  1,
				FromAddress:  fromAddress[1],
				ToAddress:    toAddress[1],
				Operator:     operatorAddress[1],
				TokenAddress: toAddress[1],
				TokenID:      big.NewInt(123),
				LogIndex:     1,
				TxLogIndex:   2,
				Timestamp:    time.Now(),
			},
		}

		err := client.Add("erc1155_transfers_test", row)
		if err == nil {
			t.Fatalf("expected error for missing parent_hash field, but got none")
		}
	})

}

// func TestClickHouseClient_Read(t *testing.T) {
// 	container := &chContainer{t: t}
// 	container.SetupClickHouseContainer()
// 	defer container.TerminateClickHouseContainer()

// 	client, err := NewClickHouseClient(&container.config, &container.config)
// 	if err != nil {
// 		t.Fatalf("failed to create ClickHouse client: %v", err)
// 	}

// 	testTable := "transactions_test"
// 	err = setupTablesForTests(client, container.ctx)
// 	if err != nil {
// 		t.Fatalf("failed to setup tables for tests: %v", err)
// 	}

// 	convertString := func(input string) string {
// 		bytes, err := convertToBytes(input)
// 		if err != nil {
// 			t.Fatalf("failed to convert to bytes: %v", err)
// 		}
// 		return string(bytes)
// 	}

// 	txHash := []string{
// 		convertString("0x5f3b03d742a967acbdcf013fa3fd63669f4b89a5f0e77dbd6db3877b07c16bb0"),
// 	}
// 	fromAddress := []string{
// 		convertString("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"),
// 	}
// 	toAddress := []string{
// 		convertString("0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"),
// 	}

// 	row := []Transaction{
// 		{
// 			ChainID:     fmt.Sprintf("%d", 1),
// 			TxIndex:     1,
// 			BlockNumber: 1,
// 			TxHash:      txHash[0],
// 			FromAddress: fromAddress[0],
// 			ToAddress:   toAddress[0],
// 			Timestamp:   time.Date(2024, 11, 30, 0, 0, 0, 0, time.UTC),
// 		},
// 	}

// 	if err := client.Add(testTable, row); err != nil {
// 		t.Fatalf("failed to add test rows: %v", err)
// 	}

// 	t.Run("ReadByFromAddress", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		from_address = CAST(unhex('5f4ec3df9cbd43714fe2740f5e3616155c5b8419'), 'FixedString(20)')
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})

// 	t.Run("ReadByToAddress", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		to_address = CAST(unhex('bce206cae7f0ec07b545edde332a47c2f75bbeb3'), 'FixedString(20)')
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})

// 	t.Run("ReadByFromOrToAddress", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		from_address = CAST(unhex('bce206cae7f0ec07b545edde332a47c2f75bbeb3'), 'FixedString(20)')
// 		OR
// 		to_address = CAST(unhex('bce206cae7f0ec07b545edde332a47c2f75bbeb3'), 'FixedString(20)')
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})

// 	t.Run("ReadByTxHash", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		tx_hash = CAST(unhex('5f3b03d742a967acbdcf013fa3fd63669f4b89a5f0e77dbd6db3877b07c16bb0'), 'FixedString(32)')
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})

// 	t.Run("ReadByChainID", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		chain_id = '1'
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})

// 	t.Run("ReadByTimestamp", func(t *testing.T) {
// 		var transaction []Transaction
// 		query := fmt.Sprintf(`
// 		SELECT * FROM %s WHERE
// 		toDate(timestamp) = '2024-11-30'
// 		`, testTable)
// 		err := client.Read(query, &transaction)
// 		if err != nil {
// 			t.Fatalf("Read from ClickHouse failed: %v", err)
// 		}
// 		if len(transaction) != 1 {
// 			t.Fatalf("expected 1 row, got %d", len(transaction))
// 		}
// 	})
// }

func convertToBytes(s string) ([]byte, error) {
	bytes, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
