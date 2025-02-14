package metadataupdates

const Table = "metadata_updates"

var Schema = map[string][]string{
	Table: {
		defaultFamily,
		updatesBlockFamily,
	},
}

const (
	defaultFamily      = "f"
	updatesBlockFamily = "blocks"

	blockKeysColumn = "keys"
)
