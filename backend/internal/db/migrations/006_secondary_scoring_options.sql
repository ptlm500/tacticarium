-- +goose Up
ALTER TABLE secondaries ADD COLUMN IF NOT EXISTS scoring_options JSONB NOT NULL DEFAULT '[]';
