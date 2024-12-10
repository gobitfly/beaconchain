package db

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/erc1155"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"golang.org/x/sync/errgroup"
)

var ClickHouseNativeWriter ch.Conn
var ZERO_ADDRESS_STRING = "0x0000000000000000000000000000000000000000"
var RETRY_DELAY = time.Second * 5
var MAX_RETRY = 5

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
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionZSTD,
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
	})
	if err != nil {
		log.Fatal(err, "Error connecting to clickhouse native writer", 0)
	}
	// verify connection
	ClickHouseTestConnection(dbWriter, writer.Name)

	return dbWriter
}

func ClickHouseTestConnection(db ch.Conn, dataBaseName string) {
	v, err := db.ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping clickhouse database %s: %w", dataBaseName, err), "", 0)
	}
	log.Debugf("connected to clickhouse database %s with version %s", dataBaseName, v)
}

func DumpToClickhouse(data interface{}, table string) error {
	start := time.Now()
	columns, err := ConvertToColumnar(data)
	if err != nil {
		return err
	}
	log.Debugf("converted to columnar in %s", time.Since(start))
	start = time.Now()
	// abort after 3 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `INSERT INTO `+table)
	if err != nil {
		return err
	}
	log.Debugf("prepared batch in %s", time.Since(start))
	start = time.Now()
	defer func() {
		if batch.IsSent() {
			return
		}
		err := batch.Abort()
		if err != nil {
			log.Warnf("failed to abort batch: %v", err)
		}
	}()
	for c := 0; c < len(columns); c++ {
		// type assert to correct type
		log.Debugf("appending column %d", c)
		switch columns[c].(type) {
		case []int64:
			err = batch.Column(c).Append(columns[c].([]int64))
		case []uint64:
			err = batch.Column(c).Append(columns[c].([]uint64))
		case []time.Time:
			// appending unix timestamps as int64 to a DateTime column is actually faster than appending time.Time directly
			// tho with how many columns we have it doesn't really matter
			err = batch.Column(c).Append(columns[c].([]time.Time))
		case []float64:
			err = batch.Column(c).Append(columns[c].([]float64))
		case []bool:
			err = batch.Column(c).Append(columns[c].([]bool))
		default:
			// warning: slow path. works but try to avoid this
			cType := reflect.TypeOf(columns[c])
			log.Warnf("fallback: column %d of type %s is not natively supported, falling back to reflection", c, cType)
			startSlow := time.Now()
			cValue := reflect.ValueOf(columns[c])
			length := cValue.Len()
			cSlice := reflect.MakeSlice(reflect.SliceOf(cType.Elem()), length, length)
			for i := 0; i < length; i++ {
				cSlice.Index(i).Set(cValue.Index(i))
			}
			err = batch.Column(c).Append(cSlice.Interface())
			log.Debugf("fallback: appended column %d in %s", c, time.Since(startSlow))
		}
		if err != nil {
			return err
		}
	}
	log.Debugf("appended all columns to batch in %s", time.Since(start))
	start = time.Now()
	err = batch.Send()
	if err != nil {
		return err
	}
	log.Debugf("sent batch in %s", time.Since(start))
	return nil
}

// ConvertToColumnar efficiently converts a slice of any struct type to a slice of slices, each representing a column.
func ConvertToColumnar(data interface{}) ([]interface{}, error) {
	start := time.Now()
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("provided data is not a slice")
	}

	if v.Len() == 0 {
		return nil, fmt.Errorf("slice is empty")
	}

	elemType := v.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("slice elements are not structs")
	}

	numFields := elemType.NumField()
	columns := make([]interface{}, numFields)
	colValues := make([]reflect.Value, numFields)

	for i := 0; i < numFields; i++ {
		fieldType := elemType.Field(i).Type
		colSlice := reflect.MakeSlice(reflect.SliceOf(fieldType), v.Len(), v.Len())
		x := reflect.New(colSlice.Type())
		x.Elem().Set(colSlice)
		columns[i] = colSlice
		colValues[i] = colSlice.Slice(0, v.Len())
	}

	var wg sync.WaitGroup
	wg.Add(numFields)

	for j := 0; j < numFields; j++ {
		go func(j int) {
			defer wg.Done()
			for i := 0; i < v.Len(); i++ {
				structValue := v.Index(i)
				colValues[j].Index(i).Set(structValue.Field(j))
			}
		}(j)
	}
	wg.Wait()

	for i, col := range colValues {
		columns[i] = col.Interface()
	}
	log.Infof("columnarized %d rows with %d columns in %s", v.Len(), numFields, time.Since(start))
	return columns, nil
}

func IndexTxsToClickHouseFromBigtable(start, end, concurrency int64) error {
	g := new(errgroup.Group)
	g.SetLimit(int(concurrency))

	log.Infof("ClickHouse indexing blocks from %d to %d", start, end)
	batchSize := int64(100)
	transformerList := []string{"TransformTx", "TransformItx", "TransformERC20", "TransformERC721", "TransformERC1155"}

	for i := start; i <= end; i += batchSize {
		firstBlock := i
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > end {
			lastBlock = end
		}

		g.Go(func() error {
			// create a buffered channel to handle blocks
			blocksChan := make(chan *types.Eth1Block, batchSize)

			go func(stream chan *types.Eth1Block) {
				log.Infof("Querying blocks from %v to %v", firstBlock, lastBlock)
				high := lastBlock
				low := lastBlock - batchSize + 1
				if firstBlock > low {
					low = firstBlock
				}

				err := BigtableClient.GetFullBlocksDescending(stream, uint64(high), uint64(low))
				if err != nil {
					log.Error(err, "error getting blocks descending", 0, map[string]interface{}{"high": high, "low": low})
				}
				close(stream)
			}(blocksChan)

			subG := new(errgroup.Group)
			subG.SetLimit(int(concurrency))

			var blockCount int64
			var erc20Batch []ERC20Batch
			var txBatch []TxBatch
			var itxBatch []InternalTxBatch
			var erc721Batch []ERC721Batch
			var erc1155Batch []ERC1155Batch

			for b := range blocksChan {
				block := b
				subG.Go(func() error {
					err := ParseDataToClickHouse(block, transformerList, &txBatch, &itxBatch, &erc20Batch, &erc721Batch, &erc1155Batch)
					if err != nil {
						log.Error(err, "error saving transactions to ClickHouse", 0)
						return err
					}

					blockCount++

					return nil
				})
			}

			if err := subG.Wait(); err != nil {
				return fmt.Errorf("block processing error: %w", err)
			}

			if blockCount >= batchSize || i == end && end-start < batchSize {
				// send transactions to ClickHouse
				err := SendAllBatches(txBatch, itxBatch, erc20Batch, erc721Batch, erc1155Batch)
				if err != nil {
					log.Error(err, "error sending batches to ClickHouse", 0)
					return err
				}

				// reset blockCount after sending the batches to ClickHouse
				blockCount = 0
			}

			return nil
		})
	}

	// wait for all main goroutines to finish
	if err := g.Wait(); err != nil {
		log.Error(err, "ClickHouse wait group error", 0)
		return err
	}

	// check if last block in cache is updated correctly
	lastBlockInCache, err := BigtableClient.GetLastBlockInDataTable()
	if err != nil {
		log.Error(err, "failed to get last block in data table in bigTable", 0)
		return err
	}

	if end > int64(lastBlockInCache) {
		err := BigtableClient.SetLastBlockInDataTable(end)
		if err != nil {
			log.Error(err, "failed to set last block in data table in bigTable", 0)
			return err
		}
	}

	log.Infof("Clickhouse transactions indexing completed")
	return nil
}

func IndexTxsToClickHouseFromRawBigtable(start, end, concurrency int64) error {
	g := new(errgroup.Group)
	g.SetLimit(int(concurrency))

	log.Infof("ClickHouse indexing blocks from %d to %d", start, end)
	batchSize := int64(100)
	transformerList := []string{"TransformTx", "TransformItx", "TransformERC20", "TransformERC721", "TransformERC1155"}

	project, instance := utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance
	bt, err := database.NewBigTable(project, instance, nil)
	if err != nil {
		panic(err)
	}

	store := raw.NewStore(database.Wrap(bt, raw.Table))

	chainId := big.NewInt(int64(utils.Config.Chain.Id))
	rawStore := rpc.NewRawStoreClient(chainId, store)

	for i := start; i <= end; i += batchSize {
		firstBlock := i
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > end {
			lastBlock = end
		}

		g.Go(func() error {
			// create a buffered channel to handle blocks
			blocksChan := make(chan *types.Eth1Block, batchSize)

			go func(stream chan *types.Eth1Block) {
				log.Infof("Querying blocks from %v to %v", firstBlock, lastBlock)

				for blockNum := firstBlock; blockNum <= lastBlock; blockNum++ {
					block, err := rawStore.GetBlock(blockNum, "geth")
					if err != nil {
						log.Error(err, "error getting block from raw bigTable", 0, map[string]interface{}{"block": blockNum})
						continue
					}
					stream <- block
				}

				time.Sleep(time.Second)
				close(stream)

			}(blocksChan)

			subG := new(errgroup.Group)
			subG.SetLimit(int(concurrency))

			var blockCount int64
			var erc20Batch []ERC20Batch
			var txBatch []TxBatch
			var itxBatch []InternalTxBatch
			var erc721Batch []ERC721Batch
			var erc1155Batch []ERC1155Batch

			for b := range blocksChan {
				block := b
				subG.Go(func() error {
					err = ParseDataToClickHouse(block, transformerList, &txBatch, &itxBatch, &erc20Batch, &erc721Batch, &erc1155Batch)
					if err != nil {
						log.Error(err, "error saving transactions to ClickHouse", 0)
						return err
					}

					blockCount++
					return nil
				})
			}

			if err := subG.Wait(); err != nil {
				return fmt.Errorf("block processing error: %w", err)
			}

			if blockCount == batchSize || i == end && end-start < batchSize {
				// send transactions to ClickHouse
				err := SendAllBatches(txBatch, itxBatch, erc20Batch, erc721Batch, erc1155Batch)
				if err != nil {
					log.Error(err, "error sending batches to ClickHouse", 0)
					return err
				}

				// reset blockCount after sending the batches to ClickHouse
				blockCount = 0
			}

			return nil
		})
	}

	// wait for all main goroutines to finish
	if err := g.Wait(); err != nil {
		log.Error(err, "ClickHouse wait group error", 0)
		return err
	}

	log.Infof("Clickhouse transactions indexing from raw bigTable completed")
	return nil
}

func SendTxBatch(data []TxBatch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	txBatch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO transactions (
			chain_id, tx_index, tx_hash, block_number, from_address, to_address, type, method, value, nonce,
			status, timestamp, tx_fee, gas, gas_price, gas_used, max_fee_per_gas, max_priority_fee_per_gas,
			max_fee_per_blob_gas, blob_gas_price, blob_gas_used, blob_tx_fee, blob_versioned_hashes,
			access_list, input_data, is_contract_creation, logs, logs_bloom
		)
	`)
	if err != nil {
		return fmt.Errorf("error while preparing TX batch for ClickHouse: %v", err)
	}

	for _, d := range data {
		err := txBatch.Append(d.ChainID, d.TxIndex, d.TxHash, d.BlockNumber, d.FromAddress, d.ToAddress,
			d.Type, d.Method, d.Value, d.Nonce, d.Status, d.Timestamp, d.TxFee, d.Gas, d.GasPrice, d.GasUsed,
			d.MaxFeePerGas, d.MaxPriorityFeePerGas, d.MaxFeePerBlobGas, d.BlobGasPrice, d.BlobGasUsed, d.BlobTxFee,
			d.BlobVersionedHashes, d.AccessList, d.InputData, d.IsContractCreation, d.Logs, d.LogsBloom)
		if err != nil {
			return fmt.Errorf("error while appending TX batch for ClickHouse: %v", err)
		}
	}

	err = txBatch.Send()
	if err != nil {
		return fmt.Errorf("failed to send TX batch to ClickHouse: %v", err)
	}

	log.Infof("Sent TX data to ClickHouse successfully")

	return nil
}

func SendInternalTxBatch(data []InternalTxBatch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	itxBatch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO internal_transactions (
			chain_id, parent_hash, block_number, from_address, to_address,
			type, value, path, gas, timestamp, error_msg
		)
	`)
	if err != nil {
		return fmt.Errorf("error while preparing ITX batch for ClickHouse: %v", err)
	}

	for _, d := range data {
		err := itxBatch.Append(d.ChainID, d.ParentHash, d.BlockNumber, d.FromAddress,
			d.ToAddress, d.Type, d.Value, d.Path, d.Gas, d.Timestamp, d.ErrorMsg)
		if err != nil {
			return fmt.Errorf("error while appending ITX batch for ClickHouse: %v", err)
		}
	}

	err = itxBatch.Send()
	if err != nil {
		return fmt.Errorf("failed to send ITX batch to ClickHouse: %v", err)
	}

	log.Infof("Sent ITX data to ClickHouse successfully")

	return nil
}

func SendERC20Batch(data []ERC20Batch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	erc20Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO erc20_transfers (
			chain_id, parent_hash, block_number, from_address, to_address, token_address,
			value, log_index, log_type, transaction_log_index, removed, timestamp
		)
	`)
	if err != nil {
		return fmt.Errorf("error while preparing ERC20 batch for ClickHouse: %v", err)
	}

	for _, d := range data {
		err := erc20Batch.Append(d.ChainID, d.ParentHash, d.BlockNumber, d.FromAddress, d.ToAddress,
			d.TokenAddress, d.Value, d.LogIndex, d.LogType, d.TxLogIndex, d.Removed, d.Timestamp)
		if err != nil {
			return fmt.Errorf("error while appending ERC20 batch for ClickHouse: %v", err)
		}
	}

	err = erc20Batch.Send()
	if err != nil {
		return fmt.Errorf("failed to send ERC20 batch to ClickHouse: %v", err)
	}

	log.Infof("Sent ERC20 data to ClickHouse successfully")

	return nil
}

func SendERC721Batch(data []ERC721Batch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	erc721Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO erc721_transfers (
			chain_id, parent_hash, block_number, from_address, to_address,
			token_address, token_id, log_index, log_type, transaction_log_index, removed, timestamp
		)
	`)
	if err != nil {
		return fmt.Errorf("error while preparing ERC721 batch for ClickHouse: %v", err)
	}

	for _, d := range data {
		err := erc721Batch.Append(d.ChainID, d.ParentHash, d.BlockNumber, d.FromAddress, d.ToAddress,
			d.TokenAddress, d.TokenID, d.LogIndex, d.LogType, d.TxLogIndex, d.Removed, d.Timestamp)
		if err != nil {
			return fmt.Errorf("error while appending ERC721 batch for ClickHouse: %v", err)
		}
	}

	err = erc721Batch.Send()
	if err != nil {
		return fmt.Errorf("failed to send ERC721 batch to ClickHouse: %v", err)
	}

	log.Infof("Sent ERC721 data to ClickHouse successfully")

	return nil
}

func SendERC1155Batch(data []ERC1155Batch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	erc1155Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO erc1155_transfers (
			chain_id, parent_hash, block_number, from_address, to_address, operator, token_address, 
			token_id, value, log_index, log_type, transaction_log_index, removed, timestamp
		)
	`)
	if err != nil {
		return fmt.Errorf("error while preparing ERC1155 batch for ClickHouse: %v", err)
	}

	for _, d := range data {
		err := erc1155Batch.Append(d.ChainID, d.ParentHash, d.BlockNumber, d.FromAddress, d.ToAddress,
			d.Operator, d.TokenAddress, d.TokenID, d.Value, d.LogIndex, d.LogType, d.TxLogIndex, d.Removed, d.Timestamp)
		if err != nil {
			return fmt.Errorf("error while appending ERC1155 batch for ClickHouse: %v", err)
		}
	}

	err = erc1155Batch.Send()
	if err != nil {
		return fmt.Errorf("failed to send ERC1155 batch to ClickHouse: %v", err)
	}

	log.Infof("Sent ERC1155 data to ClickHouse successfully")

	return nil
}

func SendAllBatches(txBatch []TxBatch, itxBatch []InternalTxBatch, erc20Batch []ERC20Batch, erc721Batch []ERC721Batch, erc1155Batch []ERC1155Batch) error {
	if txBatch != nil {
		err := SendTxBatch(txBatch)
		if err != nil {
			return fmt.Errorf("error while sending TX Batch to ClickHouse: %v", err)
		}
	}

	if itxBatch != nil {
		err := SendInternalTxBatch(itxBatch)
		if err != nil {
			return fmt.Errorf("error while sending ITX Batch to ClickHouse: %v", err)
		}
	}

	if erc20Batch != nil {
		err := SendERC20Batch(erc20Batch)
		if err != nil {
			return fmt.Errorf("error while sending ERC20 Batch to ClickHouse: %v", err)
		}
	}

	if erc721Batch != nil {
		err := SendERC721Batch(erc721Batch)
		if err != nil {
			return fmt.Errorf("error while sending ERC721 Batch to ClickHouse: %v", err)
		}
	}

	if erc1155Batch != nil {
		err := SendERC1155Batch(erc1155Batch)
		if err != nil {
			return fmt.Errorf("error while sending ERC1155 Batch to ClickHouse: %v", err)
		}
	}

	return nil
}

func mapStatusToEnum(status uint64) string {
	switch status {
	case 0:
		return "failed"
	case 1:
		return "success"
	case 2:
		return "partially failed"
	default:
		return "unknown"
	}
}

type TxBatch struct {
	ChainID              string
	TxIndex              int
	TxHash               string
	BlockNumber          uint64
	FromAddress          string
	ToAddress            string
	Type                 string
	Method               string
	Value                uint64
	Nonce                uint64
	Status               string
	Timestamp            int64
	TxFee                *big.Int
	Gas                  uint64
	GasPrice             uint64
	GasUsed              uint64
	MaxFeePerGas         uint64
	MaxPriorityFeePerGas uint64
	MaxFeePerBlobGas     uint64
	BlobGasPrice         uint64
	BlobGasUsed          uint64
	BlobTxFee            *big.Int
	BlobVersionedHashes  []string
	AccessList           []*types.AccessList
	InputData            []byte
	IsContractCreation   bool
	Logs                 []*types.Eth1Log
	LogsBloom            []byte
}

type InternalTxBatch struct {
	ChainID     string
	ParentHash  string
	BlockNumber uint64
	FromAddress string
	ToAddress   string
	Type        string
	Value       string
	Path        string
	Gas         uint64
	Timestamp   int64
	ErrorMsg    string
}

type ERC20Batch struct {
	ChainID      string
	ParentHash   string
	BlockNumber  uint64
	FromAddress  string
	ToAddress    string
	TokenAddress string
	Value        *big.Int
	LogIndex     uint64
	LogType      string
	TxLogIndex   uint64
	Removed      bool
	Timestamp    int64
}

type ERC721Batch struct {
	ChainID      string
	ParentHash   string
	BlockNumber  uint64
	FromAddress  string
	ToAddress    string
	TokenAddress string
	TokenID      *big.Int
	LogIndex     uint64
	LogType      string
	TxLogIndex   uint64
	Removed      bool
	Timestamp    int64
}

type ERC1155Batch struct {
	ChainID      string
	ParentHash   string
	BlockNumber  uint64
	FromAddress  string
	ToAddress    string
	Operator     string
	TokenAddress string
	TokenID      *big.Int
	Value        *big.Int
	LogIndex     uint64
	LogType      string
	TxLogIndex   uint64
	Removed      bool
	Timestamp    int64
}

func ParseDataToClickHouse(block *types.Eth1Block, transformerList []string, txBatch *[]TxBatch, itxBatch *[]InternalTxBatch, erc20Batch *[]ERC20Batch, erc721Batch *[]ERC721Batch, erc1155Batch *[]ERC1155Batch) error {
	// set values for logging
	startTs := time.Now()
	lastTickTs := time.Now()
	processedTxs := int64(0)
	var currentProcessed int64

	for i, tx := range block.Transactions {
		for _, transformer := range transformerList {
			switch transformer {
			case "TransformTx":
				err := parseTxs(tx, i, block.Number, block.Time.Seconds, txBatch)
				if err != nil {
					log.Error(err, "error while processing TX", 0)
				}
				currentProcessed = atomic.AddInt64(&processedTxs, 1)

			case "TransformItx":
				err := parseItx(tx, block.Number, block.Time.Seconds, itxBatch)
				if err != nil {
					log.Error(err, "error while processing ITX", 0)
				}
			case "TransformERC20":
				err := parseERC20Transfers(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds, erc20Batch)
				if err != nil {
					log.Error(err, "error while processing ERC20 transfers", 0)
				}
			case "TransformERC721":
				err := parseERC721Transfers(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds, erc721Batch)
				if err != nil {
					log.Error(err, "error while processing ERC721 transfers", 0)
				}
			case "TransformERC1155":
				err := parseERC1155Transfers(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds, erc1155Batch)
				if err != nil {
					log.Error(err, "error while processing ERC1155 transfers", 0)
				}
			default:
				log.Error(nil, "unknown transformer type", 0)
			}
		}
	}

	log.Infof("processed %v txs in %v (%.1f txs / sec) for block %d", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), block.Number)

	return nil
}

func parseTxs(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockTimestamp int64, txBatch *[]TxBatch) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		// parse contract address
		toAddress := tx.GetTo()
		isContract := false
		if !bytes.Equal(tx.GetContractAddress(), ZERO_ADDRESS) {
			toAddress = tx.GetContractAddress()
			isContract = true
		}

		// parse method type
		method := make([]byte, 0)
		if len(tx.GetData()) > 3 {
			method = tx.GetData()[:4]
		}

		status := mapStatusToEnum(tx.Status)
		txFee := new(big.Int).Mul(big.NewInt(int64(tx.GasPrice)), big.NewInt(int64(tx.GasUsed)))
		blobTxFee := new(big.Int).Mul(big.NewInt(int64(tx.BlobGasPrice)), big.NewInt(int64(tx.BlobGasUsed)))

		var blobVersionedHashes []string
		for _, blob := range tx.BlobVersionedHashes {
			blobHash := common.BytesToHash(blob)
			blobVersionedHashes = append(blobVersionedHashes, blobHash.String())
		}

		data := TxBatch{
			ChainID:              fmt.Sprintf("%d", utils.Config.Chain.Id),
			TxIndex:              txIndex,
			TxHash:               string(tx.Hash),
			BlockNumber:          blockNumber,
			FromAddress:          string(tx.From),
			ToAddress:            string(toAddress),
			Type:                 fmt.Sprintf("0x%x", tx.Type),
			Method:               fmt.Sprintf("%x", string(method)),
			Value:                tx.Value,
			Nonce:                tx.Nonce,
			Status:               status,
			Timestamp:            blockTimestamp,
			TxFee:                txFee,
			Gas:                  tx.Gas,
			GasPrice:             tx.GasPrice,
			GasUsed:              tx.GasUsed,
			MaxFeePerGas:         tx.MaxFeePerGas,
			MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
			MaxFeePerBlobGas:     tx.MaxFeePerBlobGas,
			BlobGasPrice:         tx.BlobGasPrice,
			BlobGasUsed:          tx.BlobGasUsed,
			BlobTxFee:            blobTxFee,
			BlobVersionedHashes:  blobVersionedHashes,
			AccessList:           tx.AccessList,
			InputData:            tx.Data,
			IsContractCreation:   isContract,
			Logs:                 tx.Logs,
			LogsBloom:            tx.LogsBloom,
		}

		*txBatch = append(*txBatch, data)

		// fmt.Printf("\n\n tx batch LEN in HERE %d \n", len(*txBatch))
		// if err != nil {
		// 	log.Error(err, "error appending tx data to batch", 0)
		// 	if attempt < MAX_RETRY {
		// 		log.Warn(fmt.Sprintf("Attempt %d failed to append tx data to batch. Retrying...", attempt))
		// 		time.Sleep(RETRY_DELAY)
		// 		continue
		// 	}
		// 	fmt.Errorf("failed to append tx data to batch after %d attempts: %v", MAX_RETRY, err)
		// 	panic("failed to process tx")
		// }

		return nil
	}

	return fmt.Errorf("failed to process transactions for block %d after %d attempts", blockNumber, MAX_RETRY)
}

func parseItx(tx *types.Eth1Transaction, blockNumber uint64, blockTimestamp int64, itxBatch *[]InternalTxBatch) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		if len(tx.Itx) == 0 {
			return nil
		}

		// set values for logging
		processedItx := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		for _, itx := range tx.Itx {

			data := InternalTxBatch{
				ChainID:     fmt.Sprintf("%d", utils.Config.Chain.Id),
				ParentHash:  string(tx.Hash),
				BlockNumber: blockNumber,
				FromAddress: string(itx.From),
				ToAddress:   string(itx.To),
				Type:        itx.Type,
				Value:       itx.Value,
				Path:        itx.Path,
				Gas:         itx.Gas,
				Timestamp:   blockTimestamp,
				ErrorMsg:    itx.ErrorMsg,
			}

			*itxBatch = append(*itxBatch, data)

			// if err != nil {
			// 	log.Error(err, "error appending ITX data to batch", 0)
			// 	if attempt < MAX_RETRY {
			// 		log.Warn(fmt.Sprintf("Attempt %d failed to append ITX data to batch. Retrying...", attempt))
			// 		time.Sleep(RETRY_DELAY)
			// 		continue
			// 	}
			// 	fmt.Errorf("failed to append ITX data to batch after %d attempts: %v", MAX_RETRY, err)
			// 	panic("failed to process ITX")
			// }
			currentProcessed = atomic.AddInt64(&processedItx, 1)
		}

		log.Infof("processed %v ITXs in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, common.Bytes2Hex(tx.Hash))

		return nil
	}

	return fmt.Errorf("failed to process ITX for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)

}

func parseERC20Transfers(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64, erc20Batch *[]ERC20Batch) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {

		// set values for logging
		processedERC20 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		for j, txLog := range tx.GetLogs() {
			// no events emitted continue
			if len(txLog.GetTopics()) != 3 || !bytes.Equal(txLog.GetTopics()[0], erc20.TransferTopic) {
				continue
			}

			filterer, err := erc20.NewErc20Filterer(common.Address{}, nil)
			if err != nil {
				log.Error(err, "error creating ERC20 filterer", 0)
			}

			topics := make([]common.Hash, 0, len(txLog.GetTopics()))

			for _, lTopic := range txLog.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}
			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(txLog.GetAddress()),
				Data:        txLog.Data,
				Topics:      topics,
				BlockNumber: blockNumber,
				TxHash:      common.BytesToHash(tx.Hash),
				TxIndex:     uint(txIndex),
				BlockHash:   blockHash,
				Index:       uint(j),
				Removed:     txLog.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}
			var value *big.Int
			if transfer != nil && transfer.Value != nil {
				value = transfer.Value
			}

			parsedData := ERC20Batch{
				ChainID:      fmt.Sprintf("%d", utils.Config.Chain.Id),
				ParentHash:   string(tx.Hash),
				BlockNumber:  blockNumber,
				FromAddress:  string(transfer.From.Bytes()),
				ToAddress:    string(transfer.To.Bytes()),
				TokenAddress: string(txLog.Address),
				Value:        value,
				LogIndex:     uint64(j),
				LogType:      string(topics[0].Bytes()),
				TxLogIndex:   uint64(txIndex),
				Removed:      txLog.GetRemoved(),
				Timestamp:    blockTimestamp,
			}

			*erc20Batch = append(*erc20Batch, parsedData)

			// if err != nil {
			// 	log.Error(err, "error appending ERC20 data to batch", 0)
			// 	if attempt < MAX_RETRY {
			// 		log.Warn(fmt.Sprintf("Attempt %d failed to append ERC20 data to batch. Retrying...", attempt))
			// 		time.Sleep(RETRY_DELAY)
			// 		continue
			// 	}
			// 	fmt.Errorf("failed to append ERC20 data to batch after %d attempts: %v", MAX_RETRY, err)
			// 	panic("failed to process ERC20")
			// }

			currentProcessed = atomic.AddInt64(&processedERC20, 1)
		}

		if processedERC20 != 0 {
			log.Infof("processed %v ERC20 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, common.Bytes2Hex(tx.Hash))
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC20 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}

func parseERC721Transfers(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64, erc721Batch *[]ERC721Batch) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		// set values for logging
		processedERC721 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		for j, txLog := range tx.GetLogs() {
			// no events emitted continue
			if len(txLog.GetTopics()) != 4 || !bytes.Equal(txLog.GetTopics()[0], erc721.TransferTopic) {
				continue
			}

			filterer, err := erc721.NewErc721Filterer(common.Address{}, nil)
			if err != nil {
				log.Error(err, "error creating ERC721 filterer", 0)
			}

			topics := make([]common.Hash, 0, len(txLog.GetTopics()))

			for _, lTopic := range txLog.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(txLog.GetAddress()),
				Data:        txLog.Data,
				Topics:      topics,
				BlockNumber: blockNumber,
				TxHash:      common.BytesToHash(tx.Hash),
				TxIndex:     uint(txIndex),
				BlockHash:   blockHash,
				Index:       uint(j),
				Removed:     txLog.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			tokenId := new(big.Int)
			if transfer != nil && transfer.TokenId != nil {
				tokenId = transfer.TokenId
			}

			data := ERC721Batch{
				ChainID:      fmt.Sprintf("%d", utils.Config.Chain.Id),
				ParentHash:   string(tx.Hash),
				BlockNumber:  blockNumber,
				FromAddress:  string(transfer.From.Bytes()),
				ToAddress:    string(transfer.To.Bytes()),
				TokenAddress: string(txLog.Address),
				TokenID:      tokenId,
				LogIndex:     uint64(j),
				LogType:      string(topics[0].Bytes()),
				TxLogIndex:   uint64(txIndex),
				Removed:      txLog.GetRemoved(),
				Timestamp:    blockTimestamp,
			}

			*erc721Batch = append(*erc721Batch, data)

			// if err != nil {
			// 	log.Error(err, "error appending ERC721 data to batch", 0)
			// 	if attempt < MAX_RETRY {
			// 		log.Warn(fmt.Sprintf("Attempt %d failed to append ERC721 data to batch. Retrying...", attempt))
			// 		time.Sleep(RETRY_DELAY)
			// 		continue
			// 	}
			// 	fmt.Errorf("failed to append ERC721 data to batch after %d attempts: %v", MAX_RETRY, err)
			// 	panic("failed to process ERC721")
			// }

			currentProcessed = atomic.AddInt64(&processedERC721, 1)
		}

		if processedERC721 != 0 {
			log.Infof("processed %v ERC721 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, common.Bytes2Hex(tx.Hash))
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC721 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}

func parseERC1155Transfers(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64, erc1155Batch *[]ERC1155Batch) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {

		// set values for logging
		processedERC1155 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		for j, txLog := range tx.GetLogs() {
			// no events emitted continue
			if len(txLog.GetTopics()) != 4 || (!bytes.Equal(txLog.GetTopics()[0], erc1155.TransferBulkTopic) && !bytes.Equal(txLog.GetTopics()[0], erc1155.TransferSingleTopic)) {
				continue
			}

			filterer, err := erc1155.NewErc1155Filterer(common.Address{}, nil)
			if err != nil {
				log.Error(err, "error creating ERC1155 filterer", 0)
			}

			topics := make([]common.Hash, 0, len(txLog.GetTopics()))

			for _, lTopic := range txLog.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(txLog.GetAddress()),
				Data:        txLog.Data,
				Topics:      topics,
				BlockNumber: blockNumber,
				TxHash:      common.BytesToHash(tx.Hash),
				TxIndex:     uint(txIndex),
				BlockHash:   blockHash,
				Index:       uint(j),
				Removed:     txLog.GetRemoved(),
			}

			transferBatch, _ := filterer.ParseTransferBatch(ethLog)
			transferSingle, _ := filterer.ParseTransferSingle(ethLog)
			if transferBatch == nil && transferSingle == nil {
				continue
			}

			if transferBatch != nil {
				if len(transferBatch.Ids) != len(transferBatch.Values) {
					log.Error(fmt.Errorf("error parsing ERC1155 batch transfer logs. Expected len(ids): %v len(values): %v to be the same", len(transferBatch.Ids), len(transferBatch.Values)), "", 0)
					continue
				}

				for index := range transferBatch.Ids {

					data := ERC1155Batch{
						ChainID:      fmt.Sprintf("%d", utils.Config.Chain.Id),
						ParentHash:   string(tx.Hash),
						BlockNumber:  blockNumber,
						FromAddress:  string(transferBatch.From.Bytes()),
						ToAddress:    string(transferBatch.To.Bytes()),
						Operator:     string(transferBatch.Operator.Bytes()),
						TokenAddress: string(txLog.Address),
						TokenID:      transferBatch.Ids[index],
						Value:        transferBatch.Values[index],
						LogIndex:     uint64(j),
						LogType:      string(topics[0].Bytes()),
						TxLogIndex:   uint64(txIndex),
						Removed:      txLog.GetRemoved(),
						Timestamp:    blockTimestamp,
					}

					*erc1155Batch = append(*erc1155Batch, data)

					// if err != nil {
					// 	log.Error(err, "error appending ERC1155 data to batch", 0)
					// 	if attempt < MAX_RETRY {
					// 		log.Warn(fmt.Sprintf("Attempt %d failed to append ERC1155 data to batch. Retrying...", attempt))
					// 		time.Sleep(RETRY_DELAY)
					// 		continue
					// 	}
					// 	fmt.Errorf("failed to append ERC1155 data to batch after %d attempts: %v", MAX_RETRY, err)
					// 	panic("failed to process ERC1155")
					// }

					currentProcessed = atomic.AddInt64(&processedERC1155, 1)
				}
			} else if transferSingle != nil {

				data := ERC1155Batch{
					ChainID:      fmt.Sprintf("%d", utils.Config.Chain.Id),
					ParentHash:   string(tx.Hash),
					BlockNumber:  blockNumber,
					FromAddress:  string(transferSingle.From.Bytes()),
					ToAddress:    string(transferSingle.To.Bytes()),
					Operator:     string(transferSingle.Operator.Bytes()),
					TokenAddress: string(txLog.Address),
					TokenID:      transferSingle.Id,
					Value:        transferSingle.Value,
					LogIndex:     uint64(j),
					LogType:      string(topics[0].Bytes()),
					TxLogIndex:   uint64(txIndex),
					Removed:      txLog.GetRemoved(),
					Timestamp:    blockTimestamp,
				}

				*erc1155Batch = append(*erc1155Batch, data)
				// if err != nil {
				// 	log.Error(err, "error appending ERC1155 data to batch", 0)
				// 	if attempt < MAX_RETRY {
				// 		log.Warn(fmt.Sprintf("Attempt %d failed to append ERC1155 data to batch. Retrying...", attempt))
				// 		time.Sleep(RETRY_DELAY)
				// 		continue
				// 	}
				// 	fmt.Errorf("failed to append ERC1155 data to batch after %d attempts: %v", MAX_RETRY, err)
				// 	panic("failed to process ERC1155")
				// }

				currentProcessed = atomic.AddInt64(&processedERC1155, 1)
			}
		}

		if processedERC1155 != 0 {
			log.Infof("processed %v ERC1155 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, common.Bytes2Hex(tx.Hash))
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC1155 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}
