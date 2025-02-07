package metadataupdates

const Table = "metadata_updates"

var Schema = map[string][]string{
	Table: {
		defaultFamily,
		updatesBlockFamily,
		accountFamily,
	},
}

const (
	defaultFamily      = "f"
	updatesBlockFamily = "blocks"
	accountFamily      = "a"

	accountIsContractColumn = "ISCONTRACT"
	blockKeysColumn         = "keys"
)
