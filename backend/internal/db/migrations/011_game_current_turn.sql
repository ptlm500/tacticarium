-- +goose Up
-- The games table tracked current_round and current_phase but not current_turn,
-- so reloading state from the DB always reset currentTurn to 0 (the Go zero
-- value), causing the frontend to display "Turn 0 of 2" for in-progress games.
ALTER TABLE games ADD COLUMN current_turn INT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE games DROP COLUMN current_turn;
