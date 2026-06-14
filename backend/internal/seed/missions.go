package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type missionJSON struct {
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	VPPerGameCap         int         `json:"vp_per_game_cap"`
	VPPerRoundCap        int         `json:"vp_per_round_cap"`
	DeploymentPatternIDs []string    `json:"deployment_pattern_ids"`
	GameVersion          gameVersion `json:"game_version"`
}

// SeedMissions upserts every mission (the objective record) from a 40kdc-data
// missions.json array.
func SeedMissions(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var missions []missionJSON
	if err := readJSON(filePath, &missions); err != nil {
		return 0, err
	}

	count := 0
	for _, m := range missions {
		if m.ID == "" {
			continue
		}
		patternIDs := m.DeploymentPatternIDs
		if patternIDs == nil {
			patternIDs = []string{}
		}
		gameCap := m.VPPerGameCap
		if gameCap == 0 {
			gameCap = 45
		}
		roundCap := m.VPPerRoundCap
		if roundCap == 0 {
			roundCap = 15
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO primary_missions (id, name, vp_per_game_cap, vp_per_round_cap, deployment_pattern_ids)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE
			   SET name = $2, vp_per_game_cap = $3, vp_per_round_cap = $4, deployment_pattern_ids = $5`,
			m.ID, m.Name, gameCap, roundCap, patternIDs)
		if err != nil {
			return count, fmt.Errorf("inserting mission %s: %w", m.ID, err)
		}
		count++
	}
	return count, nil
}
