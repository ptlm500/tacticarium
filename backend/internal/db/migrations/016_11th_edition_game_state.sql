-- +goose Up
-- 11th edition game-state contract migration.
--
-- The engine moves to the 11e model (asymmetric per-player missions, a board
-- with objective control, the draw-2-keep secondary deck, and evaluator-driven
-- scoring). Existing 10e game rows are incompatible and are wiped (the agreed
-- "wipe and reseed fresh" policy); only game state is affected — the 11e
-- reference data from migration 015 is untouched.

-- Wipe game data (cascades to game_players, game_events, stratagem_usage).
TRUNCATE TABLE games CASCADE;

-- games: drop 10e mission/twist columns (mission is now per-player); add board
-- + caps + start-of-turn control snapshot.
ALTER TABLE games DROP COLUMN IF EXISTS mission_pack_id;
ALTER TABLE games DROP COLUMN IF EXISTS mission_id;
ALTER TABLE games DROP COLUMN IF EXISTS mission_name;
ALTER TABLE games DROP COLUMN IF EXISTS twist_id;
ALTER TABLE games DROP COLUMN IF EXISTS twist_name;
ALTER TABLE games ADD COLUMN IF NOT EXISTS board                 JSONB NOT NULL DEFAULT '{}';
ALTER TABLE games ADD COLUMN IF NOT EXISTS vp_per_game_cap       INT NOT NULL DEFAULT 45;
ALTER TABLE games ADD COLUMN IF NOT EXISTS vp_per_round_cap      INT NOT NULL DEFAULT 15;
ALTER TABLE games ADD COLUMN IF NOT EXISTS start_of_turn_control JSONB NOT NULL DEFAULT '{}';

-- game_players: drop 10e gambit/challenger/adapt + old secondary state.
ALTER TABLE game_players DROP COLUMN IF EXISTS vp_gambit;
ALTER TABLE game_players DROP COLUMN IF EXISTS is_challenger;
ALTER TABLE game_players DROP COLUMN IF EXISTS challenger_card_id;
ALTER TABLE game_players DROP COLUMN IF EXISTS adapt_or_die_uses;
ALTER TABLE game_players DROP COLUMN IF EXISTS tactical_deck;
ALTER TABLE game_players DROP COLUMN IF EXISTS active_secondaries;
ALTER TABLE game_players DROP COLUMN IF EXISTS achieved_secondaries;
ALTER TABLE game_players DROP COLUMN IF EXISTS discarded_secondaries;
ALTER TABLE game_players DROP COLUMN IF EXISTS vp_primary_scored_slots;

-- game_players: 11e per-player state.
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS side                       TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS force_disposition          TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS force_disposition_name     TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS mission_id                 TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS mission_name               TEXT;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS primary_card               JSONB NOT NULL DEFAULT '{}';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS secondary_deck             JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS secondary_hand             JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS secondary_scored           JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS primary_scored_this_round  INT NOT NULL DEFAULT 0;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS secondary_scored_this_round INT NOT NULL DEFAULT 0;
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS pending_score_prompts      JSONB NOT NULL DEFAULT '[]';
