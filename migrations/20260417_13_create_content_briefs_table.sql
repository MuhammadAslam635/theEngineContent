-- 20260417_13_create_content_briefs_table.sql
-- Structured output produced by Agent 1 (Strategic) and Agent 2 (Creative).
-- One brief per approved idea arriving from the Idea Engine.
-- Captures every strategic and creative decision with confidence scores and reasoning.
CREATE TABLE IF NOT EXISTS content_briefs (
    id                      SERIAL PRIMARY KEY,
    idea_source             VARCHAR(255),               -- topic / idea description from Idea Engine
    idea_engine_ref         VARCHAR(255),               -- external reference id from Idea Engine if available

    -- Agent 1 — Strategic Decisions
    primary_platform_id     INT           NOT NULL REFERENCES platforms(id),
    content_type            VARCHAR(20)   NOT NULL DEFAULT 'short_form',  -- short_form | long_form
    persona_id              INT           NOT NULL REFERENCES personas(id),
    video_type              VARCHAR(20)   NOT NULL DEFAULT 'avatar',       -- avatar | faceless
    trend_signal_id         INT           REFERENCES trend_signals(id),
    outlier_reel_ref_id     INT           REFERENCES outlier_reels(id),
    agent1_confidence       NUMERIC(5,2)  DEFAULT 0,
    agent1_reasoning        TEXT,

    -- Agent 2 — Creative Decisions
    primary_angle_id        INT           REFERENCES angle_library(id),
    secondary_angle_id      INT           REFERENCES angle_library(id),
    primary_hook_id         INT           REFERENCES hook_library(id),
    agent2_confidence       NUMERIC(5,2)  DEFAULT 0,
    agent2_reasoning        TEXT,
    avoid_when_verified     BOOLEAN       DEFAULT FALSE,  -- confirms avoid_when condition checked

    -- Script selection outcome
    selected_script_id      INT           REFERENCES approved_scripts(id),
    script_selection_method VARCHAR(30)   DEFAULT 'banked',  -- banked | writer_agent
    script_attempts         SMALLINT      DEFAULT 0,
    final_script_score      NUMERIC(5,2)  DEFAULT 0,

    status                  VARCHAR(30)   DEFAULT 'pending',
    -- pending | agent1_complete | agent2_complete | script_selected | production | review | posted | failed

    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_briefs_primary_platform_id ON content_briefs(primary_platform_id);
CREATE INDEX IF NOT EXISTS idx_content_briefs_persona_id          ON content_briefs(persona_id);
CREATE INDEX IF NOT EXISTS idx_content_briefs_status              ON content_briefs(status);
CREATE INDEX IF NOT EXISTS idx_content_briefs_content_type        ON content_briefs(content_type);
CREATE INDEX IF NOT EXISTS idx_content_briefs_created_at          ON content_briefs(created_at DESC);
