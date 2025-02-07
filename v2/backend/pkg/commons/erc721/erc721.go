package erc721

import (
	"github.com/gobitfly/beaconchain/internal/contracts"
)

var abi, _ = contracts.ERC721MetaData.GetAbi()

var TransferTopic = abi.Events["Transfer"].ID
