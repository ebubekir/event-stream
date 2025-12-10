-- Create events table for analytics data
CREATE TABLE IF NOT EXISTS events
(
    -- Core identifiers
    id                           String,
    name                         LowCardinality(String),
    channel_type                 LowCardinality(String),
    
    -- Timestamps
    timestamp                    UInt16,
    previous_timestamp           UInt16,
    date                         DateTime,
    
    -- User identifiers
    user_id                      String,
    user_pseudo_id               String,
    
    -- Event Parameters (parallel arrays pattern)
    event_param_keys             Array(String),
    event_param_string_values    Array(String),
    event_param_number_values    Array(Float64),
    event_param_boolean_values   Array(UInt8),
    
    -- User Parameters (parallel arrays pattern)
    user_param_keys              Array(String),
    user_param_string_values     Array(String),
    user_param_number_values     Array(Float64),
    user_param_boolean_values    Array(UInt8),
    
    -- Device info (flattened)
    device_category                  LowCardinality(String),
    device_mobile_brand_name         LowCardinality(String),
    device_mobile_model_name         LowCardinality(String),
    device_operating_system          LowCardinality(String),
    device_operating_system_version  LowCardinality(String),
    device_language                  LowCardinality(String),
    device_browser_name              LowCardinality(String),
    device_browser_version           LowCardinality(String),
    device_hostname                  String,
    
    -- App info (flattened)
    app_info_id                  String,
    app_info_version             LowCardinality(String),
    
    -- Items (parallel arrays pattern)
    item_ids                     Array(String),
    item_names                   Array(String),
    item_brands                  Array(String),
    item_variants                Array(String),
    item_prices_in_usd           Array(Float64),
    item_quantities              Array(Int32),
    item_revenues_in_usd         Array(Float64)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(date)
ORDER BY (date, name, user_pseudo_id, id)
SETTINGS index_granularity = 8192;

-- Create index for faster user lookups
ALTER TABLE events ADD INDEX idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE events ADD INDEX idx_user_pseudo_id user_pseudo_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE events ADD INDEX idx_name name TYPE set(100) GRANULARITY 1;

