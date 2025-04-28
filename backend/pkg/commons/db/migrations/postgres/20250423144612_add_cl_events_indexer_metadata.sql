-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS consensus_layer_events_indexer_metadata (
	epoch int4 NOT NULL,
	transition_block_hash bytea NOT NULL,
	CONSTRAINT consensus_layer_events_indexer_metadata_pk PRIMARY KEY (epoch)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS consensus_layer_events_indexer_metadata;
-- +goose StatementEnd
