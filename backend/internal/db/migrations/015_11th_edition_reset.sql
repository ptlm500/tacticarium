-- +goose Up
-- 11th Edition reference data — additive "expand" phase.
--
-- 10th edition is being dropped, but this migration is deliberately additive: it
-- introduces the 11e reference layer ALONGSIDE the existing 10e tables so the
-- current handlers and tests keep working while the rest of the edition upgrade
-- lands incrementally. The destructive "contract" steps — dropping the 10e
-- mission/scoring tables (gambits, challenger_cards, mission_rules, mission_packs,
-- the 10e missions/secondaries) and switching the handlers over — happen in the
-- later PRs that migrate those handlers, keeping every PR's build green.
--
-- Reseeding the 10e CSV data is gone, so the 10e tables simply go unpopulated
-- until they are removed.

-- 1. 11e attributes on the carried-forward reference tables. Added as nullable
--    / defaulted columns so existing 10e reads are unaffected.
ALTER TABLE factions    ADD COLUMN IF NOT EXISTS parent_faction_id  TEXT;
ALTER TABLE factions    ADD COLUMN IF NOT EXISTS faction_rule_id    TEXT;
ALTER TABLE detachments ADD COLUMN IF NOT EXISTS detachment_points  INT;
ALTER TABLE detachments ADD COLUMN IF NOT EXISTS force_dispositions TEXT[] NOT NULL DEFAULT '{}';

-- 2. 11e stratagems. The 10e `stratagems` table is retained (and still served by
--    the current handlers) until the faction handlers switch over; the 11e data
--    lands in a parallel table. The same stratagem id recurs across detachments,
--    sometimes with differing attributes (e.g. full-throttle is 1CP/charge for
--    Orks' Kult of Speed but 2CP/movement for Marines' Stormlance Task Force), so
--    the faithful identity is (detachment_id, id), captured by a surrogate key.
CREATE TABLE IF NOT EXISTS stratagems_11e (
    pk                  TEXT PRIMARY KEY,
    id                  TEXT NOT NULL,
    faction_id          TEXT REFERENCES factions(id),
    detachment_id       TEXT REFERENCES detachments(id),
    name                TEXT NOT NULL,
    type                TEXT,
    category            TEXT,
    cp_cost             INT NOT NULL DEFAULT 1,
    phases              TEXT[] NOT NULL DEFAULT '{}',
    player_turn         TEXT,
    timing              TEXT,
    target_restrictions JSONB,
    ability_id          TEXT
);
CREATE INDEX IF NOT EXISTS idx_stratagems_11e_faction    ON stratagems_11e(faction_id);
CREATE INDEX IF NOT EXISTS idx_stratagems_11e_detachment ON stratagems_11e(detachment_id);
CREATE INDEX IF NOT EXISTS idx_stratagems_11e_id         ON stratagems_11e(id);
CREATE INDEX IF NOT EXISTS idx_stratagems_11e_phases     ON stratagems_11e USING GIN (phases);

-- 3. New 11e reference layer (no collisions with 10e tables).

-- Force dispositions: the five strategic-intent tags.
CREATE TABLE IF NOT EXISTS force_dispositions (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    text  TEXT NOT NULL DEFAULT ''
);

-- Deployment patterns: objective coordinates + per-player territory/zone
-- polygons. The board model derives each objective's role from these.
CREATE TABLE IF NOT EXISTS deployment_patterns (
    id                             TEXT PRIMARY KEY,
    name                           TEXT NOT NULL,
    source                         TEXT,
    description                    TEXT,
    objectives                     JSONB NOT NULL DEFAULT '[]',
    territories                    JSONB NOT NULL DEFAULT '[]',
    zones                          JSONB NOT NULL DEFAULT '[]',
    recommended_terrain_layout_ids TEXT[] NOT NULL DEFAULT '{}'
);

-- Primary missions: the 11e objective record. VP caps are per-card (per game +
-- per round). Named distinctly from the 10e `missions` table, which it replaces
-- once the mission handlers switch over.
CREATE TABLE IF NOT EXISTS primary_missions (
    id                     TEXT PRIMARY KEY,
    name                   TEXT NOT NULL,
    vp_per_game_cap        INT NOT NULL DEFAULT 45,
    vp_per_round_cap       INT NOT NULL DEFAULT 15,
    deployment_pattern_ids TEXT[] NOT NULL DEFAULT '{}'
);

-- Mission matchups: the 5x5 selector matrix. (own disposition, opponent
-- disposition) -> the primary mission that player plays.
CREATE TABLE IF NOT EXISTS mission_matchups (
    id                   TEXT PRIMARY KEY,
    disposition          TEXT NOT NULL REFERENCES force_dispositions(id),
    opponent_disposition TEXT NOT NULL REFERENCES force_dispositions(id),
    mission_id           TEXT NOT NULL REFERENCES primary_missions(id),
    UNIQUE (disposition, opponent_disposition)
);

-- Cards: unified mission cards. Primary mission cards (card_type='primary')
-- carry the mission's scoring; secondary cards are the secondary deck. The
-- awards / when_drawn / actions blocks are the 40kdc-data DSL, interpreted by
-- the scoring evaluator.
CREATE TABLE IF NOT EXISTS cards (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    card_type   TEXT NOT NULL DEFAULT 'secondary' CHECK (card_type IN ('primary', 'secondary')),
    subtype     TEXT,
    when_drawn  JSONB,
    actions     JSONB NOT NULL DEFAULT '[]',
    awards      JSONB NOT NULL DEFAULT '[]',
    text        TEXT NOT NULL DEFAULT '',
    edition     TEXT NOT NULL DEFAULT '11th',
    dataslate   TEXT NOT NULL DEFAULT 'launch'
);

CREATE INDEX IF NOT EXISTS idx_cards_card_type ON cards(card_type);
CREATE INDEX IF NOT EXISTS idx_mission_matchups_disposition ON mission_matchups(disposition);
