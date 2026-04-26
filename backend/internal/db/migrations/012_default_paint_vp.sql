-- +goose Up
-- Army-painted is now a setup-time toggle that defaults to checked (10 VP).
-- Existing rows keep their stored vp_paint; only new players inherit the new
-- default.
ALTER TABLE game_players ALTER COLUMN vp_paint SET DEFAULT 10;

-- +goose Down
ALTER TABLE game_players ALTER COLUMN vp_paint SET DEFAULT 0;
