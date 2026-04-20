-- 20260420_01_create_global_settings_table.sql
CREATE TABLE IF NOT EXISTS global_settings (
    id SERIAL PRIMARY KEY,
    key_name VARCHAR(255) UNIQUE NOT NULL,
    key_value TEXT NOT NULL,
    key_type VARCHAR(50) DEFAULT 'string',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_global_settings_key_name ON global_settings(key_name);
CREATE INDEX IF NOT EXISTS idx_global_settings_is_active ON global_settings(is_active);