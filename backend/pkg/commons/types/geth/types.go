package geth

import (
	"github.com/ethereum/go-ethereum/common"
)

type Trace struct {
	TxHash string
	Result *TraceCall
}

type TraceCall struct {
	TransactionPosition int // todo use something else, that field is not provided by geth, it's manually inserted
	Time                string
	GasUsed             string
	From                common.Address
	To                  common.Address
	Value               string
	Gas                 string
	Input               string
	Output              string
	Error               string
	RevertReason        string // todo have a look at this, it could improve revert message
	Type                string
	Calls               []*TraceCall
}
