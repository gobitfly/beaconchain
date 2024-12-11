-- +goose Up
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch MODIFY SETTING ttl_only_drop_parts = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch MODIFY TTL _inserted_at + INTERVAL 1 WEEK DELETE SETTINGS materialize_ttl_after_modify=1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch MODIFY SETTING ttl_only_drop_parts = 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch REMOVE TTL SETTINGS materialize_ttl_after_modify=1;
-- +goose StatementEnd
