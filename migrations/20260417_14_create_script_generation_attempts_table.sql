-- 20260417_14_create_script_generation_attempts_table.sql
-- Full audit trail of every Writer Agent attempt per content brief.
-- Up to 5 attempts per brief. Each attempt records score, issues found, and the adjusted script.
-- Used by the Writer Agent on attempt N+1 to target exact weaknesses from attempt N.
CREATE TABLE IF NOT EXISTS script_generation_attempts (
    id                  SERIAL PRIMARY KEY,
    content_brief_id    INT           NOT NULL REFERENCES content_briefs(id),
    attempt_number      SMALLINT      NOT NULL CHECK (attempt_number BETWEEN 1 AND 5),
    script_body         TEXT          NOT NULL,
    adjusted_script     TEXT,                     -- Validation Agent's corrected version

    -- Five-dimension scores from Validation Agent
    score_idea_alignment    NUMERIC(5,2) DEFAULT 0,   -- 0-100
    score_angle_match       NUMERIC(5,2) DEFAULT 0,
    score_hook_match        NUMERIC(5,2) DEFAULT 0,
    score_persona_fit       NUMERIC(5,2) DEFAULT 0,
    score_voicedna          NUMERIC(5,2) DEFAULT 0,
    overall_score           NUMERIC(5,2) DEFAULT 0,   -- weighted average

    issues_found        TEXT[],                   -- exact lines and rules violated
    passed_threshold    BOOLEAN       DEFAULT FALSE,  -- TRUE when overall_score >= 90
    selected            BOOLEAN       DEFAULT FALSE,  -- TRUE for the attempt that was used

    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW(),
    UNIQUE (content_brief_id, attempt_number)
);

CREATE INDEX IF NOT EXISTS idx_script_gen_attempts_brief_id ON script_generation_attempts(content_brief_id);
CREATE INDEX IF NOT EXISTS idx_script_gen_attempts_selected ON script_generation_attempts(selected);
