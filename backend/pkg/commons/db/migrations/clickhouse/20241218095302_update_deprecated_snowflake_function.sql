-- +goose Up
-- +goose StatementBegin
alter table status_reports modify column `inserted_at` Datetime MATERIALIZED snowflakeIDToDateTime(insert_id::UInt64, 1288834974657)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table status_reports modify column `inserted_at` Datetime MATERIALIZED snowflakeToDateTime(insert_id)
-- +goose StatementEnd
