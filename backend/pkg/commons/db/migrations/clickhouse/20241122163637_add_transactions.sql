-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions_ethereum (
    tx_index UInt32 CODEC(T64, ZSTD(3)),
    tx_hash String,
    block_number UInt64 CODEC(T64, ZSTD(3)),           
    from_address String,
    to_address String,
    type LowCardinality(String) CODEC(ZSTD(3)),  
    method String CODEC(ZSTD(3)), 
    value Int128,
    nonce UInt32,
    status Enum('failed' = 0, 'success' = 1, 'partialy failed' = 2) CODEC(ZSTD(3)), 
    timestamp DateTime,
    gas Int64 CODEC(T64, ZSTD(3)),
    gas_price Nullable(Int64) CODEC(ZSTD(3)),
    max_fee_per_gas Nullable(Int64) CODEC(ZSTD(3)),
    max_priority_fee_per_gas Nullable(Int64) CODEC(ZSTD(3)),
    max_fee_per_blob_gas Nullable(Int64) CODEC(ZSTD(3)),
    gas_used Int64 CODEC(T64, ZSTD(3)),
    blob_gas_price Nullable(Int64) CODEC(ZSTD(3)),
    blob_gas_used Nullable(Int64) CODEC(T64, ZSTD(3)),
    access_list Array(Nullable(String)),
    input_data Array(Nullable(UInt8)),
    contract_created Nullable(String),
    logs Array(Nullable(UInt8)),
    logs_bloom Array(Nullable(UInt8)),
    internal_data Nested(
        from_address Nullable(String),
        to_address Nullable(String),
        type Nullable(String),
        value Nullable(String),
        path Nullable(String),
        gas_limit Nullable(Int64),
        error_msg Nullable(String)
    ),
    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_method (method) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (toStartOfWeek(timestamp), from_address, block_number, tx_index)
PRIMARY KEY (toStartOfMonth(timestamp), from_address, block_number, tx_index)
PARTITION BY toStartOfQuarter(timestamp)
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE erc20_ethereum (
    parent_hash String,
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address String,
    to_address String,
    token_address String,
    value Int128,
    log_index UInt32,
    log_type Nullable(String),
    transaction_log_index Nullable(UInt32) CODEC(T64, ZSTD(3)),
    removed Boolean,
    timestamp DateTime,

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY toStartOfMonth(timestamp) 
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE erc721_ethereum(
    parent_hash String,
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address String,
    to_address String,
    token_address String,
    token_id UInt256,
    log_index UInt32,
    log_type Nullable(String),
    transaction_log_index Nullable(UInt32) CODEC(T64, ZSTD(3)),
    removed Boolean,
    timestamp DateTime,

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY toStartOfMonth(timestamp) 
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE erc1155_ethereum(
    parent_hash String,
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address String,
    to_address String,
    operator String,
    token_address String,
    token_ids Array(UInt256),
    value Array(Int128),
    log_index UInt32,
    log_type Nullable(String),
    transaction_log_index Nullable(UInt32) CODEC(T64, ZSTD(3)),
    removed Boolean,
    timestamp DateTime,

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY toStartOfMonth(timestamp) 
SETTINGS index_granularity = 8192
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions_ethereum IF EXISTS
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE erc20_ethereum IF EXISTS
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE erc721_ethereum IF EXISTS
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE erc1155_ethereum IF EXISTS
-- +goose StatementEnd