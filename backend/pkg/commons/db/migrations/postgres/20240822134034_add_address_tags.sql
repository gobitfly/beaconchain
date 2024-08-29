-- +goose Up
-- +goose StatementBegin
SELECT('up SQL query - create address_tags table');
CREATE TABLE address_tags (
    address bytea NOT NULL UNIQUE,
    tag CHARACTER VARYING(100) NOT NULL,
    PRIMARY KEY (address, tag)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT('down SQL query - drop address_tags table');
DROP TABLE execution_payloads;
-- +goose StatementEnd
