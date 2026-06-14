-- +goose Up
-- Reference data (factions, detachments) is wiped + reseeded from the 40kdc
-- snapshot as the sole authority. game_players keeps faction_id (and the legacy
-- detachment_id) as plain snapshot columns — the player's detachments now live
-- in the `detachments` JSONB — so the old FKs to the reference tables only get
-- in the way of a reseed (deleting a referenced row fails). Drop them; names
-- still resolve via LEFT JOIN once the reference rows are reseeded.

ALTER TABLE game_players DROP CONSTRAINT IF EXISTS game_players_faction_id_fkey;
ALTER TABLE game_players DROP CONSTRAINT IF EXISTS game_players_detachment_id_fkey;
