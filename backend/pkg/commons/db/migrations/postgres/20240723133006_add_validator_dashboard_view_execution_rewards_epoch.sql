-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_view_execution_rewards_epoch
AS SELECT sum(COALESCE(rb.value, ep.fee_recipient_reward * 1e18::numeric, 0::numeric)) AS value,
    b.proposer,
    b.epoch AS epoch
   FROM blocks b
     LEFT JOIN execution_payloads ep ON ep.block_hash = b.exec_block_hash
     LEFT JOIN relays_blocks rb ON rb.exec_block_hash = b.exec_block_hash
  WHERE b.status = '1'::text
  GROUP BY b.proposer, b.epoch;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS validator_dashboard_view_execution_rewards_epoch;

-- +goose StatementEnd
