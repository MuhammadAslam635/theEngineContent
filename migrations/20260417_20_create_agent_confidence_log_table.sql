-- 20260417_20_create_agent_confidence_log_table.sql
-- Per-decision confidence log for every agent output in the pipeline.
-- Outputs below 70% are flagged for human review. Adversarial audit disagreements escalate.
-- Extends audit_logs (which tracks token usage) with structured confidence and escalation data.
CREATE TABLE IF NOT EXISTS agent_confidence_log (
    id                      SERIAL PRIMARY KEY,
    audit_log_id            INT           REFERENCES audit_logs(id),  -- links to token/prompt audit trail
    content_brief_id        INT           REFERENCES content_briefs(id),
    agent_name              VARCHAR(100)  NOT NULL,
    -- strategic_agent | creative_agent | validation_agent | writer_agent | ssml_agent | adversarial_auditor
    decision_type           VARCHAR(100)  NOT NULL,
    -- e.g. platform_selection | angle_selection | hook_selection | script_validation | ssml_formatting
    confidence_score        NUMERIC(5,2)  NOT NULL,   -- 0-100
    below_threshold         BOOLEAN       GENERATED ALWAYS AS (confidence_score < 70) STORED,
    flagged_for_review      BOOLEAN       DEFAULT FALSE,
    adversarial_agrees      BOOLEAN,      -- NULL = no adversarial check run; TRUE/FALSE = outcome
    escalated               BOOLEAN       DEFAULT FALSE,
    escalation_reason       TEXT,
    retry_count             SMALLINT      DEFAULT 0,   -- 0-2, after 2 escalates to PM
    resolved                BOOLEAN       DEFAULT FALSE,
    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_confidence_log_content_brief_id  ON agent_confidence_log(content_brief_id);
CREATE INDEX IF NOT EXISTS idx_agent_confidence_log_agent_name         ON agent_confidence_log(agent_name);
CREATE INDEX IF NOT EXISTS idx_agent_confidence_log_below_threshold    ON agent_confidence_log(below_threshold);
CREATE INDEX IF NOT EXISTS idx_agent_confidence_log_escalated          ON agent_confidence_log(escalated);
