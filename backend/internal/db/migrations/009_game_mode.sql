-- +goose Up
-- Tag stratagems and detachments with the game mode they belong to. The app
-- currently only supports the "core" mode; content belonging to alternate
-- modes (e.g. Boarding Actions) is filtered out of player-facing endpoints.
--
-- The authoritative list of Boarding Actions detachments lives in
-- Detachments.csv and is applied by the seed (see backend/internal/seed). This
-- migration only adds the column (defaulting to 'core') and the one backfill
-- that can be derived unambiguously from existing data: the 6 generic Boarding
-- Actions stratagems identified by their type prefix.
ALTER TABLE stratagems  ADD COLUMN game_mode TEXT NOT NULL DEFAULT 'core';
ALTER TABLE detachments ADD COLUMN game_mode TEXT NOT NULL DEFAULT 'core';

-- Generic Boarding Actions stratagems (no faction, no detachment). Their type
-- column uses the prefix "Boarding Actions – " (en-dash U+2013).
UPDATE stratagems
SET game_mode = 'boarding_actions'
WHERE type LIKE 'Boarding Actions – %';

CREATE INDEX idx_stratagems_game_mode  ON stratagems(game_mode);
CREATE INDEX idx_detachments_game_mode ON detachments(game_mode);

-- +goose Down
DROP INDEX IF EXISTS idx_stratagems_game_mode;
DROP INDEX IF EXISTS idx_detachments_game_mode;
ALTER TABLE stratagems  DROP COLUMN game_mode;
ALTER TABLE detachments DROP COLUMN game_mode;
