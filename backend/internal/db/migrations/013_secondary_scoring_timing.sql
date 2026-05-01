-- +goose Up
ALTER TABLE secondaries ADD COLUMN IF NOT EXISTS scoring_timing TEXT NOT NULL DEFAULT 'end_of_own_turn';
