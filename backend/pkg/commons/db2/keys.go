package db2

// toSuccessor add suffix ";" has it comes after ":" in the ascii order
// this is a simple way to have an infinite bound limit
// prefix must be a real prefix and not a key
func toSuccessor(prefix string) string {
	return prefix + ";"
}

const (
	maxInt                       = 9223372036854775807
	maxExecutionLayerBlockNumber = 1000000000

	txPerBlockLimit = 10_000
	withdrawalLimit = 9999999999999
	logPerTxLimit   = 100_000
	itxPerTxLimit   = 100_000
	maxUncle        = 10
)
