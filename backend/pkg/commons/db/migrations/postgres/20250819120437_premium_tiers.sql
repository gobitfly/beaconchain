-- +goose Up
-- +goose StatementBegin

-- create enum type for tier_name
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_enum') THEN
        CREATE TYPE tier_enum AS ENUM ('FREE', 'HOBBYIST', 'BUSINESS', 'SCALE');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS subscription_tiers (
    price_id VARCHAR(255) NOT NULL,
    tier_name tier_enum NOT NULL,
    PRIMARY KEY (price_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS subscription_tiers;

-- drop enum type only if it exists (and table no longer depends on it)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_enum') THEN
        DROP TYPE tier_enum;
    END IF;
END$$;

-- +goose StatementEnd
