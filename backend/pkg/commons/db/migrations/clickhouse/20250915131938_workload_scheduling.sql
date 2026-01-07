-- +goose Up
-- +goose StatementBegin
CREATE RESOURCE IF NOT EXISTS query (query);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE WORKLOAD IF NOT EXISTS exporter_rollings SETTINGS max_concurrent_queries = 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP WORKLOAD IF EXISTS exporter_rollings;
-- +goose StatementEnd
-- +goose StatementBegin
DROP RESOURCE IF EXISTS query;
-- +goose StatementEnd
