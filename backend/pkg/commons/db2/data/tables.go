package data

const Table = "data"

var Schema = map[string][]string{
	Table: {
		defaultFamily,
	},
}

const (
	defaultFamily = "f"
	dataColumn    = "d"
)
