package erc1155

import (
	"github.com/gobitfly/beaconchain/internal/contracts"
)

var abi, _ = contracts.ERC1155MetaData.GetAbi()

var TransferBulkTopic = abi.Events["TransferBulk"].ID
var TransferSingleTopic = abi.Events["TransferSingle"].ID
