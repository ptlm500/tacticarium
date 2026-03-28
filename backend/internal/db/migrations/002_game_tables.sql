-- +goose Up
-- Games
CREATE TABLE IF NOT EXISTS games (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invite_code       TEXT UNIQUE NOT NULL,
    status            TEXT NOT NULL DEFAULT 'setup',
    mission_pack_id   TEXT REFERENCES mission_packs(id),
    mission_id        UUID REFERENCES missions(id),
    current_round     INT NOT NULL DEFAULT 0,
    current_phase     TEXT NOT NULL DEFAULT 'setup',
    active_player     INT NOT NULL DEFAULT 1,
    first_turn_player INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    winner_id         UUID REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_games_invite_code ON games(invite_code);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);

-- Game players
CREATE TABLE IF NOT EXISTS game_players (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id        UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES users(id),
    player_number  INT NOT NULL CHECK (player_number IN (1, 2)),
    faction_id     TEXT REFERENCES factions(id),
    detachment_id  TEXT REFERENCES detachments(id),
    cp             INT NOT NULL DEFAULT 0,
    vp_primary     INT NOT NULL DEFAULT 0,
    vp_secondary   INT NOT NULL DEFAULT 0,
    vp_gambit      INT NOT NULL DEFAULT 0,
    vp_paint       INT NOT NULL DEFAULT 0,
    is_ready       BOOLEAN NOT NULL DEFAULT FALSE,
    gambit_id      UUID REFERENCES gambits(id),
    UNIQUE(game_id, player_number),
    UNIQUE(game_id, user_id)
);

-- Game player secondary objectives
CREATE TABLE IF NOT EXISTS game_player_secondaries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_player_id UUID NOT NULL REFERENCES game_players(id) ON DELETE CASCADE,
    secondary_id   UUID REFERENCES secondaries(id),
    custom_name    TEXT,
    custom_max_vp  INT,
    vp_scored      INT NOT NULL DEFAULT 0
);

-- Game events (audit log)
CREATE TABLE IF NOT EXISTS game_events (
    id            BIGSERIAL PRIMARY KEY,
    game_id       UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_number INT,
    event_type    TEXT NOT NULL,
    event_data    JSONB NOT NULL DEFAULT '{}',
    round         INT,
    phase         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_game_events_game ON game_events(game_id);

-- Stratagem usage tracking
CREATE TABLE IF NOT EXISTS stratagem_usage (
    id             BIGSERIAL PRIMARY KEY,
    game_id        UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    game_player_id UUID NOT NULL REFERENCES game_players(id) ON DELETE CASCADE,
    stratagem_id   TEXT NOT NULL REFERENCES stratagems(id),
    round          INT NOT NULL,
    phase          TEXT NOT NULL,
    cp_spent       INT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
