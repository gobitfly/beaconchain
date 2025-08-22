-- +goose Up
-- +goose StatementBegin

-- create enum type for tier_name
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_enum') THEN
        CREATE TYPE tier_enum AS ENUM ('FREE', 'HOBBYIST', 'BUSINESS', 'SCALE');
    END IF;
END$$;

-- create enum type for version
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_version_enum') THEN
        CREATE TYPE subscription_version_enum AS ENUM ('stripe_v3', 'stripe_legacy');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS subscription_tiers (
    price_id VARCHAR(255) NOT NULL,
    tier_name tier_enum NOT NULL,
    version subscription_version_enum NOT NULL,
    PRIMARY KEY (price_id)
);

COMMENT ON COLUMN subscription_tiers.price_id IS 'Unique subscription price identifier (e.g. Stripe price ID)';
COMMENT ON COLUMN subscription_tiers.tier_name IS 'Our subscription tier this price belongs to. Valid values: FREE, HOBBYIST, BUSINESS, SCALE';
COMMENT ON COLUMN subscription_tiers.version IS 'Integration version. Valid values: stripe_v3 (current) or stripe_legacy (v1).';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS subscription_tiers;

-- drop enums if no longer used
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_enum') THEN
        DROP TYPE tier_enum;
    END IF;
END$$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_version_enum') THEN
        DROP TYPE subscription_version_enum;
    END IF;
END$$;

-- +goose StatementEnd
