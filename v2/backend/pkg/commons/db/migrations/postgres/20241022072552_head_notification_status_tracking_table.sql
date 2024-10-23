-- +goose Up
-- +goose StatementBegin
SELECT 'creating epochs_notified_head table';
CREATE TABLE IF NOT EXISTS epochs_notified_head (
    epoch INTEGER NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    senton TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    PRIMARY KEY (epoch, event_name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping epochs_notified_head table';
DROP TABLE IF EXISTS epochs_notified_head;
-- +goose StatementEnd
