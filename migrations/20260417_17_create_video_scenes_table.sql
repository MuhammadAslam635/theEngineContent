-- 20260417_17_create_video_scenes_table.sql
-- Individual 15-second scenes built via HeyGen Avatar Shots (SeedDance 2.0).
-- Scenes are stitched into the final video only after coherence check passes.
-- Coherence failures (wardrobe, logo) must be re-rendered before stitching proceeds.
CREATE TABLE IF NOT EXISTS video_scenes (
    id                      SERIAL PRIMARY KEY,
    content_video_id        INT           NOT NULL REFERENCES content_videos(id),
    scene_number            SMALLINT      NOT NULL,         -- 1-based ordering
    scene_script            TEXT          NOT NULL,         -- the portion of script for this scene
    duration_seconds        NUMERIC(5,2)  DEFAULT 15.00,   -- HeyGen Avatar Shots default is 15s

    -- HeyGen scene render details
    heygen_scene_id         VARCHAR(255),
    wardrobe_description    TEXT,          -- description of what Nick is wearing in this scene
    background_description  TEXT,          -- scene background / setting
    logo_present            BOOLEAN,       -- TRUE | FALSE — must be consistent across all scenes
    scene_url               TEXT,          -- URL/path to individual rendered scene file

    -- Coherence check result for this individual scene
    coherence_status        VARCHAR(20)   DEFAULT 'pending',
    -- pending | passed | failed | re_rendering
    coherence_failure_reason TEXT,         -- e.g. "logo missing — other scenes have logo"

    render_status           VARCHAR(20)   DEFAULT 'pending', -- pending | rendering | ready | failed
    render_attempts         SMALLINT      DEFAULT 0,
    created_at              TIMESTAMPTZ   DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   DEFAULT NOW(),
    UNIQUE (content_video_id, scene_number)
);

CREATE INDEX IF NOT EXISTS idx_video_scenes_content_video_id  ON video_scenes(content_video_id);
CREATE INDEX IF NOT EXISTS idx_video_scenes_coherence_status  ON video_scenes(coherence_status);
CREATE INDEX IF NOT EXISTS idx_video_scenes_render_status     ON video_scenes(render_status);
