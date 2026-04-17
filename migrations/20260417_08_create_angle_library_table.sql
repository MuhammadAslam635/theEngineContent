-- 20260417_08_create_angle_library_table.sql
-- Angles extracted from approved outlier reels.
-- The 11 classification labels are seeded but the table is data-driven —
-- new angles emerge from real content analysis over time.
CREATE TABLE IF NOT EXISTS angle_library (
    id                    SERIAL PRIMARY KEY,
    outlier_reel_id       INT           REFERENCES outlier_reels(id),  -- NULL for seeded base angles
    angle_key             VARCHAR(100)  UNIQUE NOT NULL,  -- snake_case key e.g. contrarian_argument
    display_name          VARCHAR(150)  NOT NULL,
    psychological_mechanism TEXT        NOT NULL,
    avoid_when            TEXT,         -- conditions where this angle should NOT be used
    usage_count           INT           DEFAULT 0,
    avg_performance       NUMERIC(10,2) DEFAULT 0,
    created_at            TIMESTAMPTZ   DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_angle_library_avg_performance ON angle_library(avg_performance DESC);

-- Seed the 11 base angles from the Content Engine spec
INSERT INTO angle_library (angle_key, display_name, psychological_mechanism, avoid_when) VALUES
    ('contrarian_argument', 'Contrarian Argument',      'Challenges a widely held belief — creates cognitive dissonance that forces continued watching',       'When audience is already aligned with Nick — preaching to the choir reduces impact'),
    ('hard_truth',          'Hard Truth',               'Delivers an uncomfortable reality the viewer already knows but has not faced',                        'When following a vulnerability post — two uncomfortable videos back-to-back causes disengagement'),
    ('framework_reveal',    'Framework Reveal',         'Presents Nick''s specific system — satisfies the brain''s need for order and clarity',                'Avoid as a hook for cold audiences unfamiliar with Nick — leads with system before trust is built'),
    ('vulnerability_confession', 'Vulnerability and Confession', 'Nick shares a personal failure — breaks the guru illusion and spikes parasocial trust',    'Do not follow immediately with a sales-oriented video — trust spike should be used for nurturing'),
    ('case_study',          'Case Study',               'Tells a specific student success story — proof beats promises ten to one',                            'Do not use generic outcomes — must have specific named student and verifiable result'),
    ('myth_busting',        'Myth Busting',             'Dismantles a popular false belief — viewer feels smarter after watching',                             'Avoid myths that are already widely debunked — no novelty means no engagement spike'),
    ('prediction',          'Prediction',               'Makes a bold claim about what will happen — creates urgency and FOMO',                                'Avoid on declining trend topics — FOMO requires the topic to still be live'),
    ('direct_challenge',    'Direct Challenge',         'Confronts the viewer''s current inaction — activates ego and identity',                               'Avoid overuse — more than once a week creates fatigue and unfollows'),
    ('behind_the_scenes',   'Behind the Scenes',        'Takes the viewer inside Nick''s actual business — feels exclusive',                                   'Must show real process — staged BTS destroys credibility if discovered'),
    ('reaction_commentary', 'Reaction and Commentary',  'Nick reacts to a market event — piggybacks on existing attention',                                    'Only viable during Emerging or Peaking trend stages — Saturated or Declining kills reach'),
    ('specific_number_proof','Specific Number and Proof','Opens with a precise verifiable number from Nick''s track record',                                   'Number must be verified against Nick''s credentials database — never fabricate')
ON CONFLICT (angle_key) DO NOTHING;
