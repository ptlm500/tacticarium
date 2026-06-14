-- +goose Up
-- Multi-detachment selection (11e lets a player combine detachments up to a
-- points budget). Per-player detachment moves from the single detachment_id
-- column to a JSONB list of {id, name, points}. The legacy detachment_id /
-- detachment_name columns are left in place (now unused, kept to avoid churn on
-- the shared reference FK); the engine state lives in `detachments`.
--
-- Also persist fixed_secondary_ids: a fixed-mode player's chosen set was held
-- only in memory, so it was lost if the room reloaded from the DB during setup.

ALTER TABLE game_players ADD COLUMN IF NOT EXISTS detachments         JSONB NOT NULL DEFAULT '[]';
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS fixed_secondary_ids JSONB NOT NULL DEFAULT '[]';
