-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

SELECT
    'create users_val_dashboards_notifications_history table';

CREATE TABLE
    IF NOT EXISTS users_val_dashboards_notifications_history (
        user_id INT NOT NULL,
        dashboard_id INT NOT NULL,
        group_id INT NOT NULL,
        epoch INT NOT NULL,
        event_type TEXT NOT NULL,
        event_count INT NOT NULL,
        details bytea NOT NULL,
        PRIMARY KEY (
            user_id,
            epoch,
            dashboard_id,
            group_id,
            event_type
        )
    );

/* On the users db */
SELECT
    'create machine_notifications_history table';

CREATE TABLE
    IF NOT EXISTS machine_notifications_history (
        user_id INT NOT NULL,
        epoch INT NOT NULL,
        machine_id BIGINT NOT NULL,
        machine_name TEXT NOT NULL,
        event_type TEXT NOT NULL,
        event_threshold REAL NOT NULL,
        PRIMARY KEY (user_id, epoch, machine_id, event_type)
    );

/* On the users db */
SELECT
    'create client_notifications_history table';

CREATE TABLE
    IF NOT EXISTS client_notifications_history (
        user_id INT NOT NULL,
        epoch INT NOT NULL,
        client TEXT NOT NULL,
        client_version TEXT NOT NULL,
        client_url TEXT NOT NULL,
        PRIMARY KEY (user_id, epoch, client)
    );

/* On the users db */
SELECT
    'create network_notifications_history table';

CREATE TABLE
    IF NOT EXISTS network_notifications_history (
        user_id INT NOT NULL,
        epoch INT NOT NULL,
        network SMALLINT NOT NULL,
        event_type TEXT NOT NULL,
        event_threshold REAL NOT NULL,
        PRIMARY KEY (user_id, epoch, network, event_type)
    );

SELECT
    'create notifications_do_not_disturb_ts column';

ALTER TABLE users
ADD COLUMN IF NOT EXISTS notifications_do_not_disturb_ts TIMESTAMP WITHOUT TIME ZONE;

SELECT
    'create webhook_target column';

ALTER TABLE users_val_dashboards_groups
ADD COLUMN IF NOT EXISTS webhook_target TEXT;

SELECT
    'create discord_webhook_target column';

ALTER TABLE users_val_dashboards_groups
ADD COLUMN IF NOT EXISTS webhook_format TEXT;

SELECT
    'create realtime_notifications column';

ALTER TABLE users_val_dashboards_groups
ADD COLUMN IF NOT EXISTS realtime_notifications BOOL;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

DROP TABLE IF EXISTS users_val_dashboards_notifications_history;

DROP TABLE IF EXISTS machine_notifications_history;

DROP TABLE IF EXISTS client_notifications_history;

DROP TABLE IF EXISTS network_notifications_history;

ALTER TABLE users
DROP COLUMN IF EXISTS notifications_do_not_disturb_ts;

ALTER TABLE users_val_dashboards_groups
DROP COLUMN IF EXISTS webhook_target;

ALTER TABLE users_val_dashboards_groups
DROP COLUMN IF EXISTS webhook_format;

ALTER TABLE users_val_dashboards_groups
DROP COLUMN IF EXISTS realtime_notifications;

-- +goose StatementEnd