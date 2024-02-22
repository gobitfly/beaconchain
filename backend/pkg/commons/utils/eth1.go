package utils

import (
	"math/big"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/types"

	"github.com/ethereum/go-ethereum/common"
)

var Erc20TransferEventHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
var Erc1155TransferSingleEventHash = common.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")

func Eth1BlockReward(blockNumber uint64, difficulty []byte) *big.Int {
	// no block rewards for PoS blocks
	// holesky genesis block has difficulty 1 and zero block reward (launched with pos)
	if len(difficulty) == 0 || (len(difficulty) == 1 && difficulty[0] == 1) {
		return big.NewInt(0)
	}

	if blockNumber < Config.Chain.ElConfig.ByzantiumBlock.Uint64() {
		return big.NewInt(5e+18)
	} else if blockNumber < Config.Chain.ElConfig.ConstantinopleBlock.Uint64() {
		return big.NewInt(3e+18)
	} else if Config.Chain.ClConfig.DepositChainID == 5 { // special case for goerli: https://github.com/eth-clients/goerli
		return big.NewInt(0)
	} else {
		return big.NewInt(2e+18)
	}
}

func Eth1TotalReward(block *types.Eth1BlockIndexed) *big.Int {
	blockReward := Eth1BlockReward(block.GetNumber(), block.GetDifficulty())
	uncleReward := big.NewInt(0).SetBytes(block.GetUncleReward())
	txFees := big.NewInt(0).SetBytes(block.GetTxReward())

	totalReward := big.NewInt(0).Add(blockReward, txFees)
	return totalReward.Add(totalReward, uncleReward)
}

func StripPrefix(hexStr string) string {
	return strings.Replace(hexStr, "0x", "", 1)
}

func EthBytesToFloat(b []byte) float64 {
	return WeiBytesToEther(b).InexactFloat64()
}
