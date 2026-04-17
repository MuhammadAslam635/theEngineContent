-- =============================================================================
-- FILE: 00_globals.sql
-- PURPOSE: Global types, utility functions, and shared infrastructure
-- PRINCIPLE: Any logic used in more than one place lives here as a function.
--            Tables reference these types; application layers call these funcs.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- EXTENSIONS
-- ---------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";   -- trigram search on script text


-- ---------------------------------------------------------------------------
-- SCHEMA
-- ---------------------------------------------------------------------------
CREATE SCHEMA IF NOT EXISTS content_engine;
SET search_path TO content_engine, public;


-- =============================================================================
-- GLOBAL ENUM TYPES
-- =============================================================================

-- Content format
CREATE TYPE media_type_enum AS ENUM (
    'short_form',
    'long_form'
);

-- Avatar vs background video
CREATE TYPE video_type_enum AS ENUM (
    'avatar',        -- HeyGen Avatar Shots / Avatar V
    'faceless'       -- ElevenLabs audio + Kling/B-roll visuals
);

-- Lifecycle of a video through the pipeline
CREATE TYPE production_status_enum AS ENUM (
    'pending',
    'scripting',
    'voice_generation',
    'video_rendering',
    'coherence_check',
    'packaging',
    'ready_for_review',
    'approved',
    'rejected',
    'posted',
    'failed'
);

-- Trend lifecycle from document Section 3.2
CREATE TYPE trend_stage_enum AS ENUM (
    'emerging',
    'peaking',
    'saturated',
    'declining'
);

-- Analytics interval
CREATE TYPE interval_type_enum AS ENUM (
    '24h',
    '7d',
    '30d'
);

-- Script origin
CREATE TYPE script_type_enum AS ENUM (
    'verbatim_outlier',   -- copied verbatim from competitor reel
    'ai_generated'        -- produced by Script Writer Agent
);

-- Review gate for posting packages
CREATE TYPE review_status_enum AS ENUM (
    'pending',
    'approved',
    'rejected'
);

-- Which AI agent produced a log entry
CREATE TYPE agent_name_enum AS ENUM (
    'strategic_decisions',
    'creative_decisions',
    'script_validation',
    'script_writer',
    'ssml_formatting'
);


-- =============================================================================
-- GLOBAL UTILITY FUNCTIONS
-- Used throughout triggers, constraints, and application stored procedures.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- fn_now()  —  single source of truth for current UTC timestamp
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_now()
RETURNS TIMESTAMPTZ
LANGUAGE sql STABLE PARALLEL SAFE
AS $$
    SELECT NOW() AT TIME ZONE 'UTC';
$$;


-- ---------------------------------------------------------------------------
-- fn_new_id()  —  generate a UUID v4
-- Every table PK uses this; never rely on application-layer UUID generation.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_new_id()
RETURNS UUID
LANGUAGE sql VOLATILE
AS $$
    SELECT uuid_generate_v4();
$$;


-- ---------------------------------------------------------------------------
-- fn_set_updated_at()  —  trigger function that stamps updated_at on any row
-- Attach to any table with an updated_at column via fn_attach_updated_at().
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = fn_now();
    RETURN NEW;
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_attach_updated_at(table_name)
-- Reusable helper: creates the BEFORE UPDATE trigger on any given table.
-- Call once per table that needs updated_at tracking.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_attach_updated_at(p_table TEXT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    EXECUTE FORMAT(
        'CREATE OR REPLACE TRIGGER trg_%s_updated_at
         BEFORE UPDATE ON content_engine.%I
         FOR EACH ROW EXECUTE FUNCTION content_engine.fn_set_updated_at()',
        p_table, p_table
    );
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_require_positive(val, label)
-- Reusable guard: raises if val <= 0. Used in CHECK constraints and procedures.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_require_positive(p_val NUMERIC, p_label TEXT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    IF p_val IS NOT NULL AND p_val <= 0 THEN
        RAISE EXCEPTION '% must be positive, got %', p_label, p_val;
    END IF;
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_require_between(val, lo, hi, label)
-- Reusable range guard used in confidence score constraints.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_require_between(
    p_val   NUMERIC,
    p_lo    NUMERIC,
    p_hi    NUMERIC,
    p_label TEXT
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    IF p_val IS NOT NULL AND (p_val < p_lo OR p_val > p_hi) THEN
        RAISE EXCEPTION '% must be between % and %, got %', p_label, p_lo, p_hi, p_val;
    END IF;
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_score_in_range()  —  trigger guard for 0–100 confidence score columns
-- Apply to any table with a confidence/score column via a trigger.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_score_in_range()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_ARGV[0] IS NOT NULL THEN
        PERFORM content_engine.fn_require_between(
            (row_to_json(NEW)->>TG_ARGV[0])::NUMERIC,
            0, 100,
            TG_ARGV[0]
        );
    END IF;
    RETURN NEW;
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_compute_outlier_threshold(avg_views)
-- Encapsulates the 5× rule from Section 3.1 in one place.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_compute_outlier_threshold(p_avg_views INT)
RETURNS INT
LANGUAGE sql IMMUTABLE PARALLEL SAFE
AS $$
    SELECT GREATEST(0, p_avg_views * 5);
$$;


-- ---------------------------------------------------------------------------
-- fn_classify_trend_stage(vol_30d, vol_7d, vol_24h)
-- Encapsulates the three-window trend classification from Section 3.2.
-- Returns the appropriate trend_stage_enum value.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_classify_trend_stage(
    p_vol_30d INT,
    p_vol_7d  INT,
    p_vol_24h INT
)
RETURNS trend_stage_enum
LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
AS $$
DECLARE
    v_daily_baseline NUMERIC;
    v_weekly_rate    NUMERIC;
BEGIN
    -- Guard: avoid divide-by-zero
    IF p_vol_30d IS NULL OR p_vol_30d = 0 THEN
        RETURN 'declining';
    END IF;

    v_daily_baseline := p_vol_30d::NUMERIC / 30;
    v_weekly_rate    := p_vol_7d::NUMERIC  / 7;

    -- Exploding today → peaking
    IF p_vol_24h >= v_daily_baseline * 5 THEN
        RETURN 'peaking';
    END IF;

    -- Week rate is 2× baseline → emerging
    IF v_weekly_rate >= v_daily_baseline * 2 THEN
        RETURN 'emerging';
    END IF;

    -- Week rate is below baseline → saturated or declining
    IF v_weekly_rate < v_daily_baseline * 0.75 THEN
        IF p_vol_24h < v_daily_baseline * 0.5 THEN
            RETURN 'declining';
        END IF;
        RETURN 'saturated';
    END IF;

    RETURN 'saturated';
END;
$$;


-- ---------------------------------------------------------------------------
-- fn_script_beats_baseline(view_count, baseline)
-- Single place to evaluate Phase 1 success metric: beat 38K baseline.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_script_beats_baseline(
    p_view_count INT,
    p_baseline   INT DEFAULT 38000
)
RETURNS BOOLEAN
LANGUAGE sql IMMUTABLE PARALLEL SAFE
AS $$
    SELECT COALESCE(p_view_count, 0) > p_baseline;
$$;


-- ---------------------------------------------------------------------------
-- fn_derive_outlier_ratio(view_count, avg_views)
-- Computes the ratio that determines if a reel is an outlier (>= 5.0).
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_derive_outlier_ratio(
    p_view_count INT,
    p_avg_views  INT
)
RETURNS NUMERIC
LANGUAGE sql IMMUTABLE PARALLEL SAFE
AS $$
    SELECT CASE
        WHEN COALESCE(p_avg_views, 0) = 0 THEN 0
        ELSE ROUND(p_view_count::NUMERIC / p_avg_views, 2)
    END;
$$;
