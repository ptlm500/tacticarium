-- +goose Up
-- 11th Edition reference-data reset.
--
-- 10th edition support is dropped. This migration wipes all game data and the
-- 10e mission/scoring reference tables, reshapes the carried-forward reference
-- tables (factions, detachments, stratagems) to the 40kdc-data shape, and
-- creates the new 11e reference layer (force dispositions, missions, the
-- disposition matchup matrix, unified mission cards, deployment patterns).
--
-- This migration is intentionally NOT reversible (Up-only): the 10e dataset is
-- gone and is reseeded fresh from the vendored 40kdc-data snapshot. Game-state
-- (per-player) schema changes land in a later migration alongside the engine
-- rework.

-- 1. Wipe all game data. games CASCADE clears game_players, game_events,
--    stratagem_usage and game_player_secondaries.
TRUNCATE TABLE games CASCADE;

-- 2. Drop FKs from games into reference tables we are about to replace.
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_mission_id_fkey;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_mission_pack_id_fkey;

-- 3. Drop the game_player_secondaries join table (10e shape). 11e secondary
--    state lives in the game_players.*_secondaries JSONB columns.
DROP TABLE IF EXISTS game_player_secondaries CASCADE;

-- 4. Drop 10e-only reference tables.
DROP TABLE IF EXISTS gambits CASCADE;
DROP TABLE IF EXISTS challenger_cards CASCADE;
DROP TABLE IF EXISTS mission_rules CASCADE;
DROP TABLE IF EXISTS secondaries CASCADE;
DROP TABLE IF EXISTS missions CASCADE;
DROP TABLE IF EXISTS mission_packs CASCADE;

-- 5. Drop 10e mission/twist columns from games. mission_id is kept as a plain
--    TEXT column (FK removed); per-player asymmetric missions + force
--    dispositions are added with the engine rework.
ALTER TABLE games DROP COLUMN IF EXISTS mission_pack_id;
ALTER TABLE games DROP COLUMN IF EXISTS twist_id;
ALTER TABLE games DROP COLUMN IF EXISTS twist_name;

-- 6. Reshape factions to the 40kdc-data shape.
TRUNCATE TABLE factions CASCADE;
ALTER TABLE factions DROP COLUMN IF EXISTS wahapedia_link;
ALTER TABLE factions ADD COLUMN IF NOT EXISTS parent_faction_id TEXT;
ALTER TABLE factions ADD COLUMN IF NOT EXISTS faction_rule_id   TEXT;

-- 7. Reshape detachments: add detachment points + force dispositions, drop the
--    10e game_mode tag.
ALTER TABLE detachments DROP COLUMN IF EXISTS game_mode;
ALTER TABLE detachments ADD COLUMN IF NOT EXISTS detachment_points INT;
ALTER TABLE detachments ADD COLUMN IF NOT EXISTS force_dispositions TEXT[] NOT NULL DEFAULT '{}';

-- 8. Rebuild stratagems in the 40kdc-data shape.
--    The same stratagem id recurs across detachments, sometimes with differing
--    attributes (e.g. full-throttle is 1CP/charge for Orks' Kult of Speed but
--    2CP/movement for Marines' Stormlance Task Force). The faithful identity is
--    therefore (detachment_id, id), captured by an explicit surrogate primary
--    key so a detachment's full stratagem list is preserved. The stratagem_usage
--    FK to stratagems(id) is dropped — game data is wiped and how usage
--    references a (detachment-scoped) stratagem is redefined with the engine work.
ALTER TABLE stratagem_usage DROP CONSTRAINT IF EXISTS stratagem_usage_stratagem_id_fkey;
DROP TABLE IF EXISTS stratagems CASCADE;
CREATE TABLE stratagems (
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
CREATE INDEX idx_stratagems_faction    ON stratagems(faction_id);
CREATE INDEX idx_stratagems_detachment ON stratagems(detachment_id);
CREATE INDEX idx_stratagems_id         ON stratagems(id);
CREATE INDEX idx_stratagems_phases     ON stratagems USING GIN (phases);

-- 9. New 11e reference layer.

-- Force dispositions: the five strategic-intent tags.
CREATE TABLE IF NOT EXISTS force_dispositions (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    text  TEXT NOT NULL DEFAULT ''
);

-- Deployment patterns: objective coordinates + per-player territory/zone
-- polygons. The board model derives each objective's role from these.
CREATE TABLE IF NOT EXISTS deployment_patterns (
    id                            TEXT PRIMARY KEY,
    name                          TEXT NOT NULL,
    source                        TEXT,
    description                   TEXT,
    objectives                    JSONB NOT NULL DEFAULT '[]',
    territories                   JSONB NOT NULL DEFAULT '[]',
    zones                         JSONB NOT NULL DEFAULT '[]',
    recommended_terrain_layout_ids TEXT[] NOT NULL DEFAULT '{}'
);

-- Missions: the objective record. VP caps are per-card (per game + per round).
CREATE TABLE IF NOT EXISTS missions (
    id                     TEXT PRIMARY KEY,
    name                   TEXT NOT NULL,
    vp_per_game_cap        INT NOT NULL DEFAULT 45,
    vp_per_round_cap       INT NOT NULL DEFAULT 15,
    deployment_pattern_ids TEXT[] NOT NULL DEFAULT '{}'
);

-- Mission matchups: the 5x5 selector matrix. (own disposition, opponent
-- disposition) -> the mission that player plays.
CREATE TABLE IF NOT EXISTS mission_matchups (
    id                   TEXT PRIMARY KEY,
    disposition          TEXT NOT NULL REFERENCES force_dispositions(id),
    opponent_disposition TEXT NOT NULL REFERENCES force_dispositions(id),
    mission_id           TEXT NOT NULL REFERENCES missions(id),
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
