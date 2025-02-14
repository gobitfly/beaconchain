package metadata

const Table = "metadata"

var Schema = map[string][]string{
	Table: {
		accountFamily,
		erc20MetadataFamily,
	},
}

const (
	accountFamily           = "a"
	erc20MetadataFamily     = "erc20"
	erc20ColumnPrice        = "PRICE"
	erc20ColumnTotalSupply  = "TOTALSUPPLY"
	accountIsContractColumn = "ISCONTRACT"
)
