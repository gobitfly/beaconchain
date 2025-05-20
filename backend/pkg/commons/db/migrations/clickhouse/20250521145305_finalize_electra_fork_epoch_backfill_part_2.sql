-- +goose Up
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch
    MATERIALIZE COLUMN roi_dividend settings alter_sync=0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
