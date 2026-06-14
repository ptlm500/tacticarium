-- +goose Up
-- 11th edition contract migration: drop the 10e reference tables now that the
-- handlers read exclusively from the 11e layer (factions/detachments with the
-- new columns, stratagems_11e, force_dispositions, primary_missions,
-- mission_matchups, cards, deployment_patterns). These 10e tables have been
-- unpopulated since the data reset and are no longer referenced by any handler.

DROP TABLE IF EXISTS gambits CASCADE;
DROP TABLE IF EXISTS challenger_cards CASCADE;
DROP TABLE IF EXISTS mission_rules CASCADE;
DROP TABLE IF EXISTS secondaries CASCADE;
DROP TABLE IF EXISTS missions CASCADE;
DROP TABLE IF EXISTS mission_packs CASCADE;
DROP TABLE IF EXISTS stratagems CASCADE;
