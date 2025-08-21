-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS premium_tiers (
    purchase_id VARCHAR(255) NOT NULL,
    tier_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (purchase_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS premium_tiers;

-- +goose StatementEnd
