-- 20260420_03_create_youtube_channels_table.sql
-- Full YouTube channel profile data fetched from SociaVault API.
-- Linked to competitor_accounts for outlier detection workflows.
CREATE TABLE IF NOT EXISTS youtube_channels (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER       NOT NULL,
    CONSTRAINT fk_youtube_channels_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    competitor_account_id INTEGER,
    CONSTRAINT fk_youtube_channels_competitor FOREIGN KEY (competitor_account_id) REFERENCES competitor_accounts(id) ON DELETE SET NULL,

    -- Core identifiers
    channel_id          VARCHAR(255)  UNIQUE NOT NULL,    -- YouTube channel ID (e.g. UCxcTeAKWJca6XyJ37_ZoKIQ)
    channel_url         TEXT,                             -- Full channel URL
    handle              VARCHAR(255),                     -- @handle
    name                VARCHAR(255)  NOT NULL,           -- Display name

    -- Profile
    avatar_url          TEXT,
    description         TEXT,
    country             VARCHAR(100),
    joined_date         VARCHAR(100),                     -- "Joined Aug 23, 2017"

    -- Stats
    subscriber_count    BIGINT        DEFAULT 0,
    subscriber_count_text VARCHAR(50),                    -- "2.97M subscribers"
    video_count         INTEGER       DEFAULT 0,
    view_count          BIGINT        DEFAULT 0,
    view_count_text     VARCHAR(100),

    -- Metadata
    tags                TEXT,                             -- Comma-separated tags
    email               VARCHAR(255),
    links               JSONB         DEFAULT '[]'::jsonb, -- External links array

    -- Tracking
    last_fetched_at     TIMESTAMPTZ   DEFAULT NOW(),
    is_active           BOOLEAN       DEFAULT TRUE,
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_youtube_channels_user_id ON youtube_channels(user_id);
CREATE INDEX IF NOT EXISTS idx_youtube_channels_channel_id ON youtube_channels(channel_id);
CREATE INDEX IF NOT EXISTS idx_youtube_channels_handle ON youtube_channels(handle);
CREATE INDEX IF NOT EXISTS idx_youtube_channels_is_active ON youtube_channels(is_active);
