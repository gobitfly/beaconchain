package data

import (
	"math/big"
	"time"
)

type DBTransaction struct {
	ChainID            string    `ch:"chain_id"`
	TxHash             string    `ch:"tx_hash"`
	TxIndex            uint64    `ch:"tx_index"`
	BlockNumber        uint64    `ch:"block_number"`
	Timestamp          time.Time `ch:"timestamp"`
	Method             string    `ch:"method"`
	From               string    `ch:"from"`
	To                 string    `ch:"to"`
	Value              string    `ch:"value"`
	TxFee              *big.Int  `ch:"tx_fee"`
	GasPrice           uint64    `ch:"gas_price"`
	IsContractCreation bool      `ch:"is_contract_creation"`
	ErrorMsg           string    `ch:"error_msg"`
	BlobTxFee          *big.Int  `ch:"blob_tx_fee"`
	BlobGasPrice       uint64    `ch:"blob_gas_price"`
	Status             string    `ch:"status"` // todo why string ?
	Type               string    `ch:"type"`
	InsertedAt         time.Time `ch:"inserted_at"`
}
