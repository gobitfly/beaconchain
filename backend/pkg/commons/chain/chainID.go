package chain

import (
	"math/big"
)

type IDGetter struct {
	Mainnet    *big.Int
	Sepolia    *big.Int
	Gnosis     *big.Int
	Optimistic *big.Int
}

var DefaultIDs = IDGetter{
	Mainnet:    big.NewInt(1),
	Sepolia:    big.NewInt(11155111),
	Gnosis:     big.NewInt(100),
	Optimistic: big.NewInt(10),
}

// IDs is a global variable containing all the chain ids
// it can be modified in test but don't forget to reset it to DefaultIDs
var IDs = DefaultIDs
