-- +goose Up
ALTER TABLE game_players ADD COLUMN hidden_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE game_players DROP COLUMN hidden_at;
