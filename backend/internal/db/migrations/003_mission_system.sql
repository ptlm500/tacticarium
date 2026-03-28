-- +goose Up
-- Mission rules / twists
CREATE TABLE IF NOT EXISTS mission_rules (
    id              TEXT PRIMARY KEY,
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    lore            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL
);

-- Challenger cards
CREATE TABLE IF NOT EXISTS challenger_cards (
    id              TEXT PRIMARY KEY,
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    lore            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL
);

-- Drop FK constraints that reference tables we need to recreate
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_mission_id_fkey;
ALTER TABLE game_players DROP CONSTRAINT IF EXISTS game_players_gambit_id_fkey;

-- Drop gambit column from game_players (no longer used)
ALTER TABLE game_players DROP COLUMN IF EXISTS gambit_id;

-- Recreate missions with TEXT PK and lore field
DROP TABLE IF EXISTS missions CASCADE;
CREATE TABLE missions (
    id              TEXT PRIMARY KEY,
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    lore            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL
);

-- Recreate secondaries with TEXT PK, lore, and is_fixed
DROP TABLE IF EXISTS game_player_secondaries CASCADE;
DROP TABLE IF EXISTS secondaries CASCADE;
CREATE TABLE secondaries (
    id              TEXT PRIMARY KEY,
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    lore            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL,
    max_vp          INT NOT NULL DEFAULT 5,
    is_fixed        BOOLEAN NOT NULL DEFAULT false
);

-- Recreate gambits with TEXT PK
DROP TABLE IF EXISTS gambits CASCADE;
CREATE TABLE gambits (
    id              TEXT PRIMARY KEY,
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL,
    vp_value        INT NOT NULL DEFAULT 30
);

-- Game: add twist tracking and change mission_id to TEXT
ALTER TABLE games ADD COLUMN IF NOT EXISTS twist_id TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS twist_name TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS mission_name TEXT;
ALTER TABLE games ALTER COLUMN mission_id TYPE TEXT USING mission_id::TEXT;

-- Game players: add mission system state
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS secondary_mode TEXT NOT NULL DEFAULT '';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS tactical_deck JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS active_secondaries JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS achieved_secondaries JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS discarded_secondaries JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS is_challenger BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS challenger_card_id TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS adapt_or_die_uses INT NOT NULL DEFAULT 0;
