package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBDepositsExecution struct {
	Paging   common.Paging        `json:"paging,omitempty"`
	Total    *decimal.Decimal     `json:"total"`
	Deposits []VDBDepositsDetails `json:"deposits"`
}

type VDBDepositsConsensus struct {
	Paging   common.Paging        `json:"paging,omitempty"`
	Total    *decimal.Decimal     `json:"total"`
	Deposits []VDBDepositsDetails `json:"deposits"`
}

type VDBDepositsDetails struct {
	PublicKey             common.PubKey    `json:"public_key"`
	Index                 uint64           `json:"index"`
	GroupName             string           `json:"group_anme,omitempty"`
	GroupId               uint64           `json:"group_id,omitempty"`
	Time                  time.Time        `json:"time"`
	WithdrawalCredentials common.Hash      `json:"withdrawal_credentials"`
	Amount                *decimal.Decimal `json:"amount"`

	// Execution
	Block uint64         `json:"block,omitempty"`
	From  common.Address `json:"from,omitempty"`
	// Depositor      Address   `json:"depositor,omitempty"`
	TxHash common.Hash `json:"tx_hash,omitempty"`
	Valid  bool        `json:"valid,omitempty"`

	// Consensus
	Epoch     uint64      `json:"epoch,omitempty"`
	Slot      uint64      `json:"slot,omitempty"`
	Signature common.Hash `json:"signature,omitempty"`
}
