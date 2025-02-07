package metadataupdates

import (
	"fmt"
)

const (
	// see https://cloud.google.com/bigtable/docs/using-filters#timestamp-range
	timestampGBTScale = 1000
	// tests showed it's possible to have 36900+ subcalls per tx, but very unlikely - save a bit
	timestampTraceScale = 1 << 15
	// 30m gas / 21.000 gas per transfer = 1428
	timestampTxScale = 1 << 11
	// 64 - (10 bits for timestampGBTScale + TIMESTAMP_TRACE_SCALE + TIMESTAMP_TX_SCALE)
	// = 28 bits left; with a block time of 12s, that's enough for 50+ years
	timestampBlockScale = 1 << (64 - (10 + 15 + 11))
)

// custom timestamp
func encodeIsContractUpdateTs(block_number, tx_idx, trace_idx uint64) (int64, error) {
	var res uint64

	if block_number >= timestampBlockScale {
		return 0, fmt.Errorf("error encoding IsContractTimestamp: block idx is >= %d (block %d, tx %d, trace %d)", timestampBlockScale, block_number, tx_idx, trace_idx)
	}
	res += block_number

	if tx_idx >= timestampTxScale {
		return 0, fmt.Errorf("error encoding IsContractTimestamp: tx idx is >= %d (block %d, tx %d, trace %d)", timestampTxScale, block_number, tx_idx, trace_idx)
	}
	res *= timestampTxScale
	res += tx_idx

	if trace_idx >= timestampTraceScale {
		return 0, fmt.Errorf("error encoding IsContractTimestamp: trace idx is >= %d (block %d, tx %d, trace %d)", timestampTraceScale, block_number, tx_idx, trace_idx)
	}
	res *= timestampTraceScale
	res += trace_idx

	return int64(res * timestampGBTScale), nil
}
