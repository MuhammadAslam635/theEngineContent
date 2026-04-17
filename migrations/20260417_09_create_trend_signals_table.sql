-- 20260417_09_create_trend_signals_table.sql
-- Continuous trend monitoring records per topic per platform.
-- Three time windows (30d / 7d / 24h) determine stage classification and urgency.
CREATE TABLE IF NOT EXISTS trend_signals (
    id                  SERIAL PRIMARY KEY,
    topic               VARCHAR(255)  NOT NULL,
    platform_id         INT           NOT NULL REFERENCES platforms(id),
    volume_30d          INT           DEFAULT 0,  -- post volume last 30 days (baseline)
    volume_7d           INT           DEFAULT 0,  -- post volume last 7 days (momentum)
    volume_24h          INT           DEFAULT 0,  -- post volume last 24 hours (urgency)
    stage               VARCHAR(20)   NOT NULL DEFAULT 'emerging', -- emerging | peaking | saturated | declining
    urgency_rating      SMALLINT      DEFAULT 0,  -- 0-10 computed score
    detected_at         TIMESTAMPTZ   DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,              -- when this signal should be re-evaluated
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trend_signals_platform_id  ON trend_signals(platform_id);
CREATE INDEX IF NOT EXISTS idx_trend_signals_stage        ON trend_signals(stage);
CREATE INDEX IF NOT EXISTS idx_trend_signals_detected_at  ON trend_signals(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_trend_signals_topic        ON trend_signals(topic);
