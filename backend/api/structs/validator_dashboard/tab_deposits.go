package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBDepositsExecution struct {
	Total    decimal.Decimal        `json:"total"`
	Deposits []VDBDepositsDetailsEl `json:"deposits"`
}

type VDBDepositsConsensus struct {
	Total    decimal.Decimal        `json:"total"`
	Deposits []VDBDepositsDetailsCl `json:"deposits"`
}

type VDBDepositsDetailsEl struct {
	PublicKey             common.PubKey   `json:"public_key"`
	Index                 uint64          `json:"index"`
	GroupName             string          `json:"group_anme"`
	GroupId               uint64          `json:"group_id"`
	Time                  time.Time       `json:"time"`
	WithdrawalCredentials common.Hash     `json:"withdrawal_credentials"`
	Amount                decimal.Decimal `json:"amount"`

	Block uint64         `json:"block"`
	From  common.Address `json:"from"`
	// Depositor      Address   `json:"depositor"`
	TxHash common.Hash `json:"tx_hash"`
	Valid  bool        `json:"valid"`
}

type VDBDepositsDetailsCl struct {
	PublicKey             common.PubKey   `json:"public_key"`
	Index                 uint64          `json:"index"`
	GroupName             string          `json:"group_anme"`
	GroupId               uint64          `json:"group_id"`
	Time                  time.Time       `json:"time"`
	WithdrawalCredentials common.Hash     `json:"withdrawal_credentials"`
	Amount                decimal.Decimal `json:"amount"`

	Epoch     uint64      `json:"epoch"`
	Slot      uint64      `json:"slot"`
	Signature common.Hash `json:"signature"`
}
