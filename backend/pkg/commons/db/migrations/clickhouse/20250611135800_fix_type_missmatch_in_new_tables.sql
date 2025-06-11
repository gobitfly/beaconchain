-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS users_val_dashboards_validators;
-- +goose StatementEnd
-- +goose StatementBegin
DROP DICTIONARY IF EXISTS _dict_users_val_dashboards_validators;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE DICTIONARY IF NOT EXISTS _dict_users_val_dashboards_validators
(
    `dashboard_id` Int64,
    `group_id` Int16, -- match existing table
    `validator_index` UInt64 -- match existing table
)
PRIMARY KEY dashboard_id, validator_index
SOURCE(
    POSTGRESQL(
-- +goose ENVSUB ON
        HOST '${POSTGRES_HOST?missing}'
        PORT ${POSTGRES_PORT?missing}
        DB '${POSTGRES_DB?missing}'
        TABLE 'users_val_dashboards_validators'
        USER '${POSTGRES_USER?missing}'
        PASSWORD '${POSTGRES_PASSWORD?missing}'
-- +goose ENVSUB OFF
    )
)
LIFETIME(MIN 5 MAX 15) -- refresh every 5 to 15 seconds
LAYOUT(COMPLEX_KEY_SPARSE_HASHED()) 
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE users_val_dashboards_validators
(
    `dashboard_id` Int64,
    `group_id` Int16, -- match existing table
    `validator_index` UInt64 -- match existing table
)
-- +goose ENVSUB ON
ENGINE = Dictionary(${CLICKHOUSE_DB?missing}._dict_users_val_dashboards_validators);
-- +goose ENVSUB OFF
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
