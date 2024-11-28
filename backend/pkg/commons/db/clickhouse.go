package db

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sync"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var ClickHouseNativeWriter ch.Conn
var ZERO_ADDRESS_STRING = "0x0000000000000000000000000000000000000000"

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
	ctx := context.Background()

	txBatch, err := prepareTxBatch(&ctx)
	itxBatch, err := prepareItxBatch(&ctx)
	erc20Batch, err := prepareERC20Batch(&ctx)
	erc721Batch, err := prepareERC721Batch(&ctx)
	// erc1155Batch, err := prepareERC1155Batch(&ctx)

	for i, tx := range block.Transactions {
		/////////////////////
		//  Transactions ///
		///////////////////
		if slices.Contains(transformerList, "TransformTx") {
			// parse contract address
			to := tx.GetTo()
			isContract := false
			if tx.GetContractAddress() != ZERO_ADDRESS_STRING {
				to = tx.GetContractAddress()
				isContract = true
			}

			// parse method type
			method := make([]byte, 0)
			if len(tx.GetData()) > 3 {
				method = tx.GetData()[:4]
			}

			status := mapStatusToEnum(tx.Status)

			err = txBatch.Append(
				i,
				string(tx.Hash),
				block.Number,
				string(tx.From),
				to,
				fmt.Sprintf("0x%x", tx.Type),
				fmt.Sprintf("%x", string(method)),
				tx.Value,
				tx.Nonce,
				status,
				block.Time.Seconds,
				tx.Gas,
				tx.GasPrice,
				tx.MaxFeePerGas,
				tx.MaxPriorityFeePerGas,
				tx.MaxFeePerBlobGas,
				tx.Gas,
				tx.BlobGasPrice,
				tx.BlobGasUsed,
				tx.AccessList,
				tx.Data,
				isContract,
				tx.Logs,
				tx.LogsBloom,
			)

			if err != nil {
				return fmt.Errorf("error appending tx data to batch: %v", err)
			}
		}

		////////////////////////////
		// Internal Transactions ///
		///////////////////////////
		if slices.Contains(transformerList, "TransformItx") {
			for _, itx := range tx.Itx {
				err = itxBatch.Append(
					string(tx.Hash),
					block.Number,
					itx.From,
					itx.To,
					itx.Type,
					itx.Value,
					itx.Path,
					block.Time.Seconds,
					itx.ErrorMsg,
				)

				if err != nil {
					return fmt.Errorf("error appending itx data to batch: %v", err)
				}
			}
		}

		////////////////////////////
		//   ERC20 Transfers    ///
		///////////////////////////
		if slices.Contains(transformerList, "TransformERC20") {

			for j, log := range tx.GetLogs() {
				// no events emitted continue
				if len(log.GetTopics()) != 3 || !bytes.Equal(log.GetTopics()[0], erc20.TransferTopic) {
					continue
				}

				filterer, err := erc20.NewErc20Filterer(common.Address{}, nil)
				if err != nil {
					return err
				}

				topics := make([]common.Hash, 0, len(log.GetTopics()))

				for _, lTopic := range log.GetTopics() {
					topics = append(topics, common.BytesToHash(lTopic))
				}
				ethLog := gethtypes.Log{
					Address:     common.BytesToAddress(log.GetAddress()),
					Data:        log.Data,
					Topics:      topics,
					BlockNumber: block.Number,
					TxHash:      common.HexToHash(tx.Hash),
					TxIndex:     uint(i),
					BlockHash:   common.BytesToHash(block.GetHash()),
					Index:       uint(j),
					Removed:     log.GetRemoved(),
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
					block.Number,
					transfer.From.String(),
					transfer.To.String(),
					common.BytesToAddress(log.Address).String(),
					value,
					uint(j),
					topics[0].String(),
					uint(i),
					log.GetRemoved(),
					block.Time.Seconds,
				)

				if err != nil {
					return fmt.Errorf("error appending ERC20 data to batch: %v", err)
				}
			}
		}

		////////////////////////////
		//   ERC721 Transfers   ///
		///////////////////////////
		if slices.Contains(transformerList, "TransformERC721") {

			for j, log := range tx.GetLogs() {
				// no events emitted continue
				if len(log.GetTopics()) != 4 || !bytes.Equal(log.GetTopics()[0], erc721.TransferTopic) {
					continue
				}

				filterer, err := erc721.NewErc721Filterer(common.Address{}, nil)
				if err != nil {
					return err
				}

				topics := make([]common.Hash, 0, len(log.GetTopics()))

				for _, lTopic := range log.GetTopics() {
					topics = append(topics, common.BytesToHash(lTopic))
				}

				ethLog := gethtypes.Log{
					Address:     common.BytesToAddress(log.GetAddress()),
					Data:        log.Data,
					Topics:      topics,
					BlockNumber: block.Number,
					TxHash:      common.HexToHash(tx.Hash),
					TxIndex:     uint(i),
					BlockHash:   common.BytesToHash(block.GetHash()),
					Index:       uint(j),
					Removed:     log.GetRemoved(),
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
					block.Number,
					transfer.From.String(),
					transfer.To.String(),
					common.Bytes2Hex(log.Address),
					tokenId,
					uint(j),
					topics[0].String(),
					uint(i),
					log.GetRemoved(),
					block.Time.Seconds,
				)
			}
		}

		////////////////////////////
		//   ERC1155 Transfers  ///
		///////////////////////////
		if slices.Contains(transformerList, "TransformERC1155") {
			//TODO
		}
	}

	// Send the batches to ClickHouse
	if slices.Contains(transformerList, "TransformTx") {
		err = txBatch.Send()
		if err != nil {
			return fmt.Errorf("error while sending tx batch to ClickHouse: %v", err)
		}
	}
	if slices.Contains(transformerList, "TransformItx") {
		err = itxBatch.Send()
		if err != nil {
			return fmt.Errorf("error while sending itx batch to ClickHouse: %v", err)
		}
	}

	if slices.Contains(transformerList, "TransformERC20") {
		err = erc20Batch.Send()
		if err != nil {
			return fmt.Errorf("error while sending ERC20 batch to ClickHouse: %v", err)
		}
	}

	if slices.Contains(transformerList, "TransformERC721") {
		err = erc20Batch.Send()
		if err != nil {
			return fmt.Errorf("error while sending ERC721 batch to ClickHouse: %v", err)
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

func prepareTxBatch(ctx *context.Context) (driver.Batch, error) {
	txBatch, err := ClickHouseNativeWriter.PrepareBatch(*ctx, `
		INSERT INTO transactions_ethereum (
			tx_index, tx_hash, block_number, from_address, to_address, type, method, value, nonce, status,
			timestamp, gas, gas_price, max_fee_per_gas, max_priority_fee_per_gas, max_fee_per_blob_gas, 
			gas_used, blob_gas_price, blob_gas_used, access_list, input_data, is_contract_creation, logs, logs_bloom
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing tx batch for ClickHouse: %v", err)
	}

	return txBatch, nil
}

func prepareItxBatch(ctx *context.Context) (driver.Batch, error) {
	itxBatch, err := ClickHouseNativeWriter.PrepareBatch(*ctx, `
	INSERT INTO internal_tx_ethereum (parent_hash, block_number, from_address, to_address, type, value,
			path, timestamp, error_msg)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing itx batch for ClickHouse: %v", err)
	}

	return itxBatch, nil
}

func prepareERC20Batch(ctx *context.Context) (driver.Batch, error) {
	erc20Batch, err := ClickHouseNativeWriter.PrepareBatch(*ctx, `
	INSERT INTO erc20_ethereum (parent_hash, block_number, from_address, to_address, token_address, value, 
			log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC20 batch for ClickHouse: %v", err)
	}

	return erc20Batch, nil
}

func prepareERC721Batch(ctx *context.Context) (driver.Batch, error) {
	erc721Batch, err := ClickHouseNativeWriter.PrepareBatch(*ctx, `
	INSERT INTO erc721_ethereum (parent_hash, block_number, from_address, to_address, token_address, token_id,  
			log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC721 batch for ClickHouse: %v", err)
	}

	return erc721Batch, nil
}

func prepareERC1155Batch(ctx *context.Context) (driver.Batch, error) {
	erc1155Batch, err := ClickHouseNativeWriter.PrepareBatch(*ctx, `
	INSERT INTO erc1155_ethereum (parent_hash, block_number, from_address, to_address, operator, token_address,   
			token_ids, value, log_index, log_type, transaction_log_index, removed, timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("error while preparing ERC1155 batch for ClickHouse: %v", err)
	}

	return erc1155Batch, nil
}
