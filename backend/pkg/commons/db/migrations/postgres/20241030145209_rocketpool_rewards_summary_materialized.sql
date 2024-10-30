-- +goose Up
-- +goose StatementBegin

CREATE MATERIALIZED VIEW IF NOT EXISTS rocketpool_rewards_summary AS
SELECT
    (data->>'index')::INT AS index,
    (data->>'startTime')::TIMESTAMP WITHOUT TIME ZONE AS start_time,  
    (data->>'endTime')::TIMESTAMP WITHOUT TIME ZONE AS end_time,      
    -- Remove the '0x' prefix, pad with '0' if odd-length, and cast as BYTEA
    decode(
        lpad(trim(leading '0x' from node_rewards.key), 
             length(trim(leading '0x' from node_rewards.key)) + 
             (length(trim(leading '0x' from node_rewards.key)) % 2), 
             '0'),
        'hex'
    ) AS node_address,
    (node_rewards.value->>'collateralRpl')::NUMERIC AS collateral_rpl,
    (node_rewards.value->>'smoothingPoolEth')::NUMERIC AS smoothing_pool_eth
FROM
    rocketpool_reward_tree,
    LATERAL jsonb_each(data->'nodeRewards') AS node_rewards
WITH DATA;

CREATE INDEX IF NOT EXISTS idx_node_address ON rocketpool_rewards_summary (node_address);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique ON rocketpool_rewards_summary (index, node_address);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP MATERIALIZED VIEW IF EXISTS rocketpool_rewards_summary;

-- +goose StatementEnd
