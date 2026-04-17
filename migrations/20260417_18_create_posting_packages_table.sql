-- 20260417_18_create_posting_packages_table.sql
-- Complete posting package prepared for each video: formatted file, caption, hashtags,
-- recommended posting time, and a unique tracking identifier.
-- Phase 1: manually reviewed and posted. Phase 2: semi-automated.
CREATE TABLE IF NOT EXISTS posting_packages (
    id                      SERIAL PRIMARY KEY,
    content_video_id        INT           NOT NULL UNIQUE REFERENCES content_videos(id),
    content_brief_id        INT           NOT NULL REFERENCES content_briefs(id),
    tracking_id             VARCHAR(64)   UNIQUE NOT NULL, -- unique id linking video back to all production decisions

    -- Per-platform posting details (one package may target multiple platforms)
    platform_id             INT           NOT NULL REFERENCES platforms(id),
    formatted_video_url     TEXT,          -- platform-specific formatted file
    caption                 TEXT          NOT NULL,
    hook_line               TEXT,          -- first line of caption — mirrors video hook
    hashtags                TEXT[],
    recommended_post_time   TIMESTAMPTZ,   -- peak engagement window for this platform

    -- Review gate (Phase 1 — manual)
    review_status           VARCHAR(20)   DEFAULT 'pending_review',
    -- pending_review | approved | rejected | posted
    reviewed_by_user_id     INT           REFERENCES users(id),
    reviewed_at             TIMESTAMPTZ,
    rejection_reason        TEXT,

    -- Posting
    posted_at               TIMESTAMPTZ,
    platform_post_id        VARCHAR(255),  -- native post id returned by platform after upload

    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posting_packages_content_video_id  ON posting_packages(content_video_id);
CREATE INDEX IF NOT EXISTS idx_posting_packages_platform_id       ON posting_packages(platform_id);
CREATE INDEX IF NOT EXISTS idx_posting_packages_review_status     ON posting_packages(review_status);
CREATE INDEX IF NOT EXISTS idx_posting_packages_tracking_id       ON posting_packages(tracking_id);
CREATE INDEX IF NOT EXISTS idx_posting_packages_posted_at         ON posting_packages(posted_at DESC);
