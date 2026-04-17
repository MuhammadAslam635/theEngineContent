-- 20260417_16_create_content_videos_table.sql
-- Master record for every video produced by the Content Engine.
-- One video per content brief. Tracks production tool, render status, and final output file.
CREATE TABLE IF NOT EXISTS content_videos (
    id                      SERIAL PRIMARY KEY,
    content_brief_id        INT           NOT NULL UNIQUE REFERENCES content_briefs(id),
    ssml_script_id          INT           NOT NULL REFERENCES ssml_scripts(id),

    -- Production configuration
    production_tool         VARCHAR(30)   NOT NULL DEFAULT 'heygen_avatar_shots',
    -- heygen_avatar_shots | heygen_avatar_v | kling_ai | elevenlabs_studio
    video_type              VARCHAR(20)   NOT NULL DEFAULT 'avatar', -- avatar | faceless
    heygen_job_id           VARCHAR(255),
    heygen_avatar_id        VARCHAR(255),  -- Nick's HeyGen avatar/clone ID

    -- Render status
    render_status           VARCHAR(20)   DEFAULT 'pending',
    -- pending | rendering | coherence_check | stitching | ready | failed

    coherence_check_passed  BOOLEAN,       -- NULL = not yet checked
    coherence_check_notes   TEXT,          -- issues found during scene coherence check

    -- Final output
    video_url               TEXT,          -- URL/path to final stitched video file
    video_duration_seconds  NUMERIC(8,2),
    file_size_bytes         BIGINT,
    resolution              VARCHAR(20),   -- e.g. 1080x1920 (vertical), 1920x1080 (horizontal)
    format                  VARCHAR(10)    DEFAULT 'mp4',

    render_started_at       TIMESTAMPTZ,
    render_completed_at     TIMESTAMPTZ,
    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_videos_content_brief_id ON content_videos(content_brief_id);
CREATE INDEX IF NOT EXISTS idx_content_videos_render_status    ON content_videos(render_status);
CREATE INDEX IF NOT EXISTS idx_content_videos_production_tool  ON content_videos(production_tool);
