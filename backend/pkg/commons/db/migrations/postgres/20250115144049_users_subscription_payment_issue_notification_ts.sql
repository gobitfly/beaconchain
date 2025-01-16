-- +goose Up
-- +goose StatementBegin
ALTER TABLE users_app_subscriptions
ADD COLUMN IF NOT EXISTS payment_issues_mail_ts TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE users_stripe_subscriptions
ADD COLUMN IF NOT EXISTS payment_issues_mail_ts TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users_app_subscriptions
DROP COLUMN IF EXISTS payment_issues_mail_ts;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE users_stripe_subscriptions
DROP COLUMN IF EXISTS payment_issues_mail_ts;
-- +goose StatementEnd