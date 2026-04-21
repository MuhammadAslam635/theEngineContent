-- 20260420_02_create_agent_settings_table.sql
-- Dynamic agent configuration: prompts with variable placeholders, output type (text/json).
CREATE TABLE IF NOT EXISTS agent_settings (
    id              SERIAL PRIMARY KEY,
    agent_name      VARCHAR(255) NOT NULL,
    supervisor      VARCHAR(255),
    provider        VARCHAR(100) NOT NULL,
    model_name      VARCHAR(255) NOT NULL,
    prompt          TEXT,
    variables       JSONB        DEFAULT '[]'::jsonb,    -- [{key, description, required}]
    output_type     VARCHAR(10)  DEFAULT 'text',         -- 'text' | 'json'
    output_schema   JSONB        DEFAULT NULL,           -- JSON schema when output_type = 'json'
    temperature     REAL         DEFAULT 0.7,
    max_tokens      INT          DEFAULT 2048,
    is_active       BOOLEAN      DEFAULT TRUE,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_settings_agent_name  ON agent_settings(agent_name);
CREATE INDEX IF NOT EXISTS idx_agent_settings_supervisor  ON agent_settings(supervisor);
CREATE INDEX IF NOT EXISTS idx_agent_settings_provider    ON agent_settings(provider);
CREATE INDEX IF NOT EXISTS idx_agent_settings_is_active   ON agent_settings(is_active);
