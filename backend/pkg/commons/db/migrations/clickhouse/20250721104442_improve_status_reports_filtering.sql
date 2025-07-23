-- +goose Up
-- +goose StatementBegin
ALTER TABLE status_reports ADD INDEX event_id_index event_id Type set(32) settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE status_reports MATERIALIZE INDEX event_id_index;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE gnosis.status_reports ADD INDEX expires_at_index expires_at Type minmax() settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE gnosis.status_reports MATERIALIZE INDEX expires_at_index;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
