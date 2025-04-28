-- +goose Up
-- +goose StatementBegin
ALTER TABLE consensus_layer_events 
    ADD COLUMN IF NOT EXISTS version int4 GENERATED ALWAYS AS (
        COALESCE((data ->>'version')::int4, 0)
    ) STORED;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE consensus_layer_events 
    DROP COLUMN IF EXISTS version;
-- +goose StatementEnd
