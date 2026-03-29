-- +goose Up
ALTER TABLE missions ADD COLUMN IF NOT EXISTS scoring_timing TEXT NOT NULL DEFAULT 'end_of_command_phase';
