-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions(
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    tx_index UInt64 CODEC(T64, ZSTD(3)),
    tx_hash FixedString(32) CODEC(NONE), 
    block_number UInt64 CODEC(T64, ZSTD(3)),           
    from_address FixedString(20) CODEC(ZSTD(3)), 
    to_address Nullable(FixedString(20)) CODEC(ZSTD(3)),
    type LowCardinality(String) CODEC(ZSTD(3)),  
    method String CODEC(ZSTD(3)), 
    value UInt64, 
    nonce UInt64 CODEC(T64, ZSTD(3)), 
    status Enum('failed' = 0, 'success' = 1, 'partialy failed' = 2) CODEC(ZSTD(3)), 
    timestamp DateTime, 
    tx_fee UInt256 CODEC(ZSTD(3)), 
    gas UInt64 CODEC(ZSTD(3)),
    gas_price Nullable(UInt64) CODEC(ZSTD(3)), 
    gas_used UInt64 CODEC(ZSTD(3)),
    max_fee_per_gas Nullable(UInt64) CODEC(ZSTD(3)), 
    max_priority_fee_per_gas Nullable(UInt64) CODEC(ZSTD(3)), 
    -- Blob Tx
    max_fee_per_blob_gas Nullable(UInt64) CODEC(ZSTD(3)), 
    blob_gas_price Nullable(UInt64) CODEC(ZSTD(3)), 
    blob_gas_used Nullable(UInt64) CODEC(T64, ZSTD(3)), 
    blob_tx_fee Nullable(UInt256) CODEC(ZSTD(3)), 
    blob_versioned_hashes Array(Nullable(String)) CODEC(ZSTD(3)), 
    access_list Array(Nullable(String)) CODEC(ZSTD(3)), 
    input_data Array(Nullable(UInt8)) CODEC(ZSTD(3)), 
    is_contract_creation Boolean,
    logs Array(Nullable(String)) CODEC(ZSTD(3)), 
    logs_bloom Array(Nullable(UInt8)) CODEC(ZSTD(3)), 
    inserted_timestamp DateTime DEFAULT now(), 

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_method (method) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = ReplacingMergeTree(inserted_timestamp)
ORDER BY (toStartOfWeek(timestamp), from_address, block_number, tx_index)
PRIMARY KEY (toStartOfWeek(timestamp), from_address, block_number, tx_index)
PARTITION BY (chain_id, toStartOfQuarter(timestamp))
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS internal_transactions(
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    parent_hash FixedString(32) CODEC(ZSTD(3)),
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address FixedString(20) CODEC(ZSTD(3)),
    to_address FixedString(20) CODEC(ZSTD(3)),
    type String CODEC(ZSTD(3)),
    value String CODEC(ZSTD(3)),
    path String,
    gas UInt64 CODEC(ZSTD(3)),
    timestamp DateTime,
    error_msg Nullable(String) CODEC(ZSTD(3)),
    inserted_timestamp DateTime DEFAULT now(),

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1
    
)ENGINE = ReplacingMergeTree(inserted_timestamp)
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, value)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, value)
PARTITION BY (chain_id, toStartOfQuarter(timestamp))
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS erc20_transfers (
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    parent_hash FixedString(32) CODEC(ZSTD(3)),
    block_number UInt64 CODEC(T64, ZSTD(3)), 
    from_address FixedString(20) CODEC(ZSTD(3)),
    to_address FixedString(20) CODEC(ZSTD(3)), 
    token_address FixedString(20) CODEC(ZSTD(3)), 
    value UInt256 CODEC(ZSTD(3)),
    log_index UInt32 CODEC(T64, ZSTD(3)),
    log_type FixedString(32) CODEC(ZSTD(3)),
    transaction_log_index UInt32 CODEC(ZSTD(3)),
    removed Boolean,
    timestamp DateTime,
    inserted_timestamp DateTime DEFAULT now(),

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = ReplacingMergeTree(inserted_timestamp)
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY (chain_id, toStartOfQuarter(timestamp))
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS erc721_transfers(
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    parent_hash FixedString(32) CODEC(ZSTD(3)),
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address FixedString(20) CODEC(ZSTD(3)),
    to_address FixedString(20) CODEC(ZSTD(3)),
    token_address FixedString(20) CODEC(ZSTD(3)),
    token_id UInt256 CODEC(ZSTD(3)),
    log_index UInt32 CODEC(ZSTD(3)),
    log_type FixedString(32) CODEC(ZSTD(3)),
    transaction_log_index Nullable(UInt32) CODEC(ZSTD(3)),
    removed Boolean,
    timestamp DateTime,
    inserted_timestamp DateTime DEFAULT now(),

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = ReplacingMergeTree(inserted_timestamp)
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY (chain_id, toStartOfQuarter(timestamp))
SETTINGS index_granularity = 8192
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS erc1155_transfers(
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    parent_hash FixedString(32) CODEC(ZSTD(3)),
    block_number UInt64 CODEC(T64, ZSTD(3)),
    from_address FixedString(20) CODEC(ZSTD(3)),
    to_address FixedString(20) CODEC(ZSTD(3)),
    operator FixedString(20) CODEC(ZSTD(3)),
    token_address FixedString(20) CODEC(ZSTD(3)),
    token_id UInt256 CODEC(ZSTD(3)),
    value UInt256 CODEC(ZSTD(3)),
    log_index UInt32 CODEC(ZSTD(3)),
    log_type FixedString(32) CODEC(ZSTD(3)),
    transaction_log_index Nullable(UInt32) CODEC(ZSTD(3)),
    removed Boolean,
    timestamp DateTime,
    inserted_timestamp DateTime DEFAULT now(),

    INDEX idx_to_address (to_address) TYPE bloom_filter(0.5) GRANULARITY 1,
    INDEX idx_token_address (token_address) TYPE bloom_filter(0.5) GRANULARITY 1
) ENGINE = ReplacingMergeTree(inserted_timestamp)
ORDER BY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PRIMARY KEY (toStartOfWeek(timestamp), parent_hash, from_address, block_number, log_index)
PARTITION BY (chain_id, toStartOfQuarter(timestamp))
SETTINGS index_granularity = 8192
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS internal_transactions 
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS erc20_transfers 
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS erc721_transfers
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS erc1155_transfers
-- +goose StatementEnd