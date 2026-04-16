-- 20260416_02_create_audit_logs_table.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    agent VARCHAR(255) NOT NULL,
    input_query TEXT,
    agent_prompt TEXT,
    input_tokens INT DEFAULT 0,
    output_response TEXT,
    output_usage_tokens INT DEFAULT 0,
    error TEXT,
    line_number INT,
    user_id INT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_agent ON audit_logs(agent);
