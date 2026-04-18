-- +goose Up
ALTER TABLE secondaries ADD COLUMN IF NOT EXISTS draw_restriction JSONB;
