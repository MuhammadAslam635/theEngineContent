-- 20260417_11_create_nick_credentials_table.sql
-- Nick's verified facts, numbers, and credentials that agents are allowed to use in scripts.
-- Agents ONLY draw from this table — they never fabricate statistics.
-- VoiceDNA rule: specific numbers only, real credentials only.
CREATE TABLE IF NOT EXISTS nick_credentials (
    id              SERIAL PRIMARY KEY,
    credential_key  VARCHAR(150)  UNIQUE NOT NULL,  -- machine-readable key e.g. students_enrolled_total
    category        VARCHAR(100)  NOT NULL,          -- e.g. results | experience | social_proof | financial
    display_value   TEXT          NOT NULL,          -- exact phrasing used in scripts e.g. "1,647 students"
    raw_value       NUMERIC,                         -- numeric form for comparison / formatting
    unit            VARCHAR(50),                     -- e.g. students | dollars | years | properties
    context         TEXT,                            -- one-sentence explanation for agent reasoning
    verified_at     TIMESTAMPTZ   DEFAULT NOW(),
    is_active       BOOLEAN       DEFAULT TRUE,
    created_at      TIMESTAMPTZ   DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nick_credentials_category  ON nick_credentials(category);
CREATE INDEX IF NOT EXISTS idx_nick_credentials_is_active ON nick_credentials(is_active);
