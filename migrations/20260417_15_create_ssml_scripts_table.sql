-- 20260417_15_create_ssml_scripts_table.sql
-- Output of the SSML Formatting Agent — the final script with ElevenLabs delivery instructions.
-- One SSML record per content brief. Passed directly to ElevenLabs API.
CREATE TABLE IF NOT EXISTS ssml_scripts (
    id                  SERIAL PRIMARY KEY,
    content_brief_id    INT           NOT NULL UNIQUE REFERENCES content_briefs(id),
    approved_script_id  INT           NOT NULL REFERENCES approved_scripts(id),
    ssml_body           TEXT          NOT NULL,   -- full SSML-annotated script
    platform            VARCHAR(20),              -- short_form | long_form affects pacing rules
    elevenlabs_voice_id VARCHAR(255),             -- Nick's voice ID configured in ElevenLabs account
    elevenlabs_audio_url TEXT,                    -- URL/path to generated audio file after submission
    elevenlabs_job_id   VARCHAR(255),             -- async job reference from ElevenLabs API
    audio_status        VARCHAR(20)   DEFAULT 'pending', -- pending | generating | ready | failed
    audio_duration_seconds NUMERIC(8,2),
    submitted_at        TIMESTAMPTZ,
    audio_ready_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ   DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ssml_scripts_content_brief_id ON ssml_scripts(content_brief_id);
CREATE INDEX IF NOT EXISTS idx_ssml_scripts_audio_status     ON ssml_scripts(audio_status);
