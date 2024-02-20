package apitypes

type Sort[T ~int] struct {
	Column T
	Desc   bool
}

// TODO delet this and load from config

type Network uint64

const (
	Ethereum Network = 1
	Holesky  Network = 17000
	Sepolia  Network = 11155111

	ArbitrumOneEthereum  Network = 42161
	ArbitrumOneSepolia   Network = 421614
	ArbitrumNovaEthereum Network = 42170

	OptimismEthereum Network = 10
	OptimismSepolia  Network = 11155420

	BaseEthereum Network = 8453
	BaseSepolia  Network = 84532

	Gnosis Network = 100
	Chiado Network = 10200
)
