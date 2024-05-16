package modules

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/exp/maps"

	"github.com/gobitfly/beaconchain/pkg/commons/contracts/deposit_contract"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

// if we ever end up in a situation where we possibly have gaps in the data remember: the merkletree_index is unique.
// if one is missing, simple take the blocks from the merkletree_index before and after and fetch the deposits again.
// tho it is sadly stored as a little endian encoded int so you will have to convert it to a number first.
// fixing this would've required messing with v1, better to do when its gone

type executionDepositsExporter struct {
	ModuleContext
	Client                 rpc.Client
	ErigonClient           *gethrpc.Client
	GethClient             *gethrpc.Client
	LogClient              *ethclient.Client
	LogFilterer            *deposit_contract.DepositContractFilterer
	DepositContractAddress common.Address
	LastExportedBlock      uint64
	ExportMutex            *sync.Mutex
	StopEarlyMutex         *sync.Mutex
	StopEarly              context.CancelFunc
	Signer                 gethtypes.Signer
	DepositMethod          abi.Method
}

func NewExecutionDepositsExporter(moduleContext ModuleContext) ModuleInterface {
	return &executionDepositsExporter{
		ModuleContext:          moduleContext,
		Client:                 moduleContext.ConsClient,
		DepositContractAddress: common.HexToAddress(utils.Config.Chain.ClConfig.DepositContractAddress),
		LastExportedBlock:      0,
		ExportMutex:            &sync.Mutex{},
		StopEarlyMutex:         &sync.Mutex{},
	}
}

func (d *executionDepositsExporter) OnHead(event *constypes.StandardEventHeadResponse) (err error) {
	return nil // nop
}

func (d *executionDepositsExporter) Init() error {
	d.Signer = gethtypes.NewCancunSigner(big.NewInt(int64(utils.Config.Chain.ClConfig.DepositChainID)))

	rpcClient, err := gethrpc.Dial(utils.Config.Eth1GethEndpoint)
	if err != nil {
		log.Fatal(err, "new exporter geth client error", 0)
	}
	d.GethClient = rpcClient

	client := ethclient.NewClient(rpcClient)
	d.LogClient = client
	filterer, err := deposit_contract.NewDepositContractFilterer(d.DepositContractAddress, d.LogClient)
	if err != nil {
		return err
	}
	d.LogFilterer = filterer

	erigonClient, err := gethrpc.Dial(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		log.Fatal(err, "new exporter erigon client error", 0)
	}
	d.ErigonClient = erigonClient

	// get deposit method
	parsed, err := deposit_contract.DepositContractMetaData.GetAbi()
	if err != nil {
		return err
	}
	depositMethod, ok := parsed.Methods["deposit"]
	if !ok {
		return fmt.Errorf("error getting deposit-method from deposit-contract-abi")
	}
	d.DepositMethod = depositMethod

	// check if any log_index is missing, if yes we have to do a soft re-export
	// ideally i would check for gaps in the merkletree-index column, but this is extremely annoying as its stored as little endian bytes in the db
	var isV2Check bool
	err = db.WriterDb.Get(&isV2Check, "select count(*) = count(log_index) as is_v2 from eth1_deposits")
	if err != nil {
		return err
	}

	if isV2Check {
		// get latest block from db
		err = db.WriterDb.Get(&d.LastExportedBlock, "select block_number from eth1_deposits order by block_number desc limit 1")
		if err != nil {
			if err == sql.ErrNoRows {
				d.LastExportedBlock = utils.Config.Indexer.ELDepositContractFirstBlock
			} else {
				return err
			}
		}
	} else {
		log.Warnf("log_index is missing in eth1_deposits table, starting from the beginning")
		d.LastExportedBlock = utils.Config.Indexer.ELDepositContractFirstBlock
	}

	log.Infof("initialized execution deposits exporter with last exported block: %v", d.LastExportedBlock)

	// quick kick-start
	go func() {
		err := d.OnFinalizedCheckpoint(nil)
		if err != nil {
			log.Error(err, "error during kick-start", 0)
		}
	}()

	return nil
}

func (d *executionDepositsExporter) GetName() string {
	return "ExecutionDeposits-Exporter"
}

func (d *executionDepositsExporter) OnChainReorg(event *constypes.StandardEventChainReorg) (err error) {
	return nil // nop
}

// can take however long it wants to run, is run in a separate goroutine, so no need to worry about blocking
func (d *executionDepositsExporter) OnFinalizedCheckpoint(event *constypes.StandardFinalizedCheckpointResponse) (err error) {
	// important: have to fetch the actual finalized epoch because even tho its called on finalized checkpoint it actually emits for each justified epoch
	// so we have to do an extra request to get the actual latest finalized epoch
	res, err := d.CL.GetFinalityCheckpoints("finalized")
	if err != nil {
		return err
	}

	var nearestELBlock sql.NullInt64
	err = db.ReaderDb.Get(&nearestELBlock, "select exec_block_number from blocks where slot <= $1 and exec_block_number > 0 order by slot desc limit 1", res.Data.Finalized.Epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return err
	}
	if !nearestELBlock.Valid {
		return fmt.Errorf("no block found for finalized epoch %v", res.Data.Finalized.Epoch)
	}
	log.Debugf("exporting execution layer deposits till block %v", nearestELBlock.Int64)

	err = d.exportTillBlock(uint64(nearestELBlock.Int64))
	if err != nil {
		return err
	}

	return nil
}

// this is basically synchronous, each time it gets called it will kill the previous export and replace it with itself
func (d *executionDepositsExporter) exportTillBlock(block uint64) (err error) {
	// following blocks if a previous function call is still waiting for an export to stop early
	d.StopEarlyMutex.Lock()
	if d.StopEarly != nil {
		// this will run even if the previous export has already finished
		// preventing this would require an overly complex solution
		log.Debugf("asking potentially running export to stop early")
		d.StopEarly()
	}

	// following blocks as long as the running export hasn't finished yet
	d.ExportMutex.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	d.StopEarly = cancel
	// we have over taken and allow potentially newer function calls to signal us to stop early
	d.StopEarlyMutex.Unlock()

	blockOffset := d.LastExportedBlock + 1
	blockTarget := block

	defer d.ExportMutex.Unlock()

	log.Infof("exporting execution layer deposits from %v to %v", blockOffset, blockTarget)

	depositsToSave := make([]*types.ELDeposit, 0)

	for blockOffset < blockTarget {
		tmpBlockTarget := blockOffset + 1000
		if tmpBlockTarget > blockTarget {
			tmpBlockTarget = blockTarget
		}
		log.Debugf("fetching deposits from %v to %v", blockOffset, tmpBlockTarget)
		tmp, err := d.fetchDeposits(blockOffset, tmpBlockTarget)
		if err != nil {
			return err
		}
		depositsToSave = append(depositsToSave, tmp...)
		blockOffset = tmpBlockTarget

		select {
		case <-ctx.Done(): // a newer function call has asked us to stop early
			log.Warnf("stop early signal received, stopping export early")
			blockTarget = tmpBlockTarget
		default:
			continue
		}
	}

	log.Debugf("saving %v deposits", len(depositsToSave))
	err = d.saveDeposits(depositsToSave)
	if err != nil {
		return err
	}

	d.LastExportedBlock = blockTarget

	start := time.Now()
	// update cached view
	err = d.updateCachedView()
	if err != nil {
		return err
	}

	log.Debugf("updating cached deposits view took %v", time.Since(start))

	if len(depositsToSave) > 0 {
		err = d.aggregateDeposits()
		if err != nil {
			return err
		}
	}

	return nil
}

/// --- utils ---

func (d *executionDepositsExporter) fetchDeposits(fromBlock, toBlock uint64) (depositsToSave []*types.ELDeposit, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	depositLogIterator, err := d.LogFilterer.FilterDepositEvent(&bind.FilterOpts{Start: fromBlock, End: &toBlock, Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("error getting logs from execution layer client: %w", err)
	}

	blocksToFetch := make(map[uint64]bool)
	txsToFetch := make(map[string]bool)

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return nil, err
	}

	for depositLogIterator.Next() {
		if depositLogIterator.Event == nil {
			return nil, fmt.Errorf("nil deposit-log")
		}

		depositLog :=
			depositLogIterator.Event
		err = utils.VerifyDepositSignature(&phase0.DepositData{
			PublicKey:             phase0.BLSPubKey(depositLog.Pubkey),
			WithdrawalCredentials: depositLog.WithdrawalCredentials,
			Amount:                phase0.Gwei(binary.LittleEndian.Uint64(depositLog.Amount)),
			Signature:             phase0.BLSSignature(depositLog.Signature),
		}, domain)
		validSignature := err == nil

		blocksToFetch[depositLog.Raw.BlockNumber] = true
		txsToFetch[depositLog.Raw.TxHash.Hex()] = true

		depositsToSave = append(depositsToSave, &types.ELDeposit{
			TxHash:                depositLog.Raw.TxHash.Bytes(),
			TxIndex:               uint64(depositLog.Raw.TxIndex),
			LogIndex:              uint64(depositLog.Raw.Index),
			BlockNumber:           depositLog.Raw.BlockNumber,
			PublicKey:             depositLog.Pubkey,
			WithdrawalCredentials: depositLog.WithdrawalCredentials,
			Amount:                binary.LittleEndian.Uint64(depositLog.Amount),
			Signature:             depositLog.Signature,
			MerkletreeIndex:       depositLog.Index,
			Removed:               depositLog.Raw.Removed,
			ValidSignature:        validSignature,
		})
	}

	if err := depositLogIterator.Error(); err != nil {
		return nil, fmt.Errorf("error iterating over execution layer deposit-logs: %w", err)
	}

	if len(depositsToSave) == 0 {
		return nil, nil
	}

	log.Infof("found %v execution layer deposits between block %v and %v", len(depositsToSave), fromBlock, toBlock)

	headers, txs, err := d.batchRequestHeadersAndTxs(maps.Keys(blocksToFetch), maps.Keys(txsToFetch))
	if err != nil {
		return nil, fmt.Errorf("error getting execution layer blocks: %w\nblocks to fetch: %v\n tx to fetch: %v", err, blocksToFetch, txsToFetch)
	}

	txsToTrace := make(map[string][]int)

	depositAddress := d.DepositContractAddress.Bytes()
	for i, deposit := range depositsToSave {
		// get corresponding block (for the tx-time)
		b, exists := headers[deposit.BlockNumber]
		if !exists {
			return nil, fmt.Errorf("error getting block for execution layer deposit: block does not exist in fetched map")
		}
		deposit.BlockTs = int64(b.Time)

		txHash := fmt.Sprintf("0x%x", deposit.TxHash)
		tx, exists := txs[txHash]
		if !exists {
			return nil, fmt.Errorf("error getting tx for execution layer deposit: tx does not exist in fetched map")
		}
		sender, err := d.Signer.Sender(tx)
		if err != nil {
			return nil, fmt.Errorf("error getting sender for execution layer deposit (txHash: %x): %w", deposit.TxHash, err)
		}
		deposit.FromAddress = sender.Bytes()
		if !bytes.Equal(tx.To().Bytes(), depositAddress) {
			deposit.ToAddress = tx.To().Bytes()
			txsToTrace[txHash] = append(txsToTrace[txHash], i)
		}
	}

	if len(txsToTrace) > 0 {
		if utils.Config.Indexer.DoNotTraceDeposits {
			log.Warnf("deposit tracing is disabled. remove indexer.doNotTraceDeposits to enable. already exported deposits won't be affected by the state of this flag.")
			return depositsToSave, nil
		}
		// trace tnxs to get msg.sender for deposits
		traces, err := d.getDepositTraces(maps.Keys(txsToTrace))
		if err != nil {
			return nil, fmt.Errorf("error getting traces for execution layer deposits: %w", err)
		}

		for txHash, depositIndices := range txsToTrace {
			trace, exists := traces[txHash]
			if !exists {
				return nil, fmt.Errorf("error getting traces for execution layer deposit: tx does not exist in fetched map")
			}
			for _, i := range depositIndices {
				d := depositsToSave[i]
				// find a trace that matches the event
				for ti, t := range trace {
					// if everything matches we can be sure that this is the correct trace
					if bytes.Equal(t.Pubkey, d.PublicKey) &&
						bytes.Equal(t.Signature, d.Signature) &&
						bytes.Equal(t.WithdrawalCredentials, d.WithdrawalCredentials) &&
						t.Value == d.Amount {
						d.MsgSender = t.From
						// remove trace from list
						traces[txHash] = append(traces[txHash][:ti], traces[txHash][ti+1:]...)
						break
					}
				}
				if d.MsgSender == nil {
					return nil, fmt.Errorf("error getting msg.sender for execution layer deposit: no trace matched the deposit")
				}
			}
			if len(traces[txHash]) > 0 {
				// warn if there are still traces left
				log.Warnf("still %v traces left for tx %v, ignoring. @devs consider looking into this if it does occur", len(traces[txHash]), txHash)
			}
		}
	}

	return depositsToSave, nil
}

func (d *executionDepositsExporter) saveDeposits(depositsToSave []*types.ELDeposit) error {
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	insertDepositStmt, err := tx.Prepare(`
		INSERT INTO eth1_deposits (
			tx_hash,
			tx_input,
			tx_index,
			block_number,
			block_ts,
			from_address,
			from_address_text,
			publickey,
			withdrawal_credentials,
			amount,
			signature,
			merkletree_index,
			removed,
			valid_signature,
			msg_sender,
			to_address,
			log_index
		)
		VALUES ($1, '\x00'::bytea, $2, $3, TO_TIMESTAMP($4), $5, ENCODE($6, 'hex'), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (merkletree_index) DO UPDATE SET
			tx_hash                = EXCLUDED.tx_hash,
			tx_input               = EXCLUDED.tx_input,
			tx_index               = EXCLUDED.tx_index,
			block_number           = EXCLUDED.block_number,
			block_ts               = EXCLUDED.block_ts,
			from_address           = EXCLUDED.from_address,
			from_address_text      = EXCLUDED.from_address_text,
			publickey              = EXCLUDED.publickey,
			withdrawal_credentials = EXCLUDED.withdrawal_credentials,
			amount                 = EXCLUDED.amount,
			signature              = EXCLUDED.signature,
			removed                = EXCLUDED.removed,
			valid_signature        = EXCLUDED.valid_signature,
			msg_sender             = EXCLUDED.msg_sender,
			to_address             = EXCLUDED.to_address,
			log_index              = EXCLUDED.log_index`)
	if err != nil {
		return err
	}
	defer insertDepositStmt.Close()

	for _, d := range depositsToSave {
		_, err := insertDepositStmt.Exec(d.TxHash, d.TxIndex, d.BlockNumber, d.BlockTs, d.FromAddress, d.FromAddress, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, d.MerkletreeIndex, d.Removed, d.ValidSignature, d.MsgSender, d.ToAddress, d.LogIndex)
		if err != nil {
			return fmt.Errorf("error saving execution layer deposit to db: %v: %w", fmt.Sprintf("%x", d.TxHash), err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx for execution layer deposits: %w", err)
	}

	return nil
}

func (d *executionDepositsExporter) batchRequestTraces(txsToTrace []string) (map[string]*[]*rpc.ParityTraceResult, error) {
	elems := make([]gethrpc.BatchElem, 0, len(txsToTrace))
	traces := make(map[string]*[]*rpc.ParityTraceResult, len(txsToTrace))
	errors := make([]error, 0, len(txsToTrace))

	for _, txHash := range txsToTrace {
		trace := make([]*rpc.ParityTraceResult, 0)
		err := error(nil)
		elems = append(elems, gethrpc.BatchElem{
			Method: "trace_transaction",
			Args:   []interface{}{txHash},
			Result: &trace,
			Error:  err,
		})
		traces[txHash] = &trace
		errors = append(errors, err)
	}

	lenElems := len(elems)

	if lenElems == 0 {
		return traces, nil
	}

	for i := 0; (i * 32) < lenElems; i++ {
		start := (i * 32)
		end := start + 32

		if end > lenElems {
			end = lenElems
		}
		startTime := time.Now()
		log.Debugf("batch-requesting %v traces (from %v to %v)", len(txsToTrace), start, end)
		ioErr := d.ErigonClient.BatchCall(elems[start:end])
		log.Debugf("batch-requesting %v traces took %v", len(txsToTrace), time.Since(startTime))
		if ioErr != nil {
			return nil, ioErr
		}
	}

	for _, e := range errors {
		if e != nil {
			return nil, e
		}
	}

	return traces, nil
}

type ELDepositTrace struct {
	From                  []byte
	Value                 uint64
	Pubkey                []byte
	WithdrawalCredentials []byte
	Signature             []byte
	DepositDataRoot       [32]byte
}

// batchRequestHeadersAndTxs requests the block range specified in the arguments.
// Instead of requesting each block in one call, it batches all requests into a single rpc call.
// This code is shamelessly stolen and adapted from https://github.com/prysmaticlabs/prysm/blob/2eac24c/beacon-chain/powchain/service.go#L473
func (d *executionDepositsExporter) batchRequestHeadersAndTxs(blocksToFetch []uint64, txsToFetch []string) (map[uint64]*gethtypes.Header, map[string]*gethtypes.Transaction, error) {
	elems := make([]gethrpc.BatchElem, 0, len(blocksToFetch)+len(txsToFetch))
	headers := make(map[uint64]*gethtypes.Header, len(blocksToFetch))
	txs := make(map[string]*gethtypes.Transaction, len(txsToFetch))
	errors := make([]error, 0, len(blocksToFetch)+len(txsToFetch))

	for _, b := range blocksToFetch {
		header := &gethtypes.Header{}
		err := error(nil)
		elems = append(elems, gethrpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(b))), false},
			Result: header,
			Error:  err,
		})
		headers[b] = header
		errors = append(errors, err)
	}

	for _, txHashHex := range txsToFetch {
		tx := &gethtypes.Transaction{}
		err := error(nil)
		elems = append(elems, gethrpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{txHashHex},
			Result: tx,
			Error:  err,
		})
		txs[txHashHex] = tx
		errors = append(errors, err)
	}

	lenElems := len(elems)

	if lenElems == 0 {
		return headers, txs, nil
	}

	for i := 0; (i * 100) < lenElems; i++ {
		start := (i * 100)
		end := start + 100

		if end > lenElems {
			end = lenElems
		}

		ioErr := d.GethClient.BatchCall(elems[start:end])
		if ioErr != nil {
			return nil, nil, ioErr
		}
	}

	for _, e := range errors {
		if e != nil {
			return nil, nil, e
		}
	}

	return headers, txs, nil
}

// automatically filters out unwanted traces (e.g. not deposit calls)
func (d *executionDepositsExporter) getDepositTraces(txsToTrace []string) (filteredTraces map[string][]ELDepositTrace, err error) {
	oneGwei := new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)
	depositFunctionSignature := hexutil.Encode(d.DepositMethod.ID)
	filteredTraces = make(map[string][]ELDepositTrace, 0)

	start := time.Now()
	log.Debugf("tracing %v txs", len(txsToTrace))

	traces, err := d.batchRequestTraces(txsToTrace)
	if err != nil {
		return nil, fmt.Errorf("error getting traces: %w", err)
	}

	for txHash, trace := range traces {
		for _, t := range *trace {
			// ignore failed calls
			if t.Error != "" || t.Type != "call" || t.Action.CallType != "call" || common.HexToAddress(t.Action.To) != d.DepositContractAddress || len(t.Action.Input) < 8+2 || t.Action.Input[:8+2] != depositFunctionSignature {
				continue
			}
			// lets unwrap the input
			data, err := hex.DecodeString(t.Action.Input[8+2:])
			if err != nil {
				return nil, fmt.Errorf("error decoding input: %w", err)
			}

			value, err := hexutil.DecodeBig(t.Action.Value)
			if err != nil {
				return nil, fmt.Errorf("error decoding value: %w", err)
			}

			unpackedData, err := d.DepositMethod.Inputs.Unpack(data)
			if err != nil {
				return nil, fmt.Errorf("error unpacking input: %w", err)
			}

			res := ELDepositTrace{
				From:  hexutil.MustDecode(t.Action.From),
				Value: new(big.Int).Div(value, oneGwei).Uint64(),
			}

			err = d.DepositMethod.Inputs.Copy(&res, unpackedData)
			if err != nil {
				return nil, fmt.Errorf("error copying input: %w", err)
			}

			filteredTraces[txHash] = append(filteredTraces[txHash], res)
		}
	}

	log.Debugf("tracing took %v", time.Since(start))

	return filteredTraces, nil
}

func (d *executionDepositsExporter) aggregateDeposits() error {
	/// this could be a materialized view
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("exporter_aggregate_eth1_deposits").Observe(time.Since(start).Seconds())
	}()
	_, err := db.WriterDb.Exec(`
		INSERT INTO eth1_deposits_aggregated (from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count)
		SELECT
			eth1.from_address,
			SUM(eth1.amount) as amount,
			SUM(eth1.validcount) AS validcount,
			SUM(eth1.invalidcount) AS invalidcount,
			COUNT(CASE WHEN v.status = 'slashed' THEN 1 END) AS slashedcount,
			COUNT(v.pubkey) AS totalcount,
			COUNT(CASE WHEN v.status = 'active_online' OR v.status = 'active_offline' THEN 1 END) as activecount,
			COUNT(CASE WHEN v.status = 'deposited' THEN 1 END) AS pendingcount,
			COUNT(CASE WHEN v.status = 'exited' THEN 1 END) AS voluntary_exit_count
		FROM (
			SELECT
				from_address,
				publickey,
				SUM(amount) AS amount,
				COUNT(CASE WHEN valid_signature = 't' THEN 1 END) AS validcount,
				COUNT(CASE WHEN valid_signature = 'f' THEN 1 END) AS invalidcount
			FROM eth1_deposits
			GROUP BY from_address, publickey
		) eth1
		LEFT JOIN (SELECT pubkey, status FROM validators) v ON v.pubkey = eth1.publickey
		GROUP BY eth1.from_address
		ON CONFLICT (from_address) DO UPDATE SET
			amount               = excluded.amount,
			validcount           = excluded.validcount,
			invalidcount         = excluded.invalidcount,
			slashedcount         = excluded.slashedcount,
			totalcount           = excluded.totalcount,
			activecount          = excluded.activecount,
			pendingcount         = excluded.pendingcount,
			voluntary_exit_count = excluded.voluntary_exit_count`)
	if err != nil && err != sql.ErrNoRows {
		return nil
	}
	return err
}

func (d *executionDepositsExporter) updateCachedView() error {
	err := db.CacheQuery(`
		SELECT
		    uvdv.dashboard_id,
		    uvdv.group_id,
		    ed.block_number,
		    ed.log_index
		FROM
		    eth1_deposits ed
		    INNER JOIN validators v ON ed.publickey = v.pubkey
		    INNER JOIN users_val_dashboards_validators uvdv ON v.validatorindex = uvdv.validator_index
		ORDER BY
		    uvdv.dashboard_id DESC,
		    ed.block_number DESC,
		    ed.log_index DESC;
		`, "cached_eth1_deposits_lookup", []string{"dashboard_id, block_number", "log_index"})
	return err
}
