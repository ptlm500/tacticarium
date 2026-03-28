-- +goose Up
-- Users
CREATE TABLE IF NOT EXISTS users (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    discord_id       TEXT UNIQUE NOT NULL,
    discord_username TEXT NOT NULL,
    discord_avatar   TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Factions
CREATE TABLE IF NOT EXISTS factions (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    wahapedia_link TEXT
);

-- Detachments
CREATE TABLE IF NOT EXISTS detachments (
    id         TEXT PRIMARY KEY,
    faction_id TEXT NOT NULL REFERENCES factions(id),
    name       TEXT NOT NULL,
    UNIQUE(faction_id, name)
);

-- Stratagems
CREATE TABLE IF NOT EXISTS stratagems (
    id            TEXT PRIMARY KEY,
    faction_id    TEXT REFERENCES factions(id),
    detachment_id TEXT REFERENCES detachments(id),
    name          TEXT NOT NULL,
    type          TEXT NOT NULL,
    cp_cost       INT NOT NULL DEFAULT 1,
    legend        TEXT,
    turn          TEXT NOT NULL,
    phase         TEXT NOT NULL,
    description   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stratagems_faction ON stratagems(faction_id);
CREATE INDEX IF NOT EXISTS idx_stratagems_detachment ON stratagems(detachment_id);
CREATE INDEX IF NOT EXISTS idx_stratagems_phase ON stratagems(phase);

-- Mission packs
CREATE TABLE IF NOT EXISTS mission_packs (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT
);

-- Missions
CREATE TABLE IF NOT EXISTS missions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    description     TEXT,
    deployment_map  TEXT,
    rules_text      TEXT
);

-- Secondary objectives
CREATE TABLE IF NOT EXISTS secondaries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    category        TEXT NOT NULL,
    description     TEXT NOT NULL,
    max_vp          INT NOT NULL DEFAULT 8
);

-- Gambit cards
CREATE TABLE IF NOT EXISTS gambits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_pack_id TEXT NOT NULL REFERENCES mission_packs(id),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL,
    vp_value        INT NOT NULL DEFAULT 12
);
