package data

const Table = "data"

// Schema is a map containing the bigtable table and the family
var Schema = map[string][]string{
	Table: {
		defaultFamily,
	},
}

const (
	defaultFamily = "f"
	dataColumn    = "d"
)
