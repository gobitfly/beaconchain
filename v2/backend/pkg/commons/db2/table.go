package db2

const (
	DataTable     = "data"
	UpdatesTable  = "metadata_updates"
	MetadataTable = "metadata"
	BlocksTable   = "blocks"
)

// Schema is a map containing the bigtable table and the family
var Schema = map[string][]string{
	DataTable: {
		defaultFamily,
	},
	UpdatesTable: {
		defaultFamily,
		updatesBlockFamily,
	},
	MetadataTable: {
		accountFamily,
		erc20MetadataFamily,
	},
	BlocksTable: {
		defaultBlocksFamily,
	},
}

const (
	defaultFamily       = "f"
	defaultBlocksFamily = "default"

	dataColumn = "d"

	updatesBlockFamily = "blocks"
	blockKeysColumn    = "keys"

	blocksDataColumn = "data"
)

const (
	accountFamily           = "a"
	erc20MetadataFamily     = "erc20"
	erc20ColumnPrice        = "PRICE"
	erc20ColumnTotalSupply  = "TOTALSUPPLY"
	accountIsContractColumn = "ISCONTRACT"
)
