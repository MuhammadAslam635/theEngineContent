-- 20260417_05_create_competitor_accounts_table.sql
-- Curated list of 20-50 competitor accounts monitored for outlier reel detection.
-- Each account carries its own baseline avg view count so the 5X threshold is account-relative.
CREATE TABLE IF NOT EXISTS competitor_accounts (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER       NOT NULL,
    -- Foreign key to users table
    CONSTRAINT fk_competitor_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    platform_id         INT           NOT NULL REFERENCES platforms(id),
    handle              VARCHAR(255)  NOT NULL,           -- @handle or channel id
    display_name        VARCHAR(255)  NOT NULL,
    profile_url         TEXT,
    niche               VARCHAR(100)  DEFAULT 'business_education', -- e.g. business_education, real_estate
    avg_view_count      BIGINT        DEFAULT 0,          -- rolling average, updated by Scout
    outlier_threshold   BIGINT        DEFAULT 0,          -- computed as avg_view_count * 5
    last_scanned_at     TIMESTAMPTZ,
    is_active           BOOLEAN       DEFAULT TRUE,
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW(),
    UNIQUE (platform_id, handle)
);

CREATE INDEX IF NOT EXISTS idx_competitor_accounts_platform_id ON competitor_accounts(platform_id);
CREATE INDEX IF NOT EXISTS idx_competitor_accounts_is_active   ON competitor_accounts(is_active);
