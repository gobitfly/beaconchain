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
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/pkg/commons/erc1155"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"golang.org/x/sync/errgroup"

	"google.golang.org/protobuf/types/known/timestamppb"
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
	batchSize := int64(10)
	transformerList := []string{"TransformTx", "TransformItx", "TransformERC20", "TransformERC721", "TransformERC1155"}

	// Optimized concurrency for fetching blocks and saving to ClickHouse
	for i := start; i <= end; i += batchSize {
		firstBlock := i
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > end {
			lastBlock = end
		}

		g.Go(func() error {
			// Create a buffered channel to handle blocks efficiently
			blocksChan := make(chan *types.Eth1Block, batchSize)

			// Fetch blocks asynchronously
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

			// Use another goroutine to process transactions concurrently
			subG := new(errgroup.Group)
			subG.SetLimit(int(concurrency))

			for b := range blocksChan {
				block := b
				subG.Go(func() error {
					err := SaveTransactionsToClickHouse(block, transformerList)
					if err != nil {
						log.Error(err, "error saving transactions to ClickHouse", 0)
						return err
					}
					return nil
				})
			}

			return subG.Wait()
		})
	}

	// Wait for all main goroutines to finish
	if err := g.Wait(); err != nil {
		log.Error(err, "ClickHouse wait group error", 0)
		return err
	}

	// Check if last block in cache is updated correctly
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

type Eth1Transaction struct {
	Hash                 []byte
	From                 []byte
	To                   []byte
	Value                int64
	Gas                  int64
	GasPrice             *int64
	MaxFeePerGas         *int64
	MaxPriorityFeePerGas *int64
	Nonce                uint32
	Type                 string
	Method               string
	Status               uint8
	InputData            []byte
	ContractCreated      []byte
	Logs                 []byte
	LogsBloom            []byte
	Timestamp            *timestamppb.Timestamp
	BlobGasPrice         *int64
	BlobGasUsed          *int64
	InternalData         []*InternalTransaction
}

type InternalTransaction struct {
	FromAddress *string
	ToAddress   *string
	Type        *string
	Value       *string
	Path        *string
	GasLimit    *int64
	ErrorMsg    *string
}

func SaveTransactionsToClickHouse(block *types.Eth1Block, transformerList []string) error {
	// set values for logging
	startTs := time.Now()
	lastTickTs := time.Now()
	processedTxs := int64(0)
	var currentProcessed int64

	for i, tx := range block.Transactions {
		for _, transformer := range transformerList {
			switch transformer {
			case "TransformTx":
				err := saveTxsToClickHouse(tx, i, block.Number, block.Time.Seconds)
				if err != nil {
					log.Error(err, "error while processing tx", 0)
				}
				currentProcessed = atomic.AddInt64(&processedTxs, 1)

			case "TransformItx":
				err := saveItxToClickHouse(tx, block.Number, block.Time.Seconds)
				if err != nil {
					log.Error(err, "error while processing itx", 0)
				}
			case "TransformERC20":
				err := saveERC20ToClickHouse(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds)
				if err != nil {
					log.Error(err, "error while processing ERC20 transfers", 0)
				}
			case "TransformERC721":
				err := saveERC721ToClickHouse(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds)
				if err != nil {
					log.Error(err, "error while processing ERC721 transfers", 0)
				}
			case "TransformERC1155":
				err := saveERC1155ToClickHouse(tx, i, block.Number, common.BytesToHash(block.GetHash()), block.Time.Seconds)
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

func prepareTxBatch(ctx context.Context) (driver.Batch, error) {
	txBatch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
		INSERT INTO transactions_ethereum (
			tx_index, tx_hash, block_number, from_address, to_address, type, method, value, nonce, status,
			timestamp, tx_fee, gas, gas_price, gas_used, max_fee_per_gas, max_priority_fee_per_gas,
			max_fee_per_blob_gas, blob_gas_price, blob_gas_used, blob_tx_fee, blob_versioned_hashes,
			access_list, input_data, is_contract_creation, logs, logs_bloom
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing tx batch for ClickHouse: %v", err)
	}

	return txBatch, nil
}

func prepareItxBatch(ctx context.Context) (driver.Batch, error) {
	itxBatch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
	INSERT INTO internal_tx_ethereum (parent_hash, block_number, from_address, to_address, type, value,
			path, gas, timestamp, error_msg)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing itx batch for ClickHouse: %v", err)
	}

	return itxBatch, nil
}

func prepareERC20Batch(ctx context.Context) (driver.Batch, error) {
	erc20Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
	INSERT INTO erc20_ethereum (parent_hash, block_number, from_address, to_address, token_address,
			value, log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC20 batch for ClickHouse: %v", err)
	}

	return erc20Batch, nil
}

func prepareERC721Batch(ctx context.Context) (driver.Batch, error) {
	erc721Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
	INSERT INTO erc721_ethereum (parent_hash, block_number, from_address, to_address, token_address,
			 token_id, log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC721 batch for ClickHouse: %v", err)
	}

	return erc721Batch, nil
}

func prepareERC1155Batch(ctx context.Context) (driver.Batch, error) {
	erc1155Batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `
	INSERT INTO erc1155_ethereum (parent_hash, block_number, from_address, to_address, operator, token_address, 
			token_id, value, log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC1155 batch for ClickHouse: %v", err)
	}

	return erc1155Batch, nil
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

func saveTxsToClickHouse(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockTimestamp int64) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// prepare tx batch to send to ClickHouse
		txBatch, err := prepareTxBatch(ctx)
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to prepare tx batch: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to prepare tx batch after %d attempts: %v", MAX_RETRY, err)
		}

		// parse contract address
		toAddress := tx.GetTo()
		isContract := false
		if tx.GetContractAddress() != ZERO_ADDRESS_STRING {
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

		err = txBatch.Append(
			txIndex,
			string(tx.Hash),
			blockNumber,
			string(tx.From),
			toAddress,
			fmt.Sprintf("0x%x", tx.Type),
			fmt.Sprintf("%x", string(method)),
			tx.Value,
			tx.Nonce,
			status,
			blockTimestamp,
			txFee,
			tx.Gas,
			tx.GasPrice,
			tx.GasUsed,
			tx.MaxFeePerGas,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerBlobGas,
			tx.BlobGasPrice,
			tx.BlobGasUsed,
			blobTxFee,
			blobVersionedHashes,
			tx.AccessList,
			tx.Data,
			isContract,
			tx.Logs,
			tx.LogsBloom,
		)

		if err != nil {
			log.Error(err, "error appending tx data to batch", 0)
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to append tx data to batch. Retrying...", attempt))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to append tx data to batch after %d attempts: %v", MAX_RETRY, err)
		}

		// send the tx batch to ClickHouse
		err = txBatch.Send()
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to send tx batch to ClickHouse: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to send tx batch to ClickHouse after %d attempts: %v", MAX_RETRY, err)
		}

		return nil
	}

	return fmt.Errorf("failed to process transactions for block %d after %d attempts", blockNumber, MAX_RETRY)
}

func saveItxToClickHouse(tx *types.Eth1Transaction, blockNumber uint64, blockTimestamp int64) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if len(tx.Itx) == 0 {
			return nil
		}

		// set values for logging
		processedItx := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		// prepare itx batch to send to ClickHouse
		itxBatch, err := prepareItxBatch(ctx)
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to prepare ITX batch: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to prepare ITX batch after %d attempts: %v", MAX_RETRY, err)
		}

		for _, itx := range tx.Itx {
			err = itxBatch.Append(
				string(tx.Hash),
				blockNumber,
				itx.From,
				itx.To,
				itx.Type,
				itx.Value,
				itx.Path,
				itx.Gas,
				blockTimestamp,
				itx.ErrorMsg,
			)

			if err != nil {
				log.Error(err, "error appending ITX data to batch", 0)
				if attempt < MAX_RETRY {
					log.Warn(fmt.Sprintf("Attempt %d failed to append ITX data to batch. Retrying...", attempt))
					time.Sleep(RETRY_DELAY)
					continue
				}
				return fmt.Errorf("failed to append ITX data to batch after %d attempts: %v", MAX_RETRY, err)
			}
			currentProcessed = atomic.AddInt64(&processedItx, 1)
		}

		// send the itx batch to ClickHouse
		err = itxBatch.Send()
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to send ITX batch to ClickHouse: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to send ITX batch to ClickHouse after %d attempts: %v", MAX_RETRY, err)
		}

		log.Infof("processed %v ITXs in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, tx.Hash)

		return nil
	}

	return fmt.Errorf("failed to process ITX for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)

}

func saveERC20ToClickHouse(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// set values for logging
		processedERC20 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		// prepare ERC20 batch to send to ClickHouse
		erc20Batch, err := prepareERC20Batch(ctx)
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to prepare ERC20 batch: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to prepare ERC20 batch after %d attempts: %v", MAX_RETRY, err)
		}

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
				TxHash:      common.HexToHash(tx.Hash),
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

			err = erc20Batch.Append(
				string(tx.Hash),
				blockNumber,
				transfer.From.String(),
				transfer.To.String(),
				common.BytesToAddress(txLog.Address).String(),
				value,
				uint64(j),
				topics[0].String(),
				uint64(txIndex),
				txLog.GetRemoved(),
				blockTimestamp,
			)

			if err != nil {
				log.Error(err, "error appending ERC20 data to batch", 0)
				if attempt < MAX_RETRY {
					log.Warn(fmt.Sprintf("Attempt %d failed to append ERC20 data to batch. Retrying...", attempt))
					time.Sleep(RETRY_DELAY)
					continue
				}
				return fmt.Errorf("failed to append ERC20 data to batch after %d attempts: %v", MAX_RETRY, err)
			}

			currentProcessed = atomic.AddInt64(&processedERC20, 1)
		}

		// send the ERC20 batch to ClickHouse
		err = erc20Batch.Send()
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to send ERC20 batch to ClickHouse: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to send ERC20 batch to ClickHouse after %d attempts: %v", MAX_RETRY, err)
		}

		if processedERC20 != 0 {
			log.Infof("processed %v ERC20 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, tx.Hash)
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC20 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}

func saveERC721ToClickHouse(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// set values for logging
		processedERC721 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		// prepare ERC721 batch to send to ClickHouse
		erc721Batch, err := prepareERC721Batch(ctx)
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to prepare ERC721 batch: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to prepare ERC721 batch after %d attempts: %v", MAX_RETRY, err)
		}

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
				TxHash:      common.HexToHash(tx.Hash),
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

			err = erc721Batch.Append(
				string(tx.Hash),
				blockNumber,
				transfer.From.String(),
				transfer.To.String(),
				common.Bytes2Hex(txLog.Address),
				tokenId,
				uint64(j),
				topics[0].String(),
				uint64(txIndex),
				txLog.GetRemoved(),
				blockTimestamp,
			)

			if err != nil {
				log.Error(err, "error appending ERC721 data to batch", 0)
				if attempt < MAX_RETRY {
					log.Warn(fmt.Sprintf("Attempt %d failed to append ERC721 data to batch. Retrying...", attempt))
					time.Sleep(RETRY_DELAY)
					continue
				}
				return fmt.Errorf("failed to append ERC721 data to batch after %d attempts: %v", MAX_RETRY, err)
			}

			currentProcessed = atomic.AddInt64(&processedERC721, 1)
		}

		// send the ERC721 batch to ClickHouse
		err = erc721Batch.Send()
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to send ERC721 batch to ClickHouse: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to send ERC721 batch to ClickHouse after %d attempts: %v", MAX_RETRY, err)
		}

		if processedERC721 != 0 {
			log.Infof("processed %v ERC721 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, tx.Hash)
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC721 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}

func saveERC1155ToClickHouse(tx *types.Eth1Transaction, txIndex int, blockNumber uint64, blockHash common.Hash, blockTimestamp int64) error {
	for attempt := 1; attempt <= MAX_RETRY; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// set values for logging
		processedERC1155 := int64(0)
		startTs := time.Now()
		lastTickTs := time.Now()
		var currentProcessed int64

		// prepare ERC1155 batch to send to ClickHouse
		erc1155Batch, err := prepareERC1155Batch(ctx)
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to prepare ERC1155 batch: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to prepare ERC1155 batch after %d attempts: %v", MAX_RETRY, err)
		}

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
				TxHash:      common.HexToHash(tx.Hash),
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
					err = erc1155Batch.Append(
						string(tx.Hash),
						blockNumber,
						transferBatch.From.String(),
						transferBatch.To.String(),
						transferBatch.Operator.String(),
						common.Bytes2Hex(txLog.Address),
						transferBatch.Ids[index],
						transferBatch.Values[index],
						uint64(j),
						topics[0].String(),
						uint64(txIndex),
						txLog.GetRemoved(),
						blockTimestamp,
					)

					if err != nil {
						log.Error(err, "error appending ERC1155 data to batch", 0)
						if attempt < MAX_RETRY {
							log.Warn(fmt.Sprintf("Attempt %d failed to append ERC1155 data to batch. Retrying...", attempt))
							time.Sleep(RETRY_DELAY)
							continue
						}
						return fmt.Errorf("failed to append ERC1155 data to batch after %d attempts: %v", MAX_RETRY, err)
					}

					currentProcessed = atomic.AddInt64(&processedERC1155, 1)
				}
			} else if transferSingle != nil {
				err = erc1155Batch.Append(
					string(tx.Hash),
					blockNumber,
					transferSingle.From.String(),
					transferSingle.To.String(),
					transferSingle.Operator.String(),
					common.Bytes2Hex(txLog.Address),
					transferSingle.Id,
					transferSingle.Value,
					uint64(j),
					topics[0].String(),
					uint64(txIndex),
					txLog.GetRemoved(),
					blockTimestamp,
				)

				if err != nil {
					log.Error(err, "error appending ERC1155 data to batch", 0)
					if attempt < MAX_RETRY {
						log.Warn(fmt.Sprintf("Attempt %d failed to append ERC1155 data to batch. Retrying...", attempt))
						time.Sleep(RETRY_DELAY)
						continue
					}
					return fmt.Errorf("failed to append ERC1155 data to batch after %d attempts: %v", MAX_RETRY, err)
				}

				currentProcessed = atomic.AddInt64(&processedERC1155, 1)
			}
		}

		// send the ERC1155 batch to ClickHouse
		err = erc1155Batch.Send()
		if err != nil {
			if attempt < MAX_RETRY {
				log.Warn(fmt.Sprintf("Attempt %d failed to send ERC1155 batch to ClickHouse: %v. Retrying...", attempt, err))
				time.Sleep(RETRY_DELAY)
				continue
			}
			return fmt.Errorf("failed to send ERC1155 batch to ClickHouse after %d attempts: %v", MAX_RETRY, err)
		}

		if processedERC1155 != 0 {
			log.Infof("processed %v ERC1155 transfers in %v (%.1f txs / sec) for block %d for tx %s", currentProcessed, time.Since(startTs), float64((currentProcessed))/time.Since(lastTickTs).Seconds(), blockNumber, tx.Hash)
		}

		return nil
	}

	return fmt.Errorf("failed to process ERC1155 data for block %d for tx %s after %d attempts", blockNumber, tx.Hash, MAX_RETRY)
}
