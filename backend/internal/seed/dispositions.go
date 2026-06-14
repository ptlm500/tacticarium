package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type forceDispositionJSON struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Text        string      `json:"text"`
	GameVersion gameVersion `json:"game_version"`
}

// SeedForceDispositions upserts the five force dispositions.
func SeedForceDispositions(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var dispositions []forceDispositionJSON
	if err := readJSON(filePath, &dispositions); err != nil {
		return 0, err
	}

	count := 0
	for _, d := range dispositions {
		if d.ID == "" {
			continue
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO force_dispositions (id, name, text)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO UPDATE SET name = $2, text = $3`,
			d.ID, d.Name, d.Text)
		if err != nil {
			return count, fmt.Errorf("inserting force disposition %s: %w", d.ID, err)
		}
		count++
	}
	return count, nil
}

type missionMatchupJSON struct {
	ID                  string      `json:"id"`
	Disposition         string      `json:"disposition"`
	OpponentDisposition string      `json:"opponent_disposition"`
	MissionID           string      `json:"mission_id"`
	GameVersion         gameVersion `json:"game_version"`
}

// SeedMissionMatchups upserts the 5x5 disposition->mission selector matrix.
// Force dispositions and missions must be seeded first (both are FKs).
func SeedMissionMatchups(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var matchups []missionMatchupJSON
	if err := readJSON(filePath, &matchups); err != nil {
		return 0, err
	}

	count := 0
	for _, m := range matchups {
		if m.ID == "" {
			continue
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO mission_matchups (id, disposition, opponent_disposition, mission_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE
			   SET disposition = $2, opponent_disposition = $3, mission_id = $4`,
			m.ID, m.Disposition, m.OpponentDisposition, m.MissionID)
		if err != nil {
			return count, fmt.Errorf("inserting mission matchup %s: %w", m.ID, err)
		}
		count++
	}
	return count, nil
}
