-- +goose Up

-- users val dashboards validators table

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_val_dashboards_validators (a Int64) ENGINE = Null() -- for setups where the table wasnt manually created yet
-- +goose StatementEnd
-- +goose StatementBegin
CREATE DICTIONARY IF NOT EXISTS _dict_users_val_dashboards_validators
(
    `dashboard_id` Int64,
    `group_id` Int64,
    `validator_index` Int64
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
DROP TABLE IF EXISTS _tmp_users_val_dashboards_validators; -- for safety, because we dont use IF NOT EXISTS below
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE _tmp_users_val_dashboards_validators
(
    `dashboard_id` Int64,
    `group_id` Int64,
    `validator_index` Int64
)
-- +goose ENVSUB ON
ENGINE = Dictionary(${CLICKHOUSE_DB?missing}._dict_users_val_dashboards_validators);
-- +goose ENVSUB OFF
-- +goose StatementEnd
-- +goose StatementBegin
EXCHANGE TABLES _tmp_users_val_dashboards_validators AND users_val_dashboards_validators;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _tmp_users_val_dashboards_validators; -- drop old table
-- +goose StatementEnd

-- users val dashboards groups table

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_val_dashboards_groups (a Int64) ENGINE = Null() -- for setups where the table wasnt manually created yet
-- +goose StatementEnd
-- +goose StatementBegin
CREATE DICTIONARY IF NOT EXISTS _dict_users_val_dashboards_groups
(
    `id` Int64,
    `dashboard_id` Int64,
    `name` String,
)
PRIMARY KEY id, dashboard_id
SOURCE(
    POSTGRESQL(
-- +goose ENVSUB ON
        HOST '${POSTGRES_HOST?missing}'
        PORT ${POSTGRES_PORT?missing}
        DB '${POSTGRES_DB?missing}'
        TABLE 'users_val_dashboards_groups'
        USER '${POSTGRES_USER?missing}'
        PASSWORD '${POSTGRES_PASSWORD?missing}'
-- +goose ENVSUB OFF
    )
)
LIFETIME(MIN 5 MAX 15) -- refresh every 5 to 15 seconds
LAYOUT(COMPLEX_KEY_SPARSE_HASHED()) 
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _tmp_users_val_dashboards_groups; -- for safety, because we dont use IF NOT EXISTS below
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE _tmp_users_val_dashboards_groups
(
    `id` Int64,
    `dashboard_id` Int64,
    `name` String
)
-- +goose ENVSUB ON
ENGINE = Dictionary(${CLICKHOUSE_DB?missing}._dict_users_val_dashboards_groups);
-- +goose ENVSUB OFF
-- +goose StatementEnd
-- +goose StatementBegin
EXCHANGE TABLES _tmp_users_val_dashboards_groups AND users_val_dashboards_groups;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _tmp_users_val_dashboards_groups; -- drop old table
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
select 'no rollback needed/supported for this migration' as message
-- +goose StatementEnd
