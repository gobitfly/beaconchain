package chain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type SystemContracts struct {
	ConsolidationQueueAddress common.Address
	WithdrawalQueueAddress    common.Address
}

var officialSystemContracts = SystemContracts{
	ConsolidationQueueAddress: params.ConsolidationQueueAddress,
	WithdrawalQueueAddress:    params.WithdrawalQueueAddress,
}

func SystemContractsFor(chainID *big.Int) SystemContracts {
	// add switch in case of special requirement for some chains
	return officialSystemContracts
}
