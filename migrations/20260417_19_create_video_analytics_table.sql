-- 20260417_19_create_video_analytics_table.sql
-- Performance data collected at 24h, 7d, and 30d intervals via official platform APIs.
-- Primary goal: does this video beat Nick's 38K average baseline?
-- Data feeds back into agent scoring and script database promotions.
CREATE TABLE IF NOT EXISTS video_analytics (
    id                      SERIAL PRIMARY KEY,
    posting_package_id      INT           NOT NULL REFERENCES posting_packages(id),
    platform_id             INT           NOT NULL REFERENCES platforms(id),
    tracking_id             VARCHAR(64)   NOT NULL, -- mirrors posting_packages.tracking_id for fast joins

    -- Collection window
    window                  VARCHAR(10)   NOT NULL, -- 24h | 7d | 30d
    collected_at            TIMESTAMPTZ   DEFAULT NOW(),

    -- Core metrics
    view_count              BIGINT        DEFAULT 0,
    like_count              BIGINT        DEFAULT 0,
    comment_count           BIGINT        DEFAULT 0,
    share_count             BIGINT        DEFAULT 0,
    save_count              BIGINT        DEFAULT 0,
    sends_per_reach         NUMERIC(8,4)  DEFAULT 0,  -- sends / reach — key virality signal
    completion_rate         NUMERIC(5,2)  DEFAULT 0,  -- % who watched to end
    engagement_rate         NUMERIC(5,2)  DEFAULT 0,  -- (likes+comments+shares) / views * 100

    -- Baseline comparison
    beats_baseline          BOOLEAN,       -- TRUE if view_count > 38000 (Nick's 38K target)
    baseline_multiplier     NUMERIC(6,2),  -- view_count / 38000

    -- Phase 2 triggers (evaluated post-collection)
    triggers_longform       BOOLEAN       DEFAULT FALSE, -- TRUE if 7d views >= 50000

    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    UNIQUE (posting_package_id, window)
);

CREATE INDEX IF NOT EXISTS idx_video_analytics_posting_package_id ON video_analytics(posting_package_id);
CREATE INDEX IF NOT EXISTS idx_video_analytics_platform_id        ON video_analytics(platform_id);
CREATE INDEX IF NOT EXISTS idx_video_analytics_tracking_id        ON video_analytics(tracking_id);
CREATE INDEX IF NOT EXISTS idx_video_analytics_window             ON video_analytics(window);
CREATE INDEX IF NOT EXISTS idx_video_analytics_collected_at       ON video_analytics(collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_analytics_beats_baseline     ON video_analytics(beats_baseline);
