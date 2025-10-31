-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS mobile_app_bundles (
    bundle_version INT NOT NULL PRIMARY KEY, 
    bundle_url TEXT NOT NULL,
    min_native_version INT NOT NULL,
    target_count INT,
    delivered_count INT NOT NULL DEFAULT 0
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS mobile_app_bundles;

-- +goose StatementEnd
