-- Initial schema for AI Orchestration

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    utype VARCHAR(255) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP,
    verify_otp VARCHAR(255),
    verify_otp_expires TIMESTAMP,
    last_login TIMESTAMP,
    last_login_ip VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    task_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payload JSONB,
    result JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_tasks_user_id ON ai_tasks(user_id);

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
    ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_agent ON audit_logs(agent);
