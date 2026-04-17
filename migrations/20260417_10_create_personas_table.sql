-- 20260417_10_create_personas_table.sql
-- The four target personas every piece of content is engineered to speak to.
-- Seeded from spec — extensible if new personas are added in future phases.
CREATE TABLE IF NOT EXISTS personas (
    id              SERIAL PRIMARY KEY,
    persona_key     VARCHAR(100)  UNIQUE NOT NULL,  -- snake_case e.g. w2_escapee
    display_name    VARCHAR(150)  NOT NULL,
    who_they_are    TEXT          NOT NULL,
    core_pain       TEXT          NOT NULL,
    income_range    VARCHAR(100),                   -- e.g. $150K+
    language_notes  TEXT,                           -- tone/vocabulary guidance for script agents
    is_active       BOOLEAN       DEFAULT TRUE,
    created_at      TIMESTAMPTZ   DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   DEFAULT NOW()
);

-- Seed the four approved personas from the Content Engine spec
INSERT INTO personas (persona_key, display_name, who_they_are, core_pain, income_range) VALUES
    ('w2_escapee',          'W2 Escapee',            'Employed professional feeling trapped in corporate life',       'Trading time for money with no end in sight',                '$150K+'),
    ('stuck_investor',      'Stuck Investor',         'Has available capital but no proven system to deploy it',       'Money sitting idle losing value to inflation',               NULL),
    ('aspiring_entrepreneur','Aspiring Entrepreneur', 'Has drive but has not started yet',                            'Does not know where to start — paralysed by uncertainty',    NULL),
    ('industry_switcher',   'Industry Switcher',      'Already succeeded elsewhere, wants to pivot',                  'Starting over feels like losing ground already made',        NULL)
ON CONFLICT (persona_key) DO NOTHING;
