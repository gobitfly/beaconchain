package clickhouse

const Transaction = `
CREATE TABLE IF NOT EXISTS transactions(
    chain_id LowCardinality(String) CODEC(ZSTD(3)),
    tx_hash FixedString(32) CODEC(NONE), 
    tx_index UInt64 CODEC(T64, ZSTD(3)),
    block_number UInt64 CODEC(T64, ZSTD(3)),           
    timestamp DateTime, 
    method String CODEC(ZSTD(3)), 
    from FixedString(40) CODEC(ZSTD(3)), 
    to FixedString(40) CODEC(ZSTD(3)),
    value String, 
    tx_fee UInt256 CODEC(ZSTD(3)), 
    gas_price Nullable(UInt64) CODEC(ZSTD(3)), 
    is_contract_creation Boolean,
    error_msg String, 
    blob_tx_fee Nullable(UInt256) CODEC(ZSTD(3)), 
    blob_gas_price Nullable(UInt64) CODEC(ZSTD(3)), 
    status Enum('FAILED' = 0, 'SUCCESS' = 1, 'PARTIAL' = 2) CODEC(ZSTD(3)), 
    type LowCardinality(String) CODEC(ZSTD(3)),

    inserted_at DateTime MATERIALIZED now(), 
    
    INDEX idx_from_to (from, to) TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_to_from (to, from) TYPE bloom_filter(0.1) GRANULARITY 8,
    INDEX idx_from from TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_to to TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_method method TYPE bloom_filter(0.1) GRANULARITY 8,
    INDEX idx_tx_hash tx_hash TYPE bloom_filter(0.1) GRANULARITY 8,
    INDEX idx_from_to_tx_hash (from, to, tx_hash) TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_to_from_tx_hash (to, from, tx_hash) TYPE bloom_filter(0.1) GRANULARITY 8,
    INDEX idx_type type TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_from_to_type (from, to, type) TYPE bloom_filter(0.1) GRANULARITY 1,
    INDEX idx_to_from_type (to, from, type) TYPE bloom_filter(0.1) GRANULARITY 8
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY (toStartOfQuarter(timestamp), chain_id)
ORDER BY (timestamp, block_number, tx_index)
SETTINGS index_granularity = 8192, deduplicate_merge_projection_mode = 'rebuild'
`
