-- 20260417_12_create_approved_scripts_table.sql
-- The storehouse of scripts ready to be rendered every day.
-- Sources: verbatim outlier copies (relevance-filtered) OR high-performing Writer Agent outputs.
-- The database must never run empty — short-form and long-form always banked.
CREATE TABLE IF NOT EXISTS approved_scripts (
    id                  SERIAL PRIMARY KEY,
    outlier_reel_id     INT           REFERENCES outlier_reels(id),    -- set if sourced from an outlier reel
    title               VARCHAR(255)  NOT NULL,
    script_body         TEXT          NOT NULL,
    content_type        VARCHAR(20)   NOT NULL DEFAULT 'short_form',   -- short_form | long_form
    source              VARCHAR(30)   NOT NULL DEFAULT 'outlier_copy', -- outlier_copy | writer_agent | manual
    persona_id          INT           REFERENCES personas(id),
    angle_id            INT           REFERENCES angle_library(id),
    hook_id             INT           REFERENCES hook_library(id),
    topic_tags          TEXT[],
    word_count          INT           DEFAULT 0,
    estimated_duration_seconds INT   DEFAULT 0,
    voicedna_score      NUMERIC(5,2)  DEFAULT 0,   -- 0-100, must be >= 90 to qualify
    validation_score    NUMERIC(5,2)  DEFAULT 0,   -- overall confidence score from Validation Agent
    validation_notes    TEXT,
    times_used          INT           DEFAULT 0,
    last_used_at        TIMESTAMPTZ,
    avg_views_when_used NUMERIC(10,2) DEFAULT 0,
    is_active           BOOLEAN       DEFAULT TRUE,
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_approved_scripts_content_type     ON approved_scripts(content_type);
CREATE INDEX IF NOT EXISTS idx_approved_scripts_source           ON approved_scripts(source);
CREATE INDEX IF NOT EXISTS idx_approved_scripts_persona_id       ON approved_scripts(persona_id);
CREATE INDEX IF NOT EXISTS idx_approved_scripts_angle_id         ON approved_scripts(angle_id);
CREATE INDEX IF NOT EXISTS idx_approved_scripts_validation_score ON approved_scripts(validation_score DESC);
CREATE INDEX IF NOT EXISTS idx_approved_scripts_is_active        ON approved_scripts(is_active);
