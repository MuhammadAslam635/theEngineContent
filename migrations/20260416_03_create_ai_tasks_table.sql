-- 20260416_03_create_ai_tasks_table.sql
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
