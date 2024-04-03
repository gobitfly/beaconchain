package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
	"github.com/shopspring/decimal"
)

func WeiToEther(wei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(wei, 0).DivRound(decimal.NewFromInt(params.Ether), 18)
}

func WeiBytesToEther(wei []byte) decimal.Decimal {
	return WeiToEther(new(big.Int).SetBytes(wei))
}

func GWeiToEther(gwei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(gwei, 0).Div(decimal.NewFromInt(params.GWei))
}

func GWeiBytesToEther(gwei []byte) decimal.Decimal {
	return GWeiToEther(new(big.Int).SetBytes(gwei))
}

func GWeiToWei(gwei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(gwei, 0).Mul(decimal.NewFromInt(params.GWei))
}

func GWeiBytesToWei(gwei []byte) decimal.Decimal {
	return GWeiToWei(new(big.Int).SetBytes(gwei))
}
