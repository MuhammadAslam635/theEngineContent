-- 20260417_07_create_hook_library_table.sql
-- Hook patterns extracted from approved outlier reels.
-- Data-driven only — never predefined. Each row traces back to the reel it was extracted from.
CREATE TABLE IF NOT EXISTS hook_library (
    id                  SERIAL PRIMARY KEY,
    outlier_reel_id     INT           NOT NULL REFERENCES outlier_reels(id),
    hook_text           TEXT          NOT NULL,   -- exact first 3-5 seconds verbatim
    visual_hook         TEXT,                     -- what was shown on screen during opening
    timing_seconds      NUMERIC(5,2), -- seconds before first cut
    pacing              VARCHAR(50),  -- fast | medium | slow
    emotional_trigger   VARCHAR(50),  -- frustration | curiosity | fear | aspiration
    topic_tags          TEXT[],       -- array of keyword tags
    usage_count         INT           DEFAULT 0,  -- how many times this hook has been used
    avg_performance     NUMERIC(10,2) DEFAULT 0,  -- avg views of videos using this hook
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hook_library_outlier_reel_id    ON hook_library(outlier_reel_id);
CREATE INDEX IF NOT EXISTS idx_hook_library_emotional_trigger  ON hook_library(emotional_trigger);
CREATE INDEX IF NOT EXISTS idx_hook_library_avg_performance    ON hook_library(avg_performance DESC);
