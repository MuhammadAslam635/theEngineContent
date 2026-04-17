-- 20260417_06_create_outlier_reels_table.sql
-- Raw reel records pulled from SociaVault / YouTube Data API that crossed the 5X threshold.
-- Stores the verbatim transcript and relevance filter outcome.
CREATE TABLE IF NOT EXISTS outlier_reels (
    id                    SERIAL PRIMARY KEY,
    competitor_account_id INT           NOT NULL REFERENCES competitor_accounts(id),
    platform_id           INT           NOT NULL REFERENCES platforms(id),
    external_reel_id      VARCHAR(255)  NOT NULL,          -- native reel/video id from the platform
    title                 TEXT,
    url                   TEXT,
    view_count            BIGINT        NOT NULL DEFAULT 0,
    like_count            BIGINT        DEFAULT 0,
    comment_count         BIGINT        DEFAULT 0,
    share_count           BIGINT        DEFAULT 0,
    account_avg_at_capture BIGINT       NOT NULL DEFAULT 0, -- snapshot of avg when detected
    multiplier            NUMERIC(6,2)  NOT NULL DEFAULT 0, -- view_count / account_avg_at_capture
    raw_transcript        TEXT,                             -- verbatim NLP transcription
    published_at          TIMESTAMPTZ,
    detected_at           TIMESTAMPTZ   DEFAULT NOW(),
    -- Relevance filter result
    relevance_status      VARCHAR(20)   DEFAULT 'pending', -- pending | approved | rejected
    relevance_reason      TEXT,                            -- why rejected (e.g. "references apartment buildings Nick has never bought")
    relevance_checked_at  TIMESTAMPTZ,
    created_at            TIMESTAMPTZ   DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   DEFAULT NOW(),
    UNIQUE (platform_id, external_reel_id)
);

CREATE INDEX IF NOT EXISTS idx_outlier_reels_competitor_account_id ON outlier_reels(competitor_account_id);
CREATE INDEX IF NOT EXISTS idx_outlier_reels_platform_id           ON outlier_reels(platform_id);
CREATE INDEX IF NOT EXISTS idx_outlier_reels_relevance_status      ON outlier_reels(relevance_status);
CREATE INDEX IF NOT EXISTS idx_outlier_reels_detected_at           ON outlier_reels(detected_at DESC);
