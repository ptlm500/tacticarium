package seed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type deploymentPatternJSON struct {
	ID                          string          `json:"id"`
	Name                        string          `json:"name"`
	Source                      *string         `json:"source"`
	Description                 *string         `json:"description"`
	Objectives                  json.RawMessage `json:"objectives"`
	Territories                 json.RawMessage `json:"territories"`
	Zones                       json.RawMessage `json:"zones"`
	RecommendedTerrainLayoutIDs []string        `json:"recommended_terrain_layout_ids"`
	GameVersion                 gameVersion     `json:"game_version"`
}

// SeedDeploymentPatterns upserts every deployment pattern. The objective
// coordinates and territory/zone polygons feed the board model.
func SeedDeploymentPatterns(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var patterns []deploymentPatternJSON
	if err := readJSON(filePath, &patterns); err != nil {
		return 0, err
	}

	count := 0
	for _, p := range patterns {
		if p.ID == "" {
			continue
		}
		layoutIDs := p.RecommendedTerrainLayoutIDs
		if layoutIDs == nil {
			layoutIDs = []string{}
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO deployment_patterns
			   (id, name, source, description, objectives, territories, zones, recommended_terrain_layout_ids)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (id) DO UPDATE SET
			   name = $2, source = $3, description = $4, objectives = $5,
			   territories = $6, zones = $7, recommended_terrain_layout_ids = $8`,
			p.ID, p.Name, p.Source, p.Description,
			rawOrDefault(p.Objectives, "[]"), rawOrDefault(p.Territories, "[]"),
			rawOrDefault(p.Zones, "[]"), layoutIDs)
		if err != nil {
			return count, fmt.Errorf("inserting deployment pattern %s: %w", p.ID, err)
		}
		count++
	}
	return count, nil
}
