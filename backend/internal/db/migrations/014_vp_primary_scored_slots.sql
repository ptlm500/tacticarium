-- +goose Up
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS vp_primary_scored_slots JSONB NOT NULL DEFAULT '{}';
