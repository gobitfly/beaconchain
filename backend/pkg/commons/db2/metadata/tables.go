package metadata

const Table = "metadata"

var Schema = map[string][]string{
	Table: {
		accountFamily,
	},
}

const (
	accountFamily = "a"
)
